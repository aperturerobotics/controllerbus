package hot_compiler

import (
	"context"
	"go/build"
	// "go/parser"
	"go/token"
	"go/types"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/packages"
)

// Analysis contains the result of code analysis.
type Analysis struct {
	// fset is the file set
	fset *token.FileSet
	// packages are the imported packages
	// keyed by package path
	packages map[string]*packages.Package
	// imports contains the set of packages to import
	// keyed by import path
	imports map[string]*types.Package
	// controllerFactories contains the set of packages containing controllers
	controllerFactories map[string]*packages.Package
	// module contains all factory modules
	module map[string]*packages.Module
}

// AnalyzePackages analyzes code packages using Go module package resolution.
func AnalyzePackages(
	ctx context.Context,
	le *logrus.Entry,
	workDir string,
	packagePaths []string,
) (*Analysis, error) {
	res := &Analysis{
		imports: map[string]*types.Package{
			// "context": nil,
			"github.com/aperturerobotics/controllerbus/bus":        nil,
			"github.com/aperturerobotics/controllerbus/controller": nil,
			"github.com/aperturerobotics/controllerbus/hot/plugin": nil,
		},
		controllerFactories: make(map[string]*packages.Package),
		packages:            make(map[string]*packages.Package),
		module:              make(map[string]*packages.Module),
	}

	builderCtx := build.Default
	builderCtx.BuildTags = append(builderCtx.BuildTags, "controllerbus_hot_analyze")

	var conf packages.Config
	conf.Context = ctx

	// NeedCompiledGoFiles adds CompiledGoFiles.
	// packages.NeedCompiledGoFiles |
	// NeedTypesSizes adds TypesSizes.
	// packages.NeedTypesSizes |

	conf.Fset = token.NewFileSet()
	conf.Mode = conf.Mode |
		// NeedName adds Name and PkgPath.
		packages.NeedName |
		// NeedFiles adds GoFiles and OtherFiles.
		packages.NeedFiles |
		// NeedImports adds Imports. If NeedDeps is not set, the Imports field will contain
		// "placeholder" Packages with only the ID set.
		packages.NeedImports |
		// NeedDeps adds the fields requested by the LoadMode in the packages in Imports.
		packages.NeedDeps |
		// NeedExportsFile adds ExportFile.
		packages.NeedExportsFile |
		// NeedTypes adds Types, Fset, and IllTyped.
		packages.NeedTypes |
		// NeedSyntax adds Syntax.
		packages.NeedSyntax |
		// NeedTypesInfo adds TypesInfo.
		packages.NeedTypesInfo |
		// NeedModule adds Module.
		packages.NeedModule

	conf.Dir = workDir
	conf.Logf = func(format string, args ...interface{}) {
		le.Debugf(format, args...)
	}

	loadedPackages, err := packages.Load(&conf, packagePaths...)
	if err != nil {
		return nil, err
	}
	res.fset = conf.Fset
	for _, pkg := range loadedPackages {
		res.packages[pkg.PkgPath] = pkg
	}

	le.Debugf("loaded %d init packages to analyze", len(loadedPackages))
	if len(loadedPackages) == 0 {
		return nil, errors.New("expected at least one package to be loaded")
	}
	// initPkg := loadedPackages[0]

	factoryModules := res.module

	// Find NewFactory() constructors.
	// Build a list of packages to import.
	for _, pkg := range loadedPackages {
		le := le.WithField("pkg", pkg.Types.Path())
		factoryCtorObj := pkg.Types.Scope().Lookup("NewFactory")
		factoryPkgImportPath := pkg.Types.Path()
		if factoryCtorObj != nil {
			le.Debugf("found factory ctor func: %s", factoryCtorObj.Type().String())
			res.controllerFactories[BuildPackageName(pkg.Types)] = pkg
			if _, ok := res.imports[factoryPkgImportPath]; !ok {
				le.
					WithField("import-path", factoryPkgImportPath).
					WithField("import-name", pkg.Types.Name).
					Debug("added package to plugin-file imports list")
				res.imports[factoryPkgImportPath] = pkg.Types
			}
		} else {
			le.Warn("no factory constructors found")
		}

		factoryMod := pkg.Module
		if _, ok := factoryModules[factoryMod.Path]; !ok {
			le.
				WithField("import-path", factoryPkgImportPath).
				WithField("module-path", factoryMod.Path).
				WithField("module-version", factoryMod.Version).
				Debug("added module to modules list")
			factoryModules[factoryMod.Path] = factoryMod
		}
	}

	return res, nil
}

// GetLoadedPackages returns the loaded packages.
func (a *Analysis) GetLoadedPackages() map[string]*packages.Package {
	return a.packages
}

// GetProgramCodeFiles returns file paths for packages in the program.
// exactMatchFilter and importPathPrefixFilter are optional.
func (a *Analysis) GetProgramCodeFiles(
	exactMatchFilter []string,
	importPathPrefixFilter string,
) map[string][]string {
	res := make(map[string][]string)
	watchFile := func(pakImportPath, filePath string) {
		res[pakImportPath] = append(res[pakImportPath], filePath)
	}

	// collect go files to watch
	for pakInfo, pak := range a.packages {
		_ = pakInfo
		for i := range pak.Syntax {
			pakImportPath := pak.PkgPath
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
			fsetFile := a.fset.File(pak.Syntax[i].Pos())
			watchFile(pakImportPath, fsetFile.Name())
		}
	}

	return res
}

// GetImportedModules returns the list of modules imported in the packages.
func (a *Analysis) GetImportedModules() map[string]*packages.Module {
	return a.module
}
