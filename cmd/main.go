package main

import (
	"os"

	"github.com/openebs/node-disk-manager/cmd/app/command"
	ndmlogger "github.com/openebs/node-disk-manager/pkg/logs"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// Run node-disk-manager
func run() error {
	// Init logging
	ndmlogger.InitLogs()
	defer ndmlogger.FlushLogs()

	// Create & execute new command
	cmd, err := command.NewNodeDiskManager()
	if err != nil {
		return err
	}

	return cmd.Execute()
}
