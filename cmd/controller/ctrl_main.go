/*
Copyright 2018 The OpenEBS Author

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
)

// Run will wait for SIGINT signal and exit if received
func Run(stopCh <-chan os.Signal) error {
	glog.Info("Started controller")

	<-stopCh

	glog.Info("Shutting down controller")

	return nil
}

func Watch() {
	// set up signals so we handle the first shutdown signal gracefully
	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGINT)

	if err := Run(sigCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}
