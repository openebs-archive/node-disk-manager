package controller

import (
	"github.com/openebs/node-disk-manager/pkg/controller/disk"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, disk.Add)
}
