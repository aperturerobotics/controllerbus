package cli

import (
	"errors"
	"strings"

	"github.com/urfave/cli"
)

// CompilerArgs contains common flags for the hot compiler.
type CompilerArgs struct {
	// CodegenDir is a directory to use for code-generation.
	// If empty, a temporary dir will be used.
	CodegenDir string
	// OutputPath is a path to the output file including {buildHash}.
	// Should have a .cbus.so suffix for controllerbus to recognize.
	// ex: demo.{buildHash}.cbus.so
	OutputPath string
	// PluginBinaryID is the plugin binary ID to set for the output plugin.
	PluginBinaryID string
	// PluginBinaryVersion is the plugin binary version to set for the output.
	// Can contain {buildHash}.
	// ex: dev-{buildHash}
	// If empty: uses {buildHash}.
	PluginBinaryVersion string

	// BuildPrefix sets the build prefix if compiling once
	BuildPrefix string
	// NoCleanup indicates we should not cleanup after we are done.
	NoCleanup bool
}

// BuildFlags attaches the flags to a flag set.
func (a *CompilerArgs) BuildFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "codegen-dir",
			Usage:       "path to directory to create/use for codegen, if empty uses tmpdir",
			EnvVar:      "CONTROLLER_BUS_CODEGEN_DIR",
			Value:       a.CodegenDir,
			Destination: &a.CodegenDir,
		},
		cli.StringFlag{
			Name:        "output, o",
			Usage:       "write the output plugin to `PATH` - accepts {buildHash}",
			EnvVar:      "CONTROLLER_BUS_OUTPUT",
			Value:       a.OutputPath,
			Destination: &a.OutputPath,
		},
		cli.StringFlag{
			Name:        "plugin-binary-id",
			Usage:       "binary id for the output plugin",
			EnvVar:      "CONTROLLER_BUS_PLUGIN_BINARY_ID",
			Value:       a.PluginBinaryID,
			Destination: &a.PluginBinaryID,
		},
		cli.StringFlag{
			Name:        "plugin-binary-version",
			Usage:       "binary version for the output plugin, accepts {buildHash}",
			EnvVar:      "CONTROLLER_BUS_PLUGIN_BINARY_VERSION",
			Value:       a.PluginBinaryVersion,
			Destination: &a.PluginBinaryVersion,
		},
		cli.BoolFlag{
			Name:        "no-cleanup",
			Usage:       "disable cleaning up the codegen dirs",
			EnvVar:      "CONTROLLER_BUS_NO_CLEANUP",
			Destination: &a.NoCleanup,
		},
	}
}

// BuildSubCommands builds the sub-command set.
func (a *CompilerArgs) BuildSubCommands() []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:   "compile",
			Usage:  "compile packages specified as arguments once",
			Action: a.runCompileOnce,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "build-prefix",
					Usage:       "build prefix to prepend to import paths, generated on default",
					EnvVar:      "CONTROLLER_BUS_PLUGIN_BUILD_PREFIX",
					Value:       a.BuildPrefix,
					Destination: &a.BuildPrefix,
				},
			},
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "codegen",
					Usage:  "generate code for modules in codegen path and exit",
					Action: a.runCodegenOnce,
				},
			},
		},
	}
}

// Validate validates the arguments.
func (a *CompilerArgs) Validate() error {
	if a.OutputPath == "" || !strings.HasSuffix(a.OutputPath, ".cbus.so") {
		return errors.New("output path must be set and end with .cbus.so")
	}
	// more?
	return nil
}
