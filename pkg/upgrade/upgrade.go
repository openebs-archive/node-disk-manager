/*
Copyright 2019 The OpenEBS Authors

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

package upgrade

import "fmt"

// Task interfaces gives a set of methods to be implemented
// for performing an upgrade
type Task interface {
	PreUpgrade() bool
	IsSuccess() error
}

// RunUpgrade runs all the upgrade tasks required
func RunUpgrade(tasks ...Task) error {
	for _, task := range tasks {
		_ = task.PreUpgrade()
		if err := task.IsSuccess(); err != nil {
			return fmt.Errorf("upgrade failed. Error : %v", err)
		}
	}
	return nil
}
