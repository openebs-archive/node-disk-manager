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

On Startup, if a sparse directory (SPARSE_FILE_DIR) is specified as a
environment variable, a Sparse file with specified size (SPARSE_FILE_SIZE) will
be created and an associated Disk CR will be added to Kubernetes. By default
only one sparse file will be created which can be changed by passing the desired
number of sparse files required via the environment variable SPARSE_FILE_COUNT.

On Shutdown, the status of the sparse file Disk CR will be marked as Unknown.
*/

const (
	// ENV_SPARSE_FILE_DIR defines a sparse directory
	ENV_SPARSE_FILE_DIR = "SPARSE_FILE_DIR"
	// ENV_SPARSE_FILE_SIZE defines the sparse size
	ENV_SPARSE_FILE_SIZE = "SPARSE_FILE_SIZE"
	// ENV_SPARSE_FILE_COUNT defines the number of sparse files to be created
	ENV_SPARSE_FILE_COUNT = "SPARSE_FILE_COUNT"

	// SPARSE_FILE_NAME is the name of the sparse file
	SPARSE_FILE_NAME = "ndm-sparse.img"

	// SPARSE_FILE_DEFAULT_SIZE is the default file size
	SPARSE_FILE_DEFAULT_SIZE = int64(1073741824)
	// SPARSE_FILE_MIN_SIZE is the minimum file size
	SPARSE_FILE_MIN_SIZE = int64(1073741824)
	// SPARSE_FILE_DEFAULT_COUNT defines the default sparse count files
	SPARSE_FILE_DEFAULT_COUNT = "1"

	// SPARSE_DISKTYPE is a sparse disk type
	SPARSE_DISKTYPE = "sparse"
	// SPARSE_DISK_PREFIX defines the prefix for the sparse disk
	SPARSE_DISK_PREFIX = "sparse-"
)

// GetSparseFileDir returns the full path to the sparse
//  file directory on the node.
func GetSparseFileDir() string {

	sparseFileDir := os.Getenv(ENV_SPARSE_FILE_DIR)

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

	sparseFileCountStr := os.Getenv(ENV_SPARSE_FILE_COUNT)

	if len(sparseFileCountStr) < 1 {
		sparseFileCountStr = SPARSE_FILE_DEFAULT_COUNT
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

	sparseFileSizeStr := os.Getenv(ENV_SPARSE_FILE_SIZE)
	if len(sparseFileSizeStr) < 1 {
		glog.Info("No size was specified. Using default size: ", fmt.Sprint(SPARSE_FILE_DEFAULT_SIZE))
		return SPARSE_FILE_DEFAULT_SIZE
	}

	sparseFileSize, econv := strconv.ParseInt(sparseFileSizeStr, 10, 64)
	if econv != nil {
		glog.Info("Error converting sparse file size:  ", sparseFileSizeStr)
		return 0
	}

	if sparseFileSize < SPARSE_FILE_MIN_SIZE {
		glog.Info(fmt.Sprint(sparseFileSizeStr), " is less than minimum required. Setting the size to:  ", fmt.Sprint(SPARSE_FILE_MIN_SIZE))
		return SPARSE_FILE_MIN_SIZE
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
		sparseFile := path.Join(sparseFileDir, fmt.Sprint(i)+"-"+SPARSE_FILE_NAME)
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

// GetSparseDiskUuid returns a fixed UUID for the sparse
//  disk on a given node.
func GetSparseDiskUuid(hostname string, sparseFile string) string {
	return SPARSE_DISK_PREFIX + util.Hash(hostname+sparseFile)
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
		if strings.HasSuffix(file.Name(), SPARSE_FILE_NAME) {
			fileName := path.Join(sparseFileLocation, file.Name())
			sparseUuids = append(sparseUuids, GetSparseDiskUuid(hostname, fileName))
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
	diskDetails.Uuid = GetSparseDiskUuid(c.HostName, sparseFile)
	diskDetails.HostName = c.HostName

	diskDetails.DiskType = SPARSE_DISKTYPE
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
