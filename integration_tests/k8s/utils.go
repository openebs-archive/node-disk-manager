package k8s

import "time"

func WaitForStateChange() {
	time.Sleep(k8sWaitTime)
}
