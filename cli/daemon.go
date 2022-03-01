package cli

import (
	"github.com/urfave/cli"
)

// DaemonArgs contains common flags for controller-bus daemons.
type DaemonArgs struct {
	WriteConfig bool
	ConfigPath  string
	APIListen   string
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
		cli.StringFlag{
			Name:        "api-listen",
			Usage:       "if set, will listen on address for API connections, ex :5110",
			EnvVar:      "CONTROLLER_BUS_API_LISTEN",
			Value:       ":5110",
			Destination: &a.APIListen,
		},
	}
}
