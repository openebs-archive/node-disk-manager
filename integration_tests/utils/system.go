package utils

import (
	"bytes"
	"fmt"
	"io"
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
	if err != nil {
		return fmt.Errorf("run failed %s %v", cmd, err)
	}
	return err
}

// Exec a command on the host and get the output
func ExecCommand(cmd string) (string, error) {
	substring := strings.Fields(cmd)
	name := substring[0]
	args := substring[1:]
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("exec failed %s %v", cmd, err)
	}
	return string(out), err
}

// Exec 2 commands, pipe the output of first command to second
func ExecCommandWithPipe(cmd1, cmd2 string) (string, error) {
	parts1 := strings.Fields(cmd1)
	parts2 := strings.Fields(cmd2)

	c1 := exec.Command(parts1[0], parts1[1:]...)
	c2 := exec.Command(parts2[0], parts2[1:]...)

	reader, writer := io.Pipe()
	c1.Stdout = writer
	c2.Stdin = reader

	var buffer bytes.Buffer
	c2.Stdout = &buffer

	err := c1.Start()
	if err != nil {
		return "", fmt.Errorf("error starting command: %q. Error: %v", cmd1, err)
	}
	err = c2.Start()
	if err != nil {
		return "", fmt.Errorf("error starting command: %q. Error: %v", cmd2, err)
	}
	err = c1.Wait()
	if err != nil {
		return "", fmt.Errorf("error while waiting for command: %q to exit. Error: %v", cmd1, err)
	}
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("error while closing the pipe writer. Error: %v", err)
	}
	err = c2.Wait()
	if err != nil {
		return "", fmt.Errorf("error while waiting for command: %q to exit. Error: %v", cmd2, err)
	}

	return buffer.String(), nil
}
