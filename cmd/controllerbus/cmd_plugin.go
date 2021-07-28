package main

import (
	plugin_cli "github.com/aperturerobotics/controllerbus/plugin/cli"
	"github.com/urfave/cli"
)

var pluginCompilerFlags plugin_cli.CompilerArgs

func init() {
	commands = append(
		commands,
		cli.Command{
			Name:        "plugin",
			Usage:       "plugin compiler utilities",
			Flags:       pluginCompilerFlags.BuildFlags(),
			Subcommands: pluginCompilerFlags.BuildSubCommands(),
		},
	)
}
