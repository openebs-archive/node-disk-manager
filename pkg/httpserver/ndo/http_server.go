package ndo

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/httpserver"
	"github.com/openebs/node-disk-manager/pkg/liveness"
	"net/http"
	"time"
)

func StartHttpServer() {

	http.HandleFunc("/liveness", livenessHandler)

	go func() {
		glog.Info("NDO Start http server at port:8080")
		err := http.ListenAndServe(":8080", nil)

		if err != nil {
			panic(err)
		}
	}()

}

func livenessHandler(w http.ResponseWriter, r *http.Request) {

	var resp httpserver.Response

	json.NewDecoder(r.Body).Decode(&resp)
	glog.Info("Got Hostname:", resp.Hostname)
	liveness.UpdateNodeLivenessTimeStamp(resp.Hostname, time.Now())
}
