package main

import (
	"github.com/aperturerobotics/cli"
	ccli "github.com/aperturerobotics/controllerbus/cli"
)

// cliArgs are the client arguments
var cliArgs ccli.ClientArgs

func init() {
	clientCommands := (&cliArgs).BuildCommands()
	clientFlags := (&cliArgs).BuildFlags()
	commands = append(
		commands,
		&cli.Command{
			Name:        "client",
			Usage:       "client sub-commands",
			Subcommands: clientCommands,
			Flags:       clientFlags,
		},
	)
}
