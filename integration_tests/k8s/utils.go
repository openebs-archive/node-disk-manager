package k8s

import (
	"time"
)

// WaitForStateChange sleeps the process for a fixed Duration
// so that the state changes get written in etcd and we get the
// updated result
func WaitForStateChange() {
	time.Sleep(k8sWaitTime)
}
