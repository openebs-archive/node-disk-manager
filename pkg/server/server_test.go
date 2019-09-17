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

package server

import (
	"net"
	"net/http"
	"testing"
)

func TestStartHttpServer(t *testing.T) {

	s := Server{
		ListenPort:  ":9090",
		MetricsPath: "/metrics",
		Handler:     http.HandlerFunc(index),
	}
	ErrorMessages := make(chan error)
	go func() {
		//Block port 9090 and attempt to start http server at 9090.
		if p1, err := net.Listen("tcp", "localhost:9090"); err == nil {
			defer p1.Close()
		}
		ErrorMessages <- s.Start()
	}()
	msg := <-ErrorMessages
	if msg != nil {
		t.Log("Trying to start http server in a port which is busy.")
		t.Log(msg)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	// sample handler created for testing
}
