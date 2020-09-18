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
