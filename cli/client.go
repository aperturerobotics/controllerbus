package cli

import (
	"context"
	"errors"
	"net"

	bus_api "github.com/aperturerobotics/controllerbus/bus/api"
	"github.com/aperturerobotics/starpc/srpc"
	"github.com/urfave/cli/v2"
)

// ClientArgs contains the client arguments and functions.
type ClientArgs struct {
	// ctx is the context
	ctx context.Context
	// conn is the connection instance
	conn srpc.Client
	// client is the client instance
	client bus_api.SRPCControllerBusServiceClient

	// Interactive indicates we will pretty-print outputs.
	Interactive bool

	// DialAddr is the address to dial.
	DialAddr string

	// ExecConfigSetPath is the path to the exec controller request to execute.
	ExecConfigSetPath string
}

// BuildFlags attaches the flags to a flag set.
func (a *ClientArgs) BuildFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "dial-addr",
			Usage:       "address to dial API on",
			Destination: &a.DialAddr,
			Value:       "127.0.0.1:5110",
		},
	}
}

// BuildCommands attaches the commands.
func (a *ClientArgs) BuildCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "bus-info",
			Usage:  "returns bus information",
			Action: a.RunBusInfo,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:        "interactive",
					Usage:       "print interactive (pretty print) output",
					Destination: &a.Interactive,
					Value:       true,
					EnvVars:     []string{"CONTROLLER_BUS_INTERACTIVE"},
				},
			},
		},
		{
			Name:   "exec",
			Usage:  "execute a controller configset",
			Action: a.RunExecController,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "config-set-file",
					Aliases:     []string{"f"},
					Usage:       "path to config set json or yaml file",
					EnvVars:     []string{"CONTROLLER_BUS_EXEC_CONFIG_SET_FILE"},
					Destination: &a.ExecConfigSetPath,
				},
			},
		},
	}
}

// BuildControllerBusCommand returns the controller-bus sub-command set.
func (a *ClientArgs) BuildControllerBusCommand() *cli.Command {
	cbusCmds := a.BuildCommands()
	return &cli.Command{
		Name:        "controller-bus",
		Aliases:     []string{"cbus"},
		Usage:       "ControllerBus system sub-commands.",
		Subcommands: cbusCmds,
	}
}

// SetClient sets the client instance.
func (a *ClientArgs) SetClient(client bus_api.SRPCControllerBusServiceClient) {
	a.client = client
}

// BuildClient builds the client or returns it if it has been set.
func (a *ClientArgs) BuildClient() (bus_api.SRPCControllerBusServiceClient, error) {
	if a.client != nil {
		return a.client, nil
	}

	if a.DialAddr == "" {
		return nil, errors.New("dial address is not set")
	}

	nconn, err := net.Dial("tcp", a.DialAddr)
	if err != nil {
		return nil, err
	}
	muxedConn, err := srpc.NewMuxedConn(nconn, false, nil)
	if err != nil {
		return nil, err
	}
	a.conn = srpc.NewClientWithMuxedConn(muxedConn)
	a.client = bus_api.NewSRPCControllerBusServiceClient(a.conn)
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
