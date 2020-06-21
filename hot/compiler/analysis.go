package hot_compiler

import (
	"go/build"
	// "go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/loader"
)

// Analysis contains the result of code analysis.
type Analysis struct {
	// fset is the file set
	fset *token.FileSet
	// prog is the program
	prog *loader.Program
	// imports contains the set of packages to import
	// keyed by import path
	imports map[string]*types.Package
	// factories contains the set of factories to build
	factories map[string]struct{}
}

// AnalyzePackages analyzes code packages using Go module package resolution.
func AnalyzePackages(le *logrus.Entry, packagePaths []string) (*Analysis, error) {
	res := &Analysis{
		imports: map[string]*types.Package{
			// "context": nil,
			"github.com/aperturerobotics/controllerbus/bus":        nil,
			"github.com/aperturerobotics/controllerbus/controller": nil,
			"github.com/aperturerobotics/controllerbus/hot/plugin": nil,
		},
		factories: make(map[string]struct{}),
	}

	builderCtx := build.Default
	builderCtx.BuildTags = append(builderCtx.BuildTags, "controllerbus_hot_analyze")

	var conf loader.Config
	conf.Build = &builderCtx
	// conf.ParserMode |= parser.ParseComments | parser.AllErrors
	// conf.ParserMode |= parser.DeclarationErrors
	for _, pkgPath := range packagePaths {
		conf.Import(pkgPath)
	}
	conf.Cwd, _ = os.Getwd()

	prog, err := conf.Load()
	if err != nil {
		return nil, errors.Wrap(err, "unable to load go program")
	}
	res.fset = prog.Fset
	res.prog = prog

	initPkgList := prog.InitialPackages()
	le.Infof("loaded %d init packages to analyze", len(initPkgList))
	if len(initPkgList) == 0 {
		return nil, errors.New("expected at least one package to be loaded")
	}
	// initPkg := initPkgList[0]

	// Find NewFactory() constructors.
	// Build a list of packages to import.
	for _, pkg := range initPkgList {
		le := le.WithField("pkg", pkg.Pkg.Path())
		factoryCtorObj := pkg.Pkg.Scope().Lookup("NewFactory")
		if factoryCtorObj == nil {
			le.Warn("no factory constructors found")
			continue
		}
		le.Debugf("found factory ctor func: %s", factoryCtorObj.Type().String())
		factoryPkgImportPath := factoryCtorObj.Pkg().Path()
		if _, ok := res.imports[factoryPkgImportPath]; !ok {
			le.
				WithField("import-path", factoryPkgImportPath).
				WithField("import-name", pkg.Pkg.Name).
				Infof("added package to imports list: %s", factoryPkgImportPath)
			res.imports[factoryPkgImportPath] = pkg.Pkg
		}
		res.factories[BuildPackageName(pkg.Pkg)] = struct{}{}
	}

	// TODO
	return res, nil
}

// GetProgram returns the program.
func (a *Analysis) GetProgram() *loader.Program {
	return a.prog
}

// GetProgramCodeFiles returns file paths for packages in the program.
// exactMatchFilter and importPathPrefixFilter are optional.
func (an *Analysis) GetProgramCodeFiles(
	exactMatchFilter []string,
	importPathPrefixFilter string,
) map[string][]string {
	res := make(map[string][]string)
	watchFile := func(pakImportPath, filePath string) {
		res[pakImportPath] = append(res[pakImportPath], filePath)
	}

	// collect go files to watch
	for pakInfo, pak := range an.prog.AllPackages {
		for i := range pak.Files {
			pakImportPath := pakInfo.Path()
			if importPathPrefixFilter != "" &&
				!strings.HasSuffix(pakImportPath, importPathPrefixFilter) {
				continue
			}
			if len(exactMatchFilter) != 0 {
				var found bool
				for _, ex := range exactMatchFilter {
					if ex == pakImportPath {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			fsetFile := an.fset.File(pak.Files[i].Pos())
			watchFile(pakImportPath, fsetFile.Name())
		}
	}

	return res
}
