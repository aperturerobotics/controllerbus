package main

import (
	hot_cli "github.com/aperturerobotics/controllerbus/hot/cli"
	"github.com/urfave/cli"
)

var hotCompilerFlags hot_cli.CompilerArgs

func init() {
	commands = append(
		commands,
		cli.Command{
			Name:        "hot",
			Usage:       "hot compiler utilities",
			Flags:       hotCompilerFlags.BuildFlags(),
			Subcommands: hotCompilerFlags.BuildSubCommands(),
		},
	)
}
