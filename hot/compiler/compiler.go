package hot_compiler

import (
	gast "go/ast"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

const buildTag = "controllerbus_hot_plugin"

// CompilePluginFromFile compiles the plugin from the gfile.
func CompilePluginFromFile(
	le *logrus.Entry,
	gfile *gast.File,
	intermediateGoFile string,
	outFile string,
) error {
	builderCtx := build.Default
	builderCtx.BuildTags = append(builderCtx.BuildTags, buildTag)
	dat, err := FormatFile(gfile)
	if err != nil {
		return err
	}
	// write the intermediate go file
	if err := ioutil.WriteFile(intermediateGoFile, dat, 0644); err != nil {
		return err
	}
	// start the go compiler
	ecmd := exec.Command(
		"go", "build",
		"-v",
		"-buildmode=plugin",
		"-tags",
		buildTag,
		"-o",
		outFile,
		intermediateGoFile,
	)
	ecmd.Env = make([]string, len(os.Environ()))
	copy(ecmd.Env, os.Environ())
	ecmd.Env = append(
		ecmd.Env,
		"GO111MODULE=on",
	)
	ecmd.Stderr = os.Stderr
	ecmd.Stdout = os.Stdout
	le.Infof("running go compiler: %s", ecmd.String())
	return ecmd.Run()
}
