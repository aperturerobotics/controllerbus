package hot_compiler

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/mr-tron/base58/base58"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
	"mvdan.cc/gofumpt/format"
)

const codegenModulesPluginName = "plugin"

// ModuleCompiler assembles a series of Go module files on disk to orchestrate
// "go build" commands and produce a plugin with unique import paths for the
// changed packages.
type ModuleCompiler struct {
	ctx context.Context
	le  *logrus.Entry

	buildPrefix       string
	pluginCodegenPath string
	pluginBinaryID    string
}

// NewModuleCompiler constructs a new module compiler with paths.
//
// packagesLookupPath is the working directory for "go build."
func NewModuleCompiler(
	ctx context.Context,
	le *logrus.Entry,
	buildPrefix string,
	pluginCodegenPath string,
	pluginBinaryID string,
) (*ModuleCompiler, error) {
	if pluginCodegenPath == "" {
		return nil, errors.New("codegen path cannot be empty")
	}
	if buildPrefix == "" {
		return nil, errors.New("build prefix must be specified")
	}
	pluginCodegenPath, err := filepath.Abs(pluginCodegenPath)
	if err != nil {
		return nil, err
	}
	return &ModuleCompiler{
		ctx: ctx,
		le:  le,

		buildPrefix:       buildPrefix,
		pluginCodegenPath: pluginCodegenPath,
		pluginBinaryID:    pluginBinaryID,
	}, nil
}

// GenerateModules builds the modules files in the codegen path.
//
// buildPrefix should be something like cbus-hot-abcdef (no slash)
func (m *ModuleCompiler) GenerateModules(analysis *Analysis, pluginBinaryVersion string) error {
	buildPrefix := m.buildPrefix
	if _, err := os.Stat(m.pluginCodegenPath); err != nil {
		return err
	}

	goPackageContainerDirPath := build.Default.GOPATH
	if _, err := os.Stat(goPackageContainerDirPath); err != nil {
		return errors.Wrap(err, "check GOPATH exists")
	}
	goModCachePath, err := filepath.Abs(filepath.Join(goPackageContainerDirPath, "pkg"))
	if err != nil {
		return errors.Wrap(err, "determine go mod cache path")
	}
	// goModCachePathPattern := path.Join(goModCachePath, "*")

	// Create the base plugin dir.
	codegenModulesBaseDir := filepath.Join(m.pluginCodegenPath, buildPrefix)

	codegenModulesPluginPath := filepath.Join(codegenModulesBaseDir, codegenModulesPluginName)
	codegenModulesPluginPathBin := filepath.Join(codegenModulesPluginPath, "bin")
	if err := os.MkdirAll(codegenModulesPluginPathBin, 0755); err != nil {
		return err
	}

	// Create the output code plugin go.mod.
	outPluginModDir := codegenModulesPluginPath
	outPluginModFilePath := path.Join(outPluginModDir, "go.mod")
	outPluginCodeFilePath := path.Join(codegenModulesPluginPathBin, "plugin.go")

	// outPluginGoMod will contain the go.mod for the container plugin.
	// Add the first line "module plugin"
	outPluginGoMod := &modfile.File{}
	err = outPluginGoMod.AddModuleStmt(path.Join(buildPrefix, codegenModulesPluginName))
	if err != nil {
		return err
	}

	// For each module, create a codegen module directory.
	// Add a replace statement to outPluginGoMod for each.
	genCodegenModulePath := func(modPath string) string {
		return path.Join(codegenModulesBaseDir, modPath)
	}

	moduleCodegenPaths := make(map[string]string)
	loadedModules := analysis.GetImportedModules()
	for _, mod := range loadedModules {
		srcMod := mod
		for mod.Replace != nil {
			m.le.
				WithField("mod-curr-path", mod.Path).
				WithField("mod-next-path", mod.Replace.Path).
				Debug("module was replaced with another")
			mod = mod.Replace
		}

		// determine if we can do a github.com/... module structure
		// if this is ../../, then we need to use a hash instead.
		if os.PathSeparator != '/' {
			// this is sort of hacky but we expect to use this on linux.
			// of course, windows support is desirable, and a fix will need to be added here.
			return errors.New("can only work on systems where / is the path separator")
		}

		moduleOutpPath := path.Clean("/" + srcMod.Path)[1:]
		if moduleOutpPath == "" {
			shaSum := sha1.Sum([]byte(srcMod.GoMod))
			moduleOutpPath = "module-" + base58.Encode(shaSum[:])
		}

		modPathAbs := path.Dir(mod.GoMod)
		moduleImportPath := srcMod.Path
		m.le.
			WithField("module-import", moduleImportPath).
			WithField("module-path", modPathAbs).
			WithField("module-output-path", moduleOutpPath).
			WithField("build-prefix", buildPrefix).
			Debug("creating module in code-gen directory")
		codegenModDir := genCodegenModulePath(srcMod.Path)
		if _, err := os.Stat(codegenModDir); !os.IsNotExist(err) {
			// delete if it exists already
			err = os.RemoveAll(codegenModDir)
			if err != nil {
				return err
			}
		}
		if err := os.MkdirAll(codegenModDir, 0755); err != nil {
			return err
		}
		moduleCodegenPaths[srcMod.Path] = codegenModDir
		codegenModFile := path.Join(codegenModDir, "go.mod")

		// New function: relocateGoModFile - which loads, analyzes, and
		// transforms a go.mod file into a new AST structure which is based in a
		// different working directory with a different module name.

		// Create the go.mod by parsing the old one.
		outpModFile, err := parseGoModFile(mod.GoMod)
		if err != nil {
			return err
		}

		// Relocate the go.mod references to the new go.mod path.
		if err := relocateGoModFile(outpModFile, codegenModFile); err != nil {
			return err
		}

		// Add a reference to the old module path, if the old module path was
		// not within the Go module cache path.
		//
		// Note: HasPrefix is deprecated but OK for this use case.
		isThirdPartyModule := filepath.HasPrefix(modPathAbs, goModCachePath)
		if !isThirdPartyModule {
			// Add a replace to the absolute path of the containing repo.
			//
			// Ex: github.com/aperturerobotics/controllerbus =>
			// /home/myhome/repos/controllerbus/...
			err = outpModFile.AddReplace(srcMod.Path, "", modPathAbs, "")
			if err != nil {
				return err
			}

			// Also add the replacement to the final plugin go.mod.
			err = outPluginGoMod.AddReplace(srcMod.Path, "", modPathAbs, "")
			if err != nil {
				return err
			}
		} else {
			m.le.WithField("module-path", mod.Path).Debug("detected an out-of-tree module")
		}

		// The outpModFile was created by first copying the go.mod from the
		// containing module repository, so it contains all require and
		// replace statements as necessary. The output plugin container
		// module also needs to have the same blocks copied in order to
		// ensure the correct module dependency versions are resolved.
		//
		// Transform+copy the contents of the original module's go.mod file
		// into the target plugin go.mod file.
		xformSrcModFile, err := parseGoModFile(mod.GoMod)
		if err != nil {
			return err
		}
		relocateGoModFile(xformSrcModFile, outPluginModFilePath)
		for _, def := range xformSrcModFile.Replace {
			err = outPluginGoMod.AddReplace(
				def.Old.Path, def.Old.Version,
				def.New.Path, def.New.Version,
			)
			if err != nil {
				return err
			}
		}
		for _, def := range xformSrcModFile.Require {
			err = outPluginGoMod.AddRequire(def.Mod.Path, def.Mod.Version)
			if err != nil {
				return err
			}
		}
		outPluginGoMod.Cleanup()

		// For each peer module that will be code-gen, add a replace statement.
		// Note: replace statements /could/ be added on-demand, but more work.
		for _, peerMod := range loadedModules {
			if peerMod == mod || peerMod.Path == srcMod.Path {
				continue
			}
			peerModCodegenDir := genCodegenModulePath(peerMod.Path)
			peerModRelativePath, err := filepath.Rel(codegenModDir, peerModCodegenDir)
			if err != nil {
				return err
			}
			peerModRelativePath = ensureStartsWithDotSlash(peerModRelativePath)

			prefixPeerModPath := path.Join(buildPrefix, peerMod.Path)
			err = outpModFile.AddReplace(prefixPeerModPath, "", peerModRelativePath, "")
			if err != nil {
				return err
			}

			peerModRelativePathToPlugin, err := filepath.Rel(outPluginModDir, peerModCodegenDir)
			if err != nil {
				return err
			}
			peerModRelativePathToPlugin = ensureStartsWithDotSlash(peerModRelativePathToPlugin)
			err = outPluginGoMod.AddReplace(prefixPeerModPath, "", peerModRelativePathToPlugin, "")
			if err != nil {
				return err
			}
		}

		patchedModPath := path.Join(buildPrefix, srcMod.Path)
		_ = outpModFile.AddModuleStmt(patchedModPath)

		outpModFile.SortBlocks()
		outpModFile.Cleanup()
		destGoMod, err := outpModFile.Format()
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(
			codegenModFile,
			destGoMod,
			0644,
		); err != nil {
			return err
		}

	}

	rewritePackagesImports := func(pkgCodeFile *ast.File) {
		for _, pkgCodeImport := range pkgCodeFile.Imports {
			pkgCodeImportPath := pkgCodeImport.Path.Value
			if len(pkgCodeImportPath) < 2 {
				continue
			}
			pkgCodeImportPath = pkgCodeImportPath[1:]
			pkgCodeImportPath = pkgCodeImportPath[:len(pkgCodeImportPath)-1]
			targetPkg, ok := analysis.packages[pkgCodeImportPath]
			if !ok {
				continue
			}
			replacedTargetPath := path.Join(buildPrefix, targetPkg.Types.Path())
			pkgCodeImport.Path.Value = fmt.Sprintf("%q", replacedTargetPath)
		}
	}

	formatCodeFile := func(pkgCodeFile *ast.File) ([]byte, error) {
		format.File(analysis.fset, pkgCodeFile, format.Options{LangVersion: "1.14"})
		var outBytes bytes.Buffer
		var printerConf printer.Config
		printerConf.Mode |= printer.SourcePos
		err := printer.Fprint(&outBytes, analysis.fset, pkgCodeFile)
		return outBytes.Bytes(), err
	}

	// Copy the code files, adjusting the import paths for the new prefixed import paths.
	// Parse the source Go file, adjust the imports, format + write.
	for _, factoryPkg := range analysis.packages {
		if factoryPkg.IllTyped {
			var errs []string
			for _, er := range factoryPkg.Errors {
				errs = append(errs, er.Error())
			}
			return errors.Errorf(
				"package %s contains errors: %v",
				factoryPkg.PkgPath,
				errs,
			)
		}
		factoryPkgModPath := factoryPkg.Module.Path
		factoryPkgCodegenPath, ok := moduleCodegenPaths[factoryPkgModPath]
		if !ok {
			return errors.Errorf(
				"no codegen path was built for module %s for package %s",
				factoryPkgModPath,
				factoryPkg.Types.Path(),
			)
		}

		for _, pkgCodeFile := range factoryPkg.Syntax {
			pkgCodeFilePath := analysis.fset.File(pkgCodeFile.Pos()).Name()
			// pkgCodeFilename := path.Base(pkgCodeFilePath)

			// rewrite any imports if necessary.
			// any generated packages will have the new prefix before them.
			rewritePackagesImports(pkgCodeFile)

			// build relative path to the code file from the module root.
			codeFileRelativeToModule, err := filepath.Rel(factoryPkg.Module.Dir, pkgCodeFilePath)
			if err != nil {
				return err
			}

			// write the new formatted file to the output
			pkgCodeOutPath := filepath.Join(factoryPkgCodegenPath, codeFileRelativeToModule)
			pkgCodeOutDirPath := filepath.Dir(pkgCodeOutPath)
			m.le.
				WithField("orig-path", pkgCodeFilePath).
				WithField("target-path", pkgCodeOutPath).
				Debug("formatting code file")
			outData, err := formatCodeFile(pkgCodeFile)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(pkgCodeOutDirPath, 0755); err != nil {
				return err
			}
			err = ioutil.WriteFile(pkgCodeOutPath, outData, 0644)
			if err != nil {
				return err
			}
		}
	}

	pluginGoMod, err := outPluginGoMod.Format()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(outPluginModFilePath, pluginGoMod, 0644)
	if err != nil {
		return err
	}

	// Build the actual plugin file itself.
	gfile, err := CodegenPluginWrapperFromAnalysis(
		m.le,
		analysis,
		m.pluginBinaryID,
		pluginBinaryVersion,
	)
	if err != nil {
		return err
	}
	if buildPrefix != "" {
		gfile.Decls = append(gfile.Decls, &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{ast.NewIdent("HotPluginBuildPrefix")},
					Values: []ast.Expr{&ast.BasicLit{
						Kind:  token.STRING,
						Value: fmt.Sprintf("%q", buildPrefix),
					}},
				},
			},
		})
	}
	// Format to output pass #1
	pluginCodeData, err := formatCodeFile(gfile)
	if err != nil {
		return err
	}
	// we have to write it and then adjust paths, to populate fields in ast code.
	gfile, err = parser.ParseFile(
		analysis.fset,
		outPluginCodeFilePath,
		pluginCodeData,
		parser.ParseComments|parser.AllErrors,
	)
	if err != nil {
		return err
	}
	// Adjust the import paths.
	rewritePackagesImports(gfile)
	// Format to output pass #2
	pluginCodeData, err = formatCodeFile(gfile)
	if err != nil {
		return err
	}
	if buildPrefix != "" {
		pluginCodeData = append(pluginCodeData, []byte(
			"\nvar HotPluginBuildUUID = `"+buildPrefix+"`\n",
		)...)
	}
	if err := ioutil.WriteFile(outPluginCodeFilePath, pluginCodeData, 0644); err != nil {
		return err
	}

	return nil
}

// CompilePlugin compiles the plugin once.
// The module structure should have been built already.
func (m *ModuleCompiler) CompilePlugin(outFile string) error {
	le := m.le
	buildPrefix := m.buildPrefix
	codegenModulesBaseDir := m.pluginCodegenPath
	if buildPrefix != "" {
		codegenModulesBaseDir = filepath.Join(codegenModulesBaseDir, buildPrefix)
	}
	pluginDir := filepath.Join(codegenModulesBaseDir, codegenModulesPluginName, "bin")
	pluginDirAbs, err := filepath.Abs(pluginDir)
	if err != nil {
		return err
	}

	// build the intermediate output dir
	tmpName, err := ioutil.TempDir("", "controllerbus-hot-compiler-tmpdir")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpName)

	// start the go compiler execution
	ecmd := exec.Command(
		"go", "build",
		"-v", "-trimpath",
		"-buildmode=plugin",
		"-tags",
		buildTag,
		"-o",
		outFile,
		".",
	)
	ecmd.Dir = pluginDirAbs
	ecmd.Env = append(
		os.Environ(),
		"GO111MODULE=on",
	)
	ecmd.Stderr = os.Stderr
	ecmd.Stdout = os.Stdout
	le.
		WithField("work-dir", ecmd.Dir).
		Debugf("running go compiler: %s", ecmd.String())
	return ecmd.Run()
}

// Cleanup removes the codegen files, optionally with a build hash.
func (m *ModuleCompiler) Cleanup() {
	buildPrefix := m.buildPrefix
	codegenModulesBaseDir := m.pluginCodegenPath
	if codegenModulesBaseDir == "" {
		return
	}
	if buildPrefix != "" {
		codegenModulesBaseDir = filepath.Join(codegenModulesBaseDir, buildPrefix)
	}
	_ = os.RemoveAll(codegenModulesBaseDir)
}
