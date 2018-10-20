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
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
	"github.com/openebs/node-disk-manager/pkg/util"

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
be created and an associated Disk CR will be added to Kubernetes. By default
only one sparse file will be created which can be changed by passing the desired
number of sparse files required via the environment variable EnvSparseFileCount.

On Shutdown, the status of the sparse file Disk CR will be marked as Unknown.
*/

const (
	//EnvSparseFileDir - if it specified as a environment variable,a Sparse file with specified size will be created
	EnvSparseFileDir = "SPARSE_FILE_DIR"
	//EnvSparseFileSize define the size of created Sparse file
	EnvSparseFileSize = "SPARSE_FILE_SIZE"
	//EnvSparseFileCount is quantity of created Sparse files
	EnvSparseFileCount = "SPARSE_FILE_COUNT"

	//SparseFileName is a name of Sparse file
	SparseFileName = "ndm-sparse.img"
	//SparseFileDefaultSize is a Sparse file default size
	SparseFileDefaultSize = int64(1073741824)
	//SparseFileMinSize is a Sparse file min size
	SparseFileMinSize = int64(1073741824)
	//SparseFileDefaultCount is a default quantity of Sparse files
	SparseFileDefaultCount = "1"

	//SparseDiskType is a Sparse disk type
	SparseDiskType = "sparse"
	//SparseDiskPrefix is a Sparse disk prefix
	SparseDiskPrefix = "sparse-"
)

// GetSparseFileDir returns the full path to the sparse
//  file directory on the node.
func GetSparseFileDir() string {

	sparseFileDir := os.Getenv(EnvSparseFileDir)

	if len(sparseFileDir) < 1 {
		return ""
	}

	info, err := os.Stat(sparseFileDir)
	if os.IsNotExist(err) || !info.Mode().IsDir() {
		glog.Info("Specified directory doesnt exist:  ", sparseFileDir)
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
		glog.Info("Error converting sparse file count:  ", sparseFileCountStr)
		return 0
	}

	return sparseFileCount
}

// GetSparseFileSize returns the size of the sparse file to be
//  created by NDM. Returns 0, if invalid size is specified.
func GetSparseFileSize() int64 {

	sparseFileSizeStr := os.Getenv(EnvSparseFileSize)
	if len(sparseFileSizeStr) < 1 {
		glog.Info("No size was specified. Using default size: ", fmt.Sprint(SparseFileDefaultSize))
		return SparseFileDefaultSize
	}

	sparseFileSize, econv := strconv.ParseInt(sparseFileSizeStr, 10, 64)
	if econv != nil {
		glog.Info("Error converting sparse file size:  ", sparseFileSizeStr)
		return 0
	}

	if sparseFileSize < SparseFileMinSize {
		glog.Info(fmt.Sprint(sparseFileSizeStr), " is less than minimum required. Setting the size to:  ", fmt.Sprint(SparseFileMinSize))
		return SparseFileMinSize
	}

	return sparseFileSize
}

// InitializeSparseFiles will check if the sparse file exist or have to be
//  created and will update or create the associated Disk CR accordingly
func (c *Controller) InitializeSparseFiles() {
	sparseFileDir := GetSparseFileDir()
	sparseFileSize := GetSparseFileSize()
	sparseFileCount := GetSparseFileCount()

	if len(sparseFileDir) < 1 || sparseFileSize < 1 || sparseFileCount < 1 {
		glog.Info("No sparse file path/size provided. Skip creating sparse files.")
		return
	}

	for i := 0; i < sparseFileCount; i++ {
		sparseFile := path.Join(sparseFileDir, fmt.Sprint(i)+"-"+SparseFileName)
		err := CheckAndCreateSparseFile(sparseFile, sparseFileSize)
		if err != nil {
			glog.Info("Error creating sparse file: ", sparseFile, "Error: ", err)
			continue
		}
		c.MarkSparseDiskStateActive(sparseFile, sparseFileSize)
	}
}

// CheckAndCreateSparseFile will reuse the existing sparse file if it already exists,
//   for handling cases where NDM is upgraded or restarted. If the file doesn't exist
//   a new file will be created.
func CheckAndCreateSparseFile(sparseFile string, sparseFileSize int64) error {
	sparseFileInfo, err := util.SparseFileInfo(sparseFile)
	if err != nil {
		glog.Info("Check for existing file returned error: ", err)
		glog.Info("Creating a new Sparse file: ", sparseFile)
		err = util.SparseFileCreate(sparseFile, sparseFileSize)
	} else {
		glog.Info("Sparse file already exists: ", sparseFileInfo.Name())
	}
	return err
}

// GetSparseDiskUUID returns a fixed UUID for the sparse
//  disk on a given node.
func GetSparseDiskUUID(hostname string, sparseFile string) string {
	return SparseDiskPrefix + util.Hash(hostname+sparseFile)
}

// GetActiveSparseDisksUuids returns UUIDs for the sparse
// disks present in a given node.
func GetActiveSparseDisksUuids(hostname string) []string {
	sparseFileLocation := GetSparseFileDir()
	sparseUuids := make([]string, 0)
	files, err := ioutil.ReadDir(sparseFileLocation)
	if err != nil {
		glog.Error("Failed to read sperse file names : ", err)
		return sparseUuids
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), SparseFileName) {
			fileName := path.Join(sparseFileLocation, file.Name())
			sparseUuids = append(sparseUuids, GetSparseDiskUUID(hostname, fileName))
		}
	}
	return sparseUuids
}

// MarkSparseDiskStateActive will either create a Disk CR if it doesn't exist, or it will
//   update the state of the existing CR as Active.  Note that, when the NDM is going being
//   gracefully shutdown, all its Disk CRs are marked with State as Unknown.
func (c *Controller) MarkSparseDiskStateActive(sparseFile string, sparseFileSize int64) {
	// Fill in the details of the sparse disk
	diskDetails := NewDiskInfo()
	diskDetails.Uuid = GetSparseDiskUUID(c.HostName, sparseFile)
	diskDetails.HostName = c.HostName

	diskDetails.DiskType = SparseDiskType
	diskDetails.Path = sparseFile

	sparseFileInfo, err := util.SparseFileInfo(sparseFile)
	if err != nil {
		glog.Info("Error fetching the size of sparse file: ", err)
		glog.Error("Failed to create a disk CR for sparse file: ", sparseFile)
		return
	}

	diskDetails.Capacity = uint64(sparseFileInfo.Size())

	//If a Disk CR already exits, update it. If not create a new one.
	glog.Info("Updating the Disk CR for Sparse Disk: ", diskDetails.Uuid)
	c.CreateDisk(diskDetails.ToDisk())

}
