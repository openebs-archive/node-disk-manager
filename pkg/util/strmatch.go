/*
Copyright 2018 The OpenEBS Authors.

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

import (
	"regexp"
	"strings"
)

// Contains is a util function which returns true if one key is present in array
// else it returns false
func Contains(s []string, k string) bool {
	for _, e := range s {
		if e == k {
			return true
		}
	}
	return false
}

// ContainsIgnoredCase is a util function which returns true if one key is present
// in array else it returns false. This function is not case sensitive.
func ContainsIgnoredCase(s []string, k string) bool {
	for _, e := range s {
		if strings.ToLower(e) == strings.ToLower(k) {
			return true
		}
	}
	return false
}

// MatchIgnoredCase is a util function which returns true if any of the keys
// are present as a string in given string - s
// This function is not case sensitive.
func MatchIgnoredCase(keys []string, s string) bool {
	for _, k := range keys {
		if strings.Contains(strings.ToLower(s), strings.ToLower(k)) {
			return true
		}
	}
	return false
}

// MatchRegex is a utility function which returns true if the string -s
// matches with the regex specified.
func MatchRegex(regex, s string) bool {
	r := regexp.MustCompile(regex)
	return r.MatchString(s)
}
