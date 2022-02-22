package exec

import (
	"context"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// NewCmd builds a new exec cmd with defaults.
func NewCmd(proc string, args ...string) *exec.Cmd {
	ecmd := exec.Command(proc, args...)
	ecmd.Env = make([]string, len(os.Environ()))
	copy(ecmd.Env, os.Environ())
	ecmd.Stderr = os.Stderr
	ecmd.Stdout = os.Stdout
	return ecmd
}

// StartAndWait runs the given process and waits for ctx or process to complete.
func StartAndWait(ctx context.Context, le *logrus.Entry, ecmd *exec.Cmd) error {
	if ecmd.Process == nil {
		le.Debugf("exec: %s", ecmd.String())
		if err := ecmd.Start(); err != nil {
			return err
		}
	}

	outErr := make(chan error, 1)
	go func() {
		outErr <- ecmd.Wait()
	}()
	select {
	case <-ctx.Done():
		_ = ecmd.Process.Kill()
		<-outErr
		return ctx.Err()
	case err := <-outErr:
		if err != nil {
			le.WithError(err).Debug("process exited with error")
		} else {
			le.Debug("process exited")
		}
		return err
	}
}
