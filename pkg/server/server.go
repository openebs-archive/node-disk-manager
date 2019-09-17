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
	"net/http"

	"github.com/golang/glog"
)

type Server struct {
	ListenPort  string
	MetricsPath string
	Handler     http.Handler
}

// Start boots up the server that runs on the specified port.
// Returns an error if there is no connection established.
func (s *Server) Start() error {
	http.Handle(s.MetricsPath, s.Handler)
	glog.Info("Starting HTTP server at http://localhost" + s.ListenPort + s.MetricsPath)
	err := http.ListenAndServe(s.ListenPort, nil)
	if err != nil {
		glog.Error("error starting http server :", err)
		return err
	}
	return nil
}
