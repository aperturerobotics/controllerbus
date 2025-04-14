package main

import (
	"fmt"
	"os"

	"github.com/aperturerobotics/cli"
)

// Commands are the CLI commands
var commands []*cli.Command

func main() {
	app := cli.NewApp()
	app.Name = "controllerbus"
	app.HideVersion = true
	app.Usage = "command-line node and tools for controllerbus"
	app.Commands = commands

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
