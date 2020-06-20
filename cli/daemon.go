package cli

import (
	"github.com/urfave/cli"
)

// DaemonArgs contains common flags for controller-bus daemons.
type DaemonArgs struct {
	WriteConfig bool
	ConfigPath  string
}

// BuildFlags attaches the flags to a flag set.
func (a *DaemonArgs) BuildFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "path to configuration yaml file",
			EnvVar:      "CONTROLLER_BUS_CONFIG",
			Value:       "controllerbus_daemon.yaml",
			Destination: &a.ConfigPath,
		},
		cli.BoolFlag{
			Name:        "write-config",
			Usage:       "write the daemon config file on startup",
			EnvVar:      "CONTROLLER_BUS_WRITE_CONFIG",
			Destination: &a.WriteConfig,
		},
	}
}
