package main

import (
	"github.com/openebs/node-disk-manager/ndm-exporter/cmd"
	"github.com/openebs/node-disk-manager/pkg/logs"
)

func main() {
	// init logger
	logs.InitLogs()
	defer logs.FlushLogs()

	cmd.Execute()
}
