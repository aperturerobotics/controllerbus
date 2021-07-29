package hot_compiler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	gast "go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/aperturerobotics/controllerbus/plugin"
	b58 "github.com/mr-tron/base58/base58"
	"github.com/sirupsen/logrus"
)

// GeneratePluginWrapper generates a wrapper package for a list of packages
// containing controller factories.
func GeneratePluginWrapper(
	ctx context.Context,
	le *logrus.Entry,
	an *Analysis,
	binaryName, binaryVersion string,
) (*gast.File, error) {
	// Build the plugin main package.
	return CodegenPluginWrapperFromAnalysis(le, an, binaryName, binaryVersion)
}

// FormatFile formats the output file.
func FormatFile(gf *gast.File) ([]byte, error) {
	var outDat bytes.Buffer
	outDat.WriteString("// +build " + buildTag + "\n\n")
	// fset := prog.Fset
	mergeImports(gf)
	fset := token.NewFileSet()
	if err := format.Node(&outDat, fset, gf); err != nil {
		return nil, err
	}
	return outDat.Bytes(), nil
}

// BuildPackageName builds the unique name for the package.
func BuildPackageName(pkg *types.Package) string {
	// for now just use package name
	return pkg.Name()
}

// CodegenPluginWrapperFromAnalysis codegens a plugin wrapper from analysis.
func CodegenPluginWrapperFromAnalysis(
	le *logrus.Entry,
	a *Analysis,
	binaryID, binaryVersion string,
) (*gast.File, error) {
	var allDecls []gast.Decl
	importStrs := make([]string, 0, len(a.imports))
	for impPkg := range a.imports {
		importStrs = append(importStrs, impPkg)
	}
	sort.Strings(importStrs)
	for _, impPath := range importStrs {
		impPkg := a.imports[impPath]
		// impPkg may be nil
		var impIdent *gast.Ident
		if impPkg != nil {
			impIdent = gast.NewIdent(BuildPackageName(impPkg))
		}
		allDecls = append(allDecls, &gast.GenDecl{
			Tok: token.IMPORT,
			Specs: []gast.Spec{
				&gast.ImportSpec{
					Name: impIdent,
					Path: &gast.BasicLit{
						Kind:  token.STRING,
						Value: `"` + impPath + `"`,
					},
				},
			},
		})
	}

	// BinaryID const
	allDecls = append(allDecls, &gast.GenDecl{
		Doc: &gast.CommentGroup{
			List: []*gast.Comment{{
				Text: "// BinaryID is the binary identifier.\n",
			}},
		},
		Tok: token.CONST,
		Specs: []gast.Spec{
			&gast.ValueSpec{
				Names:  []*gast.Ident{gast.NewIdent("BinaryID")},
				Values: []gast.Expr{&gast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", binaryID)}},
			},
		},
	})

	// BinaryVersion const
	allDecls = append(allDecls, &gast.GenDecl{
		Doc: &gast.CommentGroup{
			List: []*gast.Comment{{
				Text: "// BinaryVersion is the binary version string.\n",
			}},
		},
		Tok: token.CONST,
		Specs: []gast.Spec{
			&gast.ValueSpec{
				Names:  []*gast.Ident{gast.NewIdent("BinaryVersion")},
				Values: []gast.Expr{&gast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", binaryVersion)}},
			},
		},
	})

	// Construct the elements of the slice to return from BinaryFactories.
	var buildControllersElts []gast.Expr
	for fpkg := range a.controllerFactories {
		buildControllersElts = append(buildControllersElts, &gast.CallExpr{
			Args: []gast.Expr{
				gast.NewIdent("b"),
			},
			Fun: &gast.SelectorExpr{
				Sel: gast.NewIdent("NewFactory"),
				X:   gast.NewIdent(fpkg),
			},
		})
	}

	// BinaryFactories are the factories included in the binary.
	allDecls = append(allDecls, &gast.GenDecl{
		Doc: &gast.CommentGroup{
			List: []*gast.Comment{{
				Text: "// BinaryFactories are the factories included in the binary.\n",
			}},
		},
		Tok: token.VAR,
		Specs: []gast.Spec{
			&gast.ValueSpec{
				Names: []*gast.Ident{gast.NewIdent("BinaryFactories")},
				Values: []gast.Expr{&gast.FuncLit{
					Type: &gast.FuncType{
						Params: &gast.FieldList{
							List: []*gast.Field{{
								Names: []*gast.Ident{gast.NewIdent("b")},
								Type: &gast.SelectorExpr{
									X:   gast.NewIdent("bus"),
									Sel: gast.NewIdent("Bus"),
								},
							}},
						},
						Results: &gast.FieldList{List: []*gast.Field{
							{Type: &gast.ArrayType{Elt: &gast.SelectorExpr{
								X:   gast.NewIdent("controller"),
								Sel: gast.NewIdent("Factory"),
							}}},
						}},
					},
					Body: &gast.BlockStmt{List: []gast.Stmt{
						&gast.ReturnStmt{
							Results: []gast.Expr{
								&gast.CompositeLit{
									Elts: buildControllersElts,
									Type: &gast.ArrayType{Elt: &gast.SelectorExpr{
										X:   gast.NewIdent("controller"),
										Sel: gast.NewIdent("Factory"),
									}},
								},
							},
						},
					}},
				}},
			},
		},
	})

	// Plugin is the top-level static plugin container.
	// type Plugin = plugin.StaticPlugin
	allDecls = append(allDecls, &gast.GenDecl{
		Doc: &gast.CommentGroup{
			List: []*gast.Comment{{
				Text: "// Plugin is the top-level static plugin container.\n",
			}},
		},
		Tok: token.TYPE,
		Specs: []gast.Spec{
			&gast.TypeSpec{
				Name:   gast.NewIdent("Plugin"),
				Assign: 789,
				Type: &gast.SelectorExpr{
					X:   gast.NewIdent("plugin"),
					Sel: gast.NewIdent("StaticPlugin"),
				},
			},
		},
	})

	// NewPlugin constructs the static container plugin.
	allDecls = append(allDecls, &gast.FuncDecl{
		Doc: &gast.CommentGroup{
			List: []*gast.Comment{{
				Text: "// NewPlugin constructs the static container plugin.\n",
			}},
		},
		Name: gast.NewIdent("NewPlugin"),
		Body: &gast.BlockStmt{List: []gast.Stmt{
			&gast.ReturnStmt{
				Results: []gast.Expr{
					&gast.CallExpr{
						Fun: &gast.SelectorExpr{
							X:   gast.NewIdent("plugin"),
							Sel: gast.NewIdent("NewStaticPlugin"),
						},
						Args: []gast.Expr{
							gast.NewIdent("BinaryID"),
							gast.NewIdent("BinaryVersion"),
							gast.NewIdent("BinaryFactories"),
						},
					},
				},
			},
		}},
		Type: &gast.FuncType{
			Params: &gast.FieldList{List: []*gast.Field{}},
			Results: &gast.FieldList{
				List: []*gast.Field{
					{
						Type: &gast.StarExpr{
							X: gast.NewIdent("Plugin"),
						},
					},
				},
			},
		},
	})

	// ControllerBusPlugin const
	allDecls = append(allDecls, &gast.GenDecl{
		Doc: &gast.CommentGroup{
			List: []*gast.Comment{{
				Text: "// " + plugin.PluginGlobalVar + " is the variable read by the plugin loader.",
			}},
		},
		Tok: token.VAR,
		Specs: []gast.Spec{
			&gast.ValueSpec{
				Names: []*gast.Ident{gast.NewIdent(plugin.PluginGlobalVar)},
				Type: &gast.SelectorExpr{
					X:   gast.NewIdent("plugin"),
					Sel: gast.NewIdent("Plugin"),
				},
				Values: []gast.Expr{&gast.CallExpr{
					Fun: gast.NewIdent("NewPlugin"),
				}},
			},
		},
	})

	// _ is a type assertion for Plugin
	allDecls = append(allDecls, &gast.GenDecl{
		Doc: &gast.CommentGroup{
			List: []*gast.Comment{{
				Text: "// _ is a type assertion\n",
			}},
		},
		Tok: token.VAR,
		Specs: []gast.Spec{
			&gast.ValueSpec{
				Names: []*gast.Ident{gast.NewIdent("_")},
				Type: &gast.SelectorExpr{
					X:   gast.NewIdent("plugin"),
					Sel: gast.NewIdent("Plugin"),
				},
				Values: []gast.Expr{&gast.ParenExpr{X: &gast.CallExpr{
					Args: []gast.Expr{gast.NewIdent("nil")},
					Fun: &gast.ParenExpr{
						X: &gast.StarExpr{X: gast.NewIdent("Plugin")},
					},
				}}},
			},
		},
	})

	return &gast.File{
		Name:    gast.NewIdent("main"),
		Package: 5, // Force after build tag.
		Decls:   allDecls,
	}, nil
}

// BuildPlugin builds a plugin using a temporary code-gen path.
//
// Automates the end-to-end build process with reasonable defaults.
func BuildPlugin(ctx context.Context, le *logrus.Entry, packageSearchPath, outputPath string, packages []string) error {
	var err error
	packageSearchPath, err = filepath.Abs(packageSearchPath)
	if err != nil {
		return err
	}

	le.Infof("analyzing %d packages for plugin", len(packages))
	an, err := AnalyzePackages(ctx, le, packageSearchPath, packages)
	if err != nil {
		return err
	}

	// deterministic prefix gen
	var buildUid string
	{
		hs := sha256.New()
		for _, p := range packages {
			_, _ = hs.Write([]byte(p))
		}
		buildUid = b58.Encode(hs.Sum(nil))
	}

	// cannot use /tmp for this, need ~/.cache dir
	// c.CodegenDir, err = ioutil.TempDir("", "cbus-codegen")
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	codegenDir := filepath.Join(userCacheDir, "cbus-codegen-"+buildUid)
	le.Debugf("created tmpdir for code-gen process: %s", codegenDir)

	codegenDir, err = filepath.Abs(codegenDir)
	if err != nil {
		return err
	}

	// remove codegen dir on exit
	defer func() {
		_ = os.RemoveAll(codegenDir)
	}()
	if err := os.MkdirAll(codegenDir, 0755); err != nil {
		return err
	}

	buildPrefix := "cbus-plugin-" + (buildUid[:8])
	pluginBinaryID := buildPrefix
	le.
		WithField("build-prefix", buildPrefix).
		Infof("creating compiler for plugin with packages: %v", packages)
	mc, err := NewModuleCompiler(ctx, le, buildPrefix, codegenDir, pluginBinaryID)
	if err != nil {
		return err
	}

	pluginBinaryVersion := buildPrefix + "-{buildHash}"
	err = mc.GenerateModules(an, pluginBinaryVersion)
	if err != nil {
		return err
	}

	outputPath, err = filepath.Abs(outputPath)
	if err == nil {
		err = os.MkdirAll(path.Dir(outputPath), 0755)
	}
	if err != nil {
		return err
	}
	return mc.CompilePlugin(outputPath)
}
