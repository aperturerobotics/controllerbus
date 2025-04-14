package main

import (
	ccli "github.com/aperturerobotics/controllerbus/cli"
	"github.com/aperturerobotics/cli"
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
