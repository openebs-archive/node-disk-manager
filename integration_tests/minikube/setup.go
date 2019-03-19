package minikube

import (
	"fmt"
	"strings"
)

// Return the status of the minikube cluster.
// The map contains status of each component.
// apiserver, kubelet etc.
func (minikube Minikube) Status() (map[string]string, error) {
	command := minikube.Command + " status"
	output, err := execCommand(command)
	if err != nil {
		return nil, fmt.Errorf("error getting status of minikube. Error: %v", err)
	}
	status := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		keyval := strings.SplitN(line, ":", 2)
		if len(keyval) == 1 {
			status[strings.TrimSpace(keyval[0])] = ""
		} else {
			status[strings.TrimSpace(keyval[0])] = strings.TrimSpace(keyval[1])
		}
	}
	return status, nil
}
