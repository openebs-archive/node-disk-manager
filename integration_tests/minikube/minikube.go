package minikube

import (
	"github.com/openebs/node-disk-manager/integration_tests/utils"
	"time"
)

const (
	// CommandName is the command to execute the minikube binary
	CommandName = "minikube"
	// Running is the active/running status of minikube cluster
	Running = "Running"
)

var runCommand = utils.RunCommandWithSudo
var execCommand = utils.ExecCommandWithSudo

// Minikube contains the objects to be used while interfacing
// with a minikube cluster
type Minikube struct {
	// the command to execute for minikube cluster
	Command string

	// timeout duration for all minikube operations
	Timeout time.Duration

	// wait time for minikube operation. If the command does not return
	// we wait continously till timeout is reached
	WaitTime time.Duration
}

// NewMinikube returns a new minikube struct with the command to execute
// and the default wait-timeout
func NewMinikube() Minikube {
	return Minikube{
		Command:  CommandName,
		Timeout:  time.Minute,
		WaitTime: time.Second,
	}
}
