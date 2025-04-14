package cli

import (
	"encoding/json"
	"os"

	bus_api "github.com/aperturerobotics/controllerbus/bus/api"
	"github.com/aperturerobotics/cli"
)

// RunBusInfo runs the bus information command.
func (a *ClientArgs) RunBusInfo(_ *cli.Context) error {
	ctx := a.GetContext()
	c, err := a.BuildClient()
	if err != nil {
		return err
	}

	ni, err := c.GetBusInfo(ctx, &bus_api.GetBusInfoRequest{})
	if err != nil {
		return err
	}

	if a.Interactive {
		_, _ = os.Stdout.Write(ni.PrintPrettyStatus())
		return nil
	} else {
		dat, err := json.MarshalIndent(ni, "", "\t")
		if err != nil {
			return err
		}

		os.Stdout.WriteString(string(dat))
		os.Stdout.WriteString("\n")
	}
	return nil
}
