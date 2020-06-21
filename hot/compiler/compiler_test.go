package hot_compiler

import (
	"context"
	"strings"
	"testing"

	gdiff "github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sirupsen/logrus"
)

const expectedCodegen = `package main

import (
	"context"
	"github.com/aperturerobotics/controllerbus/bus"
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
// _ is a type assertion
var _ hot_plugin.HotPlugin = ((*Plugin)(nil))
`

func TestCodegen(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	genFile, err := GeneratePluginWrapper(
		ctx,
		le,
		"testingCodegen",
		"0.0.0",
		[]string{
			"github.com/aperturerobotics/controllerbus/example/boilerplate/controller",
		},
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
