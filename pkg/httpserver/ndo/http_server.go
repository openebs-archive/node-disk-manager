package ndo

import (
	"encoding/json"
	//"fmt"
	"github.com/openebs/node-disk-manager/pkg/common"
	"github.com/openebs/node-disk-manager/pkg/httpserver"
	"net/http"
)

func StartHttpServer() {

	http.HandleFunc("/liveness", livenessHandler)

	go func() {
		err := http.ListenAndServe(":8080", nil)

		if err != nil {
			panic(err)
		}
	}()

}

func livenessHandler(w http.ResponseWriter, r *http.Request) {

	var resp httpserver.Response

	json.NewDecoder(r.Body).Decode(&resp)
	common.UpdateNodeLivenessTimeStamp(resp.Hostname)
}
