package hot_compiler

import (
	"context"
	"strings"
	"testing"

	gdiff "github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sirupsen/logrus"
)

const expectedCodegen = `//+build controllerbus_hot_plugin

package main

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	boilerplate_controller "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	"github.com/aperturerobotics/controllerbus/hot/plugin"
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
type Plugin = hot_plugin.StaticPlugin
// NewPlugin constructs the static container plugin.
func NewPlugin() *Plugin {
	return hot_plugin.NewStaticPlugin(BinaryID, BinaryVersion, BinaryFactories)
}
// ControllerBusHotPlugin is the variable read by the plugin loader.
var ControllerBusHotPlugin hot_plugin.HotPlugin = NewPlugin()
// _ is a type assertion
var _ hot_plugin.HotPlugin = ((*Plugin)(nil))
`

func TestCodegen(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	packagePaths := []string{
		"github.com/aperturerobotics/controllerbus/example/boilerplate/controller",
	}
	an, err := AnalyzePackages(le, packagePaths)
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
