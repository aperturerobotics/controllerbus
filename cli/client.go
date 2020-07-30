package cli

import (
	"context"
	"errors"

	"github.com/aperturerobotics/controllerbus/bus/api"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

// ClientArgs contains the client arguments and functions.
type ClientArgs struct {
	// ctx is the context
	ctx context.Context
	// client is the client instance
	client bus_api.ControllerBusServiceClient

	// DialAddr is the address to dial.
	DialAddr string

	// ExecConfigSetPath is the path to the exec controller request to execute.
	ExecConfigSetPath string
}

// BuildFlags attaches the flags to a flag set.
func (a *ClientArgs) BuildFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "dial-addr",
			Usage:       "address to dial API on",
			Destination: &a.DialAddr,
			Value:       "127.0.0.1:5110",
		},
	}
}

// BuildCommands attaches the commands.
func (a *ClientArgs) BuildCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "bus-info",
			Usage:  "returns bus information",
			Action: a.RunBusInfo,
		},
		{
			Name:   "exec",
			Usage:  "execute a controller configset",
			Action: a.RunExecController,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "config-set-file, f",
					Usage:       "path to config set json or yaml file",
					EnvVar:      "CONTROLLER_BUS_EXEC_CONFIG_SET_FILE",
					Destination: &a.ExecConfigSetPath,
				},
			},
		},
	}
}

// SetClient sets the client instance.
func (a *ClientArgs) SetClient(client bus_api.ControllerBusServiceClient) {
	a.client = client
}

// BuildClient builds the client or returns it if it has been set.
func (a *ClientArgs) BuildClient() (bus_api.ControllerBusServiceClient, error) {
	if a.client != nil {
		return a.client, nil
	}

	if a.DialAddr == "" {
		return nil, errors.New("dial address is not set")
	}

	clientConn, err := grpc.Dial(a.DialAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	a.client = bus_api.NewControllerBusServiceClient(clientConn)
	return a.client, nil
}

// SetContext sets the context.
func (a *ClientArgs) SetContext(c context.Context) {
	a.ctx = c
}

// GetContext returns the context.
func (a *ClientArgs) GetContext() context.Context {
	if c := a.ctx; c != nil {
		return c
	}
	return context.TODO()
}
