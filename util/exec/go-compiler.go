package exec

import (
	"os/exec"
)

// ExecGoCompiler runs the Go compiler with a given command.
func ExecGoCompiler(args ...string) *exec.Cmd {
	ecmd := NewCmd("go", args...)
	ecmd.Env = append(
		ecmd.Env,
		"GO111MODULE=on",
	)
	return ecmd
}

// ExecGoTidyModules runs the Go compiler to tidy go.mod and go.sum
func ExecGoTidyModules(extraArgs ...string) *exec.Cmd {
	return ExecGoCompiler("mod", "tidy")
}
