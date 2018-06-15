package common

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Debug governs whether to print verbose logs or not
// It can be set by Environment Variable `CITF_VERBOSE_LOG``
var Debug bool

func init() {
	debug := os.Getenv("CITF_VERBOSE_LOG")

	if strings.ToLower(debug) == "true" {
		Debug = true
	} else {
		Debug = false
	}
}

// ExecCommand executes the command supplied and return the output as well as error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommand(cmd string) (string, error) {
	if Debug {
		fmt.Printf("Executing command: %q\n", cmd)
	}
	// splitting head => g++ parts => rest of the command
	// python equivalent: parts = [x.strip() for x in cmd.split() if x.strip()]
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]

	out, err := exec.Command(head, parts...).Output()
	return string(out), err
}

// ExecCommandSync executes the command supplied and return the output as well as error
// It also takes one sync.WaitGroup object as an argument which it notifies once command is executed
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandSync(cmd string, wg *sync.WaitGroup) (string, error) {
	out, err := ExecCommand(cmd)
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return string(out), err
}

// ExecCommandWithSudo executes the command supplied with `sudo` and return the output as well as error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandWithSudo(cmd string) (string, error) {
	return ExecCommand("sudo " + cmd)
}

// ExecCommandWithSudoSync executes the command supplied with `sudo` and return the output as well as error
// It also takes one sync.WaitGroup object as an argument which it notifies once command is executed
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandWithSudoSync(cmd string, wg *sync.WaitGroup) (string, error) {
	out, err := ExecCommandWithSudo(cmd)
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return string(out), err
}

// execCommandArrayWithGivenStdin is the internal function which does the core job of
// all ExecCommand*WithGivenStdin* functions
func execCommandArrayWithGivenStdin(head string, parts []string, stdin string) (string, error) {
	command := exec.Command(head, parts...)

	command.Stdin = bytes.NewBuffer([]byte(stdin))

	out, err := command.Output()
	return string(out), err
}

// ExecCommandArrayWithGivenStdin executes the command supplied
// then feed the supplied stdin to commands stdin then return the output as well as error
// don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandArrayWithGivenStdin(cmd []string, stdin string) (string, error) {
	return execCommandArrayWithGivenStdin(cmd[0], cmd[1:], stdin)
}

// ExecCommandWithGivenStdin executes the command supplied
// then feed the supplied stdin to commands stdin then return the output as well as error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandWithGivenStdin(cmd, stdin string) (string, error) {
	return ExecCommandArrayWithGivenStdin(strings.Fields(cmd), stdin)
}

// ExecCommandArrayWithGivenStdinWithSudo executes the command supplied with `sudo`
// then feed the supplied stdin to commands stdin then return the output as well as error
// don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandArrayWithGivenStdinWithSudo(cmd []string, stdin string) (string, error) {
	return execCommandArrayWithGivenStdin("sudo", cmd, stdin)
}

// ExecCommandWithGivenStdinWithSudo executes the command supplied with `sudo`
// then feed the supplied stdin to commands stdin then return the output as well as error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandWithGivenStdinWithSudo(cmd, stdin string) (string, error) {
	return ExecCommandArrayWithGivenStdinWithSudo(strings.Fields(cmd), stdin)
}

// ExecPipeTwoCommandsArray takes two commands in its parameter. It runs first command
// and feed its output to second command as input
func ExecPipeTwoCommandsArray(cmd1, cmd2 []string) (string, error) {
	if Debug {
		fmt.Printf("Executing command: %q\n", strings.Join(cmd1, " ")+" | "+strings.Join(cmd2, " "))
	}
	c1 := exec.Command(cmd1[0], cmd1[1:]...)
	c2 := exec.Command(cmd2[0], cmd2[1:]...)

	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	// var err error
	// c2.Stdin, err = c1.StdoutPipe()
	// if err != nil {
	// 	return "", fmt.Errorf("Error getting stdout pipe of command: %q. Error: %+v", cmd1, err)
	// }

	var b2 bytes.Buffer
	c2.Stdout = &b2

	err := c1.Start()
	if err != nil {
		return "", fmt.Errorf("Error starting command: %q. Error: %+v", cmd1, err)
	}
	err = c2.Start()
	if err != nil {
		return "", fmt.Errorf("Error starting command: %q. Error: %+v", cmd2, err)
	}
	err = c1.Wait()
	if err != nil {
		return "", fmt.Errorf("Error while waiting for command: %q to exit. Error: %+v", cmd1, err)
	}
	err = w.Close()
	if err != nil {
		return "", fmt.Errorf("Error while closing the pipe writer. Error: %+v", err)
	}
	err = c2.Wait()
	if err != nil {
		return "", fmt.Errorf("Error while waiting for command: %q to exit. Error: %+v", cmd2, err)
	}

	return b2.String(), nil
}

// ExecPipeTwoCommands takes two commands in its parameter. It runs first command
// and feed its output to second command as input
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecPipeTwoCommands(cmd1, cmd2 string) (string, error) {
	// splitting head => g++ parts => rest of the command
	// python equivalent: parts = [x.strip() for x in cmd.split() if x.strip()]
	parts1 := strings.Fields(cmd1)
	parts2 := strings.Fields(cmd2)

	return ExecPipeTwoCommandsArray(parts1, parts2)
}

// RunCommand executes the command supplied and return the error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommand(cmd string) error {
	if Debug {
		fmt.Printf("Executing command: %q\n", cmd)
	}
	// splitting head => g++ parts => rest of the command
	// python equivalent: parts = [x.strip() for x in cmd.split() if x.strip()]
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]

	err := exec.Command(head, parts...).Run()
	return err
}

// RunCommandWithSudo executes the command supplied and return the error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommandWithSudo(cmd string) error {
	return RunCommand("sudo " + cmd)
}

// RunCommandSync executes the command supplied and return the error
// It also takes one sync.WaitGroup object as an argument which it notifies once command is executed
func RunCommandSync(cmd string, wg *sync.WaitGroup) (string, error) {
	err := RunCommand(cmd)
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return "", err
}

// runCommandArrayWithGivenStdin is the internal function which does the core job of
// all runCommand*WithGivenStdin* functions
func runCommandArrayWithGivenStdin(head string, parts []string, stdin string) error {
	command := exec.Command(head, parts...)

	command.Stdin = bytes.NewBuffer([]byte(stdin))

	err := command.Run()
	return err
}

// RunCommandArrayWithGivenStdin runs the command supplied
// then feed the supplied stdin to commands stdin then return the error
// don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommandArrayWithGivenStdin(cmd []string, stdin string) error {
	return runCommandArrayWithGivenStdin(cmd[0], cmd[1:], stdin)
}

// RunCommandWithGivenStdin runs the command supplied
// then feed the supplied stdin to commands stdin then return error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommandWithGivenStdin(cmd, stdin string) error {
	return RunCommandArrayWithGivenStdin(strings.Fields(cmd), stdin)
}

// RunCommandArrayWithGivenStdinWithSudo runs the command supplied with `sudo`
// then feed the supplied stdin to commands stdin then return the error
// don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommandArrayWithGivenStdinWithSudo(cmd []string, stdin string) error {
	return runCommandArrayWithGivenStdin("sudo", cmd, stdin)
}

// RunCommandWithGivenStdinWithSudo runs the command supplied with `sudo`
// then feed the supplied stdin to commands stdin then return the error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommandWithGivenStdinWithSudo(cmd, stdin string) error {
	return RunCommandArrayWithGivenStdinWithSudo(strings.Fields(cmd), stdin)
}

// GetenvFallback search for the key in environment, if it is there then
// this function returns the value otherwise it returns the fallback value
func GetenvFallback(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
