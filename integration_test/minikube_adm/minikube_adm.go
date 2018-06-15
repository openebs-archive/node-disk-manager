package minikubeadm

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/golang/glog"
	. "github.com/openebs/node-disk-manager/integration_test/common"
)

var (
	useSudo     = true // Default value to use sudo
	execCommand = ExecCommandWithSudo
	runCommand  = RunCommandWithSudo
)

func init() {
	useSudoEnv := strings.ToLower(strings.TrimSpace(os.Getenv("USE_SUDO")))
	if useSudoEnv == "true" { // If it is mentioned in the environment variable to use sudo
		useSudo = true // use sudo then
	} else if useSudoEnv == "false" { // Else if it is mentioned in the environment variable not to use sudo
		useSudo = false // do not use sudo
	} // Else use default value mentioned above

	if !useSudo {
		execCommand = ExecCommand
		runCommand = RunCommand
	}
}

// waitForDotKubeDirToBeCreated waits for `.kube` to be created
func waitForDotKubeDirToBeCreated() {
	homeDir := os.Getenv("HOME")

	fmt.Println("Waiting for `.kube` to be created...")
	for {
		if _, err := os.Stat(path.Join(homeDir, ".kube")); err == nil {
			fmt.Println(path.Join(homeDir, ".kube") + " created.")
			break
		} else if _, err := os.Stat("/root/.kube"); err == nil {
			fmt.Println("/root/.kube created.")
			break
		}
		time.Sleep(time.Second)
	}
}

// waitForDotMinikubeDirToBeCreated waits for `.minikube` to be created
func waitForDotMinikubeDirToBeCreated() {
	homeDir := os.Getenv("HOME")

	fmt.Println("Waiting for `.minikube` to be created...")
	for {
		if _, err := os.Stat(path.Join(homeDir, ".minikube")); err == nil {
			fmt.Println(path.Join(homeDir, ".minikube") + " created.")
			break
		} else if _, err := os.Stat("/root/.minikube"); err == nil {
			fmt.Println("/root/.minikube created.")
			break
		}
		time.Sleep(time.Second)
	}
}

// runPostStartCommandsForMinikube runs the commands required when run minikube as --vm-driver=none
// Assumption: Environment variables `USER` and `HOME` is well defined.
func runPostStartCommandsForMinikubeNoneDriver() {
	userName := os.Getenv("USER")
	homeDir := os.Getenv("HOME")
	commands := []string{
		"mv /root/.kube " + homeDir + "/.kube",
		"chown -R " + userName + " " + homeDir + "/.kube",
		"chgrp -R " + userName + " " + homeDir + "/.kube",
		"mv /root/.minikube " + homeDir + "/.minikube",
		"chown -R " + userName + " " + homeDir + "/.minikube",
		"chgrp -R " + userName + " " + homeDir + "/.minikube",
	}

	for _, command := range commands {
		fmt.Printf("Running %q\n", command)
		output, err := execCommand(command)
		if err != nil {
			fmt.Printf("Running %q failed. Error: %+v\n", command, err)
		} else {
			fmt.Printf("Run %q successfully. Output: %s\n", command, output)
		}
	}
}

// StartMinikube method starts minikube with `--vm-driver=none` option.
func StartMinikube() {
	err := runCommand("minikube start --vm-driver=none")
	// We can also use following:
	// "minikube start --vm-driver=none --feature-gates=MountPropagation=true --cpus=1 --memory=1024 --v=3 --alsologtostderr"
	if err != nil {
		glog.Fatal(err)
	}

	envChangeMinikubeNoneUser := os.Getenv("CHANGE_MINIKUBE_NONE_USER")
	if Debug {
		fmt.Printf("Environ CHANGE_MINIKUBE_NONE_USER = %q\n", envChangeMinikubeNoneUser)
	}
	if envChangeMinikubeNoneUser == "true" {
		// Below commands shall automatically run in this case.
		if Debug {
			fmt.Println("Returning from setup.")
		}
		return
	}

	waitForDotKubeDirToBeCreated()

	waitForDotMinikubeDirToBeCreated()

	runPostStartCommandsForMinikubeNoneDriver()
}

// Setup checks if a teardown is required before minikube start
// if so it does that and then start the minikube.
// It does nothing when minikube is already running.
// it prints status too.
func Setup() {
	// Minikube Status timeout is 1 minute
	minikubeStatus, err := CheckStatusTillTimeout(time.Minute)

	if Debug {
		if err != nil {
			fmt.Printf("Error occured while checking minikube status. Error: %+v\n", err)
		} else {
			fmt.Printf("minikube status: %q\n", minikubeStatus)
		}
	}

	teardownRequired := false
	startRequired := false

	status, ok := minikubeStatus["minikube"]
	if !ok {
		fmt.Println("\"minikube\" not present in status. May be minikube is not accessible. Aborting...")
		os.Exit(1)
	}
	if status == "" { // This means cluster itself is not there
		fmt.Println("cluster is not up. will start the machine")
		startRequired = true // So, Start the minikube
	} else if status == "Stopped" { // Cluster is there but it is stopped
		fmt.Println("minikube cluster is present but not \"Running\", so will tearing down the machine then start again.")
		teardownRequired = true // We need to teardown it first
		startRequired = true    // Then also we need to start the machine
	} else if status != "Running" { // If cluster is there and machine is not in "Stopped" or "Running" state
		// Then there is a problem
		fmt.Printf("minikube is in unknown state. State: %q. Aborting...", status)
		os.Exit(1)
	} else { // Else minikube is Running so we need not do anything.
		fmt.Println("minikube is already Running.")
	}

	// If we figured out that a teardown is needed then do so
	if teardownRequired {
		err = Teardown()
		if err != nil {
			fmt.Printf("Error while deleting machine. Error: %+v\n", err)
		} else {
			fmt.Println("minikube deleted.")
		}
	}

	// If we figured out that a start is needed then do so
	if startRequired {
		StartMinikube()
	}
}

// CheckStatus checks minikube status and parse it to a map .
// :return:   map: minikube status parsed into dict.
//          error: if any error occurs, otherwise nil
// Note: error can come when machine is stopped too. But in this case status will be filled too
func CheckStatus() (map[string]string, error) {
	// Caller of this function should have proper rights to check minikube status
	command := "minikube status"
	statusStr, err := execCommand(command)

	status := map[string]string{}
	for _, line := range strings.Split(TrimWhitespaces(statusStr), "\n") {
		keyval := strings.SplitN(line, ":", 2)
		if len(keyval) == 1 {
			status[TrimWhitespaces(keyval[0])] = ""
		} else {
			status[TrimWhitespaces(keyval[0])] = TrimWhitespaces(keyval[1])
		}
	}
	return status, err
}

// CheckStatusTillTimeout checks the status and in case where it does not find "minikube:" in status string,
// it retries until timeout. Then it returns the last status as well as the last error.
func CheckStatusTillTimeout(timeout time.Duration) (map[string]string, error) {
	startTime := time.Now()
	var status map[string]string
	var err error
	for time.Since(startTime) < timeout {
		status, err = CheckStatus()
		if _, ok := status["minikube"]; !ok {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return status, err
}

// Teardown deletes minikube
func Teardown() error {
	// Caller of this function should have proper rights to delete minikube
	return runCommand("minikube delete")
}

// ClearContainers removes all the docker containers present on the machine
func ClearContainers() error {
	// CAUTION: This function call deletes all docker containers
	containersStr, err := execCommand("docker ps -aq")
	if err != nil {
		return err
	}
	if containersStr != "" {
		containers := strings.Fields(containersStr)
		for _, container := range containers {
			err = runCommand("docker rm -f " + container)
			if err != nil {
				fmt.Printf("Error occured while deleting docker container: %s. Error: %+v\n", container, err)
			} else {
				fmt.Printf("Deleted container: %s\n", container)
			}
		}
	}
	return nil
}
