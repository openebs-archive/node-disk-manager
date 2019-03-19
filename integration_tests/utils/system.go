package utils

import (
	"os/exec"
	"strings"
)

// Run a command with sudo permissions
func RunCommandWithSudo(cmd string) error {
	return RunCommand("sudo " + cmd)
}

// Exec a command with sudo permissions and return the output
// as a string
func ExecCommandWithSudo(cmd string) (string, error) {
	return ExecCommand("sudo " + cmd)
}

// Run a command on the host
func RunCommand(cmd string) error {
	substring := strings.Fields(cmd)
	name := substring[0]
	args := substring[1:]
	err := exec.Command(name, args...).Run()
	return err
}

// Exec a command on the host and get the output
func ExecCommand(cmd string) (string, error) {
	substring := strings.Fields(cmd)
	name := substring[0]
	args := substring[1:]
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}
