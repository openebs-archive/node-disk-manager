/*
Copyright 2019 OpenEBS Authors.

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

package cleaner

import "os"

const (
	// EnvCleanUpJobImage is the environment variable for getting the
	// job container image
	EnvCleanUpJobImage = "CLEANUP_JOB_IMAGE"
)

var (
	// defaultCleanUpJobImage is the default job container image
	defaultCleanUpJobImage = "openebs/linux-utils:3.9"
)

// getCleanUpImage gets the image to be used for the cleanup job
func getCleanUpImage() string {
	image, ok := os.LookupEnv(EnvCleanUpJobImage)
	if !ok {
		return defaultCleanUpJobImage
	}
	return image
}
