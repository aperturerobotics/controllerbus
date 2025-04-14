package cli

import (
	"context"
	"io"
	"os"

	"github.com/aperturerobotics/cli"
	controller_exec "github.com/aperturerobotics/controllerbus/controller/exec"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

// RunExecController runs the execute configset command.
func (a *ClientArgs) RunExecController(_ *cli.Context) error {
	ctx := a.GetContext()

	req := &controller_exec.ExecControllerRequest{}
	if csPath := a.ExecConfigSetPath; csPath != "" {
		data, err := os.ReadFile(csPath)
		if err != nil {
			return err
		}
		jdat, err := yaml.YAMLToJSON(data)
		if err != nil {
			return errors.Wrap(err, "parse configset file")
		}
		req.ConfigSetYaml = string(jdat)
	}

	u, err := a.BuildClient()
	if err != nil {
		return err
	}

	execClient, err := u.ExecController(ctx, req)
	if err != nil {
		return err
	}

	for {
		resp, err := execClient.Recv()
		if err != nil {
			if err == io.EOF || err == context.Canceled {
				err = nil
			}
			return err
		}

		data, err := resp.MarshalJSON()
		if err != nil {
			return err
		}
		os.Stdout.Write(data)
		os.Stdout.WriteString("\n")
	}
}
