package hot_compiler

import (
	"context"
	"os"
	"strings"
	"testing"

	gdiff "github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sirupsen/logrus"
)

const expectedCodegen = `// +build controllerbus_plugin

package main

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	boilerplate_controller "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	"github.com/aperturerobotics/controllerbus/plugin"
)
// BinaryID is the binary identifier.
const BinaryID = "testingCodegen"
// BinaryVersion is the binary version string.
const BinaryVersion = "0.0.0"
// BinaryFactories are the factories included in the binary.
var BinaryFactories = func(b bus.Bus) []controller.Factory {
	return []controller.Factory{boilerplate_controller.NewFactory(b)}
}
// Plugin is the top-level static plugin container.
type Plugin = plugin.StaticPlugin
// NewPlugin constructs the static container plugin.
func NewPlugin() *Plugin {
	return plugin.NewStaticPlugin(BinaryID, BinaryVersion, BinaryFactories)
}
// ControllerBusPlugin is the variable read by the plugin loader.
var ControllerBusPlugin plugin.Plugin = NewPlugin()
// _ is a type assertion
var _ plugin.Plugin = ((*Plugin)(nil))
`

func TestCodegen(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	packagePaths := []string{
		"github.com/aperturerobotics/controllerbus/example/boilerplate/controller",
	}
	workDir, _ := os.Getwd()
	an, err := AnalyzePackages(ctx, le, workDir, packagePaths)
	if err != nil {
		t.Fatal(err.Error())
	}
	genFile, err := GeneratePluginWrapper(
		ctx,
		le,
		an,
		"testingCodegen",
		"0.0.0",
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	dat, err := FormatFile(genFile)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(string(dat))
	output := strings.TrimSpace(string(dat))
	expected := strings.TrimSpace(expectedCodegen)
	if output != expected {
		dmp := gdiff.New()
		diffs := dmp.DiffMain(output, expected, false)
		t.Fatal(dmp.DiffPrettyText(diffs))
	}
}
