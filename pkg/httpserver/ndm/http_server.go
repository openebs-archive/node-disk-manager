package ndm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/httpserver"
	//"k8s.io/klog/glog"
)

func StartHttpServer() {

	go func() {
		err := http.ListenAndServe(":8080", nil)

		if err != nil {
			panic(err)
		}
	}()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			svc_ip := os.Getenv("NDO_OPERATOR_SERVICE_IP")
			if len(svc_ip) == 0 {
				glog.Error("Serivce IP not found")
			}

			select {
			case <-ticker.C:
				resp, err := createResponse()
				if err != nil {
					glog.Error(err)
				}

				//_, err = http.Post(svc_ip+"/liveness", "application/json", bytes.NewBuffer(resp))
				_, err = http.Post("http://"+svc_ip+":8080/liveness", "application/json", bytes.NewBuffer(resp))
				if err != nil {
					glog.Error(err)
				}
			}
		}
	}()
}

func createResponse() ([]byte, error) {

	host, _ := os.Hostname()
	response := httpserver.Response{
		Hostname: host,
	}

	resp, err := json.Marshal(&response)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	return resp, nil

}
