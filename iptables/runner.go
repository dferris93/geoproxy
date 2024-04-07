package iptables

import (
	"fmt"
	"os/exec"
)

type Runner interface {
	RunCommand(command string, args ...string) (string, error)
}

type RealRunner struct{}

func (r *RealRunner) RunCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run command: %v output: %s", err, out)
	}
	return string(out), nil
}