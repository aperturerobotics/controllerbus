package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	plugin_compiler "github.com/aperturerobotics/controllerbus/plugin/compiler"
	plugin_shared "github.com/aperturerobotics/controllerbus/plugin/loader/shared-library/filesystem"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	codegenDirPathRel = "./codegen-module/"
	binaryID          = "controllerbus/examples/hot-demo/codegen-demo/1"
	binaryVersion     = "0.0.1"
)

// NOTE: must have same version of:
//   "github.com/aperturerobotics/controllerbus/hot/plugin",
var packagesList = []string{
	// usual demo boilerplate controller
	"github.com/aperturerobotics/controllerbus/example/boilerplate/controller",
	// example of a basic demo controller package with a non-trivial relative module reference.
	"./demo-controller",
	// example of a cross-module reference to a package that does not contain a
	// factory, but rather is a indirect dependency. "lifting" packages like
	// this is necessary for hot-loading packages which reference newer versions
	// of utility packages. this adds thee same plugin id prefix to the package
	// and will load a copy of the package unique to the plugin.
	"github.com/pkg/errors",
	// example of a un-referenced package which can still be copied + configured
	// properly in the output dir.
	"github.com/aperturerobotics/controllerbus/cmd/controllerbus",
}

const configSetYaml = `
boilerplate-demo-0:
  config:
    exampleField: testing
  id: controllerbus/example/boilerplate/1
  revision: 1
loader-demo:
  config:
    exitAfterDur: 3s
  id: controllerbus/example/hot-demo/demo-controller/1
  revision: 1
`

func main() {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	if err := run(ctx, le); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
}

func run(ctx context.Context, le *logrus.Entry) error {
	codegenDirPath, err := filepath.Abs(codegenDirPathRel)
	if err != nil {
		return err
	}
	// clear codegen path
	if err := os.RemoveAll(codegenDirPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	le.Info("compiling example package")
	if err := os.MkdirAll(codegenDirPath, 0755); err != nil {
		return err
	}

	packagesLookupPath, _ := os.Getwd()
	buildPrefix := "cbus-plugin-abcdef"
	moduleCompiler, err := plugin_compiler.NewModuleCompiler(
		ctx,
		le,
		buildPrefix,
		codegenDirPath,
		"hot-demo-module",
	)
	if err != nil {
		return err
	}
	analysis, err := plugin_compiler.AnalyzePackages(ctx, le, packagesLookupPath, packagesList)
	if err != nil {
		return err
	}
	pluginBinaryVersion := buildPrefix
	if err := moduleCompiler.GenerateModules(analysis, pluginBinaryVersion); err != nil {
		return err
	}

	outCliPath := filepath.Join(
		codegenDirPath,
		buildPrefix,
		"github.com/aperturerobotics/controllerbus/cmd/controllerbus",
	)
	outPluginsPath := filepath.Join(outCliPath, "plugins")
	if err := os.MkdirAll(outPluginsPath, 0755); err != nil {
		return err
	}
	outPath := filepath.Join(outPluginsPath, fmt.Sprintf("example.%s.cbus.so", buildPrefix))
	if err := moduleCompiler.CompilePlugin(outPath); err != nil {
		return err
	}

	relToTarget, err := filepath.Rel(packagesLookupPath, filepath.Join(outPluginsPath, ".."))
	if err != nil {
		return err
	}

	le.Infof("package compiled, starting: %s", relToTarget)

	// write the config file
	outConfigPath := filepath.Join(outCliPath, "config.yaml")
	if err := ioutil.WriteFile(outConfigPath, []byte(configSetYaml), 0755); err != nil {
		return err
	}

	ecmd := plugin_compiler.ExecGoTidyModules()
	ecmd.Dir = outCliPath
	le.Debugf("running go mod tidy: %s", ecmd.String())
	err = ecmd.Run()
	if err != nil {
		return err
	}

	le.Info("loading the compiled plugin")
	bus, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		return err
	}
	sr.AddFactory(plugin_shared.NewFactory(bus))

	le.Warn("You will most likely see an error, see: https://github.com/golang/go/issues/27751")
	_, _, loaderRef, err := loader.WaitExecControllerRunning(
		ctx,
		bus,
		resolver.NewLoadControllerWithConfig(&plugin_shared.Config{
			Dir:   outPluginsPath,
			Watch: true,
		}),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "listen on grpc api")
	}
	defer loaderRef.Release()

	// keep running until ctrl-c or timeout
	oc, ocCancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	select {
	case <-oc.Done():
	case <-time.After(time.Second * 3):
	}
	ocCancel()
	return nil
}
