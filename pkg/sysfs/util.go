/*
Copyright 2020 The OpenEBS Authors

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

package sysfs

import (
	"io/ioutil"
	"strconv"
	"strings"
)

// readSysFSFileAsInt64 reads a file and
// converts that content into int64
func readSysFSFileAsInt64(sysFilePath string) (int64, error) {
	b, err := ioutil.ReadFile(sysFilePath)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSuffix(string(b), "\n"), 10, 64)
}

func readSysFSFileAsString(sysFilePath string) (string, error) {
	b, err := ioutil.ReadFile(sysFilePath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// addDevPrefix adds the /dev prefix to all the device names
func addDevPrefix(devNames []string) []string {
	result := make([]string, 0)
	for _, devName := range devNames {
		result = append(result, "/dev/"+devName)
	}
	return result
}
