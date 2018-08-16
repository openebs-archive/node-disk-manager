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

package util

import "os"

//import "syscall"

//SparseFileCreate will create a new sparse file if none exists
// at the give path and will set the size to specified value
func SparseFileCreate(path string, size int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return f.Truncate(size)
}

//SparseFileDelete will delete the sparse file if it exists
func SparseFileDelete(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

//SparseFileInfo will return the stats of the sparse file
func SparseFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
