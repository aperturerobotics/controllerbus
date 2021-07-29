package plugin_compiler

import (
	"os"
	"os/exec"
)

// ExecGoCompiler runs the Go compiler with a given command.
func ExecGoCompiler(args ...string) *exec.Cmd {
	ecmd := exec.Command("go", args...)
	ecmd.Dir, _ = os.Getwd()
	ecmd.Env = make([]string, len(os.Environ()))
	copy(ecmd.Env, os.Environ())
	ecmd.Env = append(
		ecmd.Env,
		"GO111MODULE=on",
	)
	ecmd.Stderr = os.Stderr
	ecmd.Stdout = os.Stdout
	return ecmd
}

// ExecGoTidyModules runs the Go compiler to tidy go.mod and go.sum
func ExecGoTidyModules(extraArgs ...string) *exec.Cmd {
	return ExecGoCompiler("mod", "tidy")
}
