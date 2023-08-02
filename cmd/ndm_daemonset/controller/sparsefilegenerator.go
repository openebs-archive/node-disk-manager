/*
Copyright 2018 OpenEBS Authors.

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
	"strings"

	"github.com/openebs/node-disk-manager/blockdevice"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/klog/v2"

	"fmt"
	"os"
	"path"
	"strconv"
)

/*
Sparse File help simulate disk objects that can be used for testing and
proto typing solutions built using node-disk-manager(NDM).  Sparse files
will be created if NDM is provided with the location where sparse files
should be located.

On Startup, if a sparse directory (EnvSparseFileDir) is specified as a
environment variable, a Sparse file with specified size (EnvSparseFileSize) will
be created and an associated BlockDevice CR will be added to Kubernetes. By default
only one sparse file will be created which can be changed by passing the desired
number of sparse files required via the environment variable EnvSparseFileCount.

On Shutdown, the status of the sparse file BlockDevice CR will be marked as Unknown.
*/

const (

	// EnvSparseFileDir - defines a sparse directory.
	// if it is specified as a environment variable,
	// a sparse file with specified size (EnvSparseFileSize) will
	// be created inside specified directory (EnvSparseFileDir)
	// and an associated BlockDevice CR will be added to Kubernetes.
	EnvSparseFileDir = "SPARSE_FILE_DIR"
	//EnvSparseFileSize define the size of created sparse file
	EnvSparseFileSize = "SPARSE_FILE_SIZE"
	//EnvSparseFileCount defines the number of sparse files to be created
	EnvSparseFileCount = "SPARSE_FILE_COUNT"

	//SparseFileName is a name of Sparse file
	SparseFileName = "ndm-sparse.img"
	//SparseFileDefaultSize defines the default sparse file default size
	SparseFileDefaultSize = int64(1073741824)
	//SparseFileMinSize defines the minimum size for sparse file
	SparseFileMinSize = int64(1073741824)
	//SparseFileDefaultCount defines the default sparse count files
	SparseFileDefaultCount = "1"

	//SparseBlockDevicePrefix defines the prefix for the sparse device
	SparseBlockDevicePrefix = "sparse-"
)

// GetSparseFileDir returns the full path to the sparse
// file directory on the node.
func GetSparseFileDir() string {

	sparseFileDir := os.Getenv(EnvSparseFileDir)

	if len(sparseFileDir) < 1 {
		return ""
	}

	info, err := os.Stat(sparseFileDir)
	if os.IsNotExist(err) || !info.Mode().IsDir() {
		klog.Info("Specified directory doesnt exist:  ", sparseFileDir)
		return ""
	}

	return sparseFileDir
}

// GetSparseFileCount returns the number of sparse files to be
//  created by NDM. Returns 0, if invalid count is specified.
func GetSparseFileCount() int {

	sparseFileCountStr := os.Getenv(EnvSparseFileCount)

	if len(sparseFileCountStr) < 1 {
		sparseFileCountStr = SparseFileDefaultCount
	}

	sparseFileCount, econv := strconv.Atoi(sparseFileCountStr)
	if econv != nil {
		klog.Info("Error converting sparse file count:  ", sparseFileCountStr)
		return 0
	}

	return sparseFileCount
}

// GetSparseFileSize returns the size of the sparse file to be
//  created by NDM. Returns 0, if invalid size is specified.
func GetSparseFileSize() int64 {

	sparseFileSizeStr := os.Getenv(EnvSparseFileSize)
	if len(sparseFileSizeStr) < 1 {
		klog.Info("No size was specified. Using default size: ", fmt.Sprint(SparseFileDefaultSize))
		return SparseFileDefaultSize
	}

	fileSize, econv := strconv.ParseFloat(sparseFileSizeStr, 64)
	if econv != nil {
		klog.Error("Error converting sparse file size:  ", econv)
		return 0
	}
	sparseFileSize := int64(fileSize)

	if sparseFileSize < SparseFileMinSize {
		klog.Info(fmt.Sprint(sparseFileSizeStr), " is less than minimum required. Setting the size to:  ", fmt.Sprint(SparseFileMinSize))
		return SparseFileMinSize
	}

	return sparseFileSize
}

// InitializeSparseFiles will check if the sparse file exist or have to be
// created and will update or create the associated BlockDevice CR accordingly
func (c *Controller) InitializeSparseFiles() {
	sparseFileDir := GetSparseFileDir()
	sparseFileSize := GetSparseFileSize()
	sparseFileCount := GetSparseFileCount()

	if len(sparseFileDir) < 1 || sparseFileSize < 1 || sparseFileCount < 1 {
		klog.Info("No sparse file path/size provided. Skip creating sparse files.")
		return
	}

	for i := 0; i < sparseFileCount; i++ {
		sparseFile := path.Join(sparseFileDir, fmt.Sprint(i)+"-"+SparseFileName)
		err := CheckAndCreateSparseFile(sparseFile, sparseFileSize)
		if err != nil {
			klog.Info("Error creating sparse file: ", sparseFile, "Error: ", err)
			continue
		}
		c.MarkSparseBlockDeviceStateActive(sparseFile, sparseFileSize)
	}
}

// CheckAndCreateSparseFile will reuse the existing sparse file if it already exists,
// for handling cases where NDM is upgraded or restarted. If the file doesn't exist
// a new file will be created.
func CheckAndCreateSparseFile(sparseFile string, sparseFileSize int64) error {
	sparseFileInfo, err := util.SparseFileInfo(sparseFile)
	if err != nil {
		klog.Info("Check for existing file returned error: ", err)
		klog.Info("Creating a new Sparse file: ", sparseFile)
		err = util.SparseFileCreate(sparseFile, sparseFileSize)
	} else {
		klog.Info("Sparse file already exists: ", sparseFileInfo.Name())
	}
	return err
}

// GetSparseBlockDeviceUUID returns a fixed UUID for the sparse
// disk on a given node.
func GetSparseBlockDeviceUUID(hostname, sparseFile string) string {
	return SparseBlockDevicePrefix + util.Hash(hostname+sparseFile)
}

// GetActiveSparseBlockDevicesUUID returns UUIDs for the sparse
// disks present in a given node.
func GetActiveSparseBlockDevicesUUID(hostname string) []string {
	sparseFileLocation := GetSparseFileDir()
	sparseUuids := make([]string, 0)
	files, err := os.ReadDir(sparseFileLocation)
	if err != nil {
		klog.Error("Failed to read sparse file names : ", err)
		return sparseUuids
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), SparseFileName) {
			fileName := path.Join(sparseFileLocation, file.Name())
			sparseUuids = append(sparseUuids, GetSparseBlockDeviceUUID(hostname, fileName))
		}
	}
	return sparseUuids
}

// MarkSparseBlockDeviceStateActive will either create a BlockDevice CR if it doesn't exist, or it will
// update the state of the existing CR as Active.  Note that, when the NDM is going being
// gracefully shutdown, all its BlockDevice CRs are marked with State as Unknown.
func (c *Controller) MarkSparseBlockDeviceStateActive(sparseFile string, sparseFileSize int64) {
	// Fill in the details of the sparse disk
	BlockDeviceDetails := NewDeviceInfo()
	BlockDeviceDetails.UUID = GetSparseBlockDeviceUUID(c.NodeAttributes[HostNameKey], sparseFile)
	BlockDeviceDetails.NodeAttributes = c.NodeAttributes

	BlockDeviceDetails.DeviceType = blockdevice.SparseBlockDeviceType
	BlockDeviceDetails.Path = sparseFile

	sparseFileInfo, err := util.SparseFileInfo(sparseFile)
	if err != nil {
		klog.Info("Error fetching the size of sparse file: ", err)
		klog.Error("Failed to create a block device CR for sparse file: ", sparseFile)
		return
	}

	BlockDeviceDetails.Capacity = uint64(sparseFileInfo.Size())

	//If a BlockDevice CR already exits, update it. If not create a new one.
	klog.Info("Updating the BlockDevice CR for Sparse file: ", BlockDeviceDetails.UUID)
	bd, err := BlockDeviceDetails.ToDevice(c)
	if err != nil {
		klog.Error("Failed to create a block device resource CR, Error: ", err)
		return
	}

	err = c.CreateBlockDevice(bd)
	if err != nil {
		klog.Error("Failed to create a block device resource in etcd, Error: ", err)
	}
}
