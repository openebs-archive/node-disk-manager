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
package metrics

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiskStats(t *testing.T) {
	newFile, err := os.Create("/tmp/diskstats")
	if err != nil {
		t.Fatal(err)
	}
	str := "   7       0 loop0 0 0 0 0 0 0 0 0 0 0 0\n   7       1 loop1 0 0 0 0 0 0 0 0 0 0 0\n   7       2 loop2 0 0 0 0 0 0 0 0 0 0 0\n   7       3 loop3 0 0 0 0 0 0 0 0 0 0 0\n   7       4 loop4 0 0 0 0 0 0 0 0 0 0 0\n   7       5 loop5 0 0 0 0 0 0 0 0 0 0 0\n   7       6 loop6 0 0 0 0 0 0 0 0 0 0 0\n   7       7 loop7 0 0 0 0 0 0 0 0 0 0 0\n   8       0 sda 495701 23442 94355120 8274748 243909 213921 99432746 83971964 0 2551868 92253708\n   8       1 sda1 65 10 4656 3828 814 2202 24128 191208 0 142896 195032\n   8       2 sda2 129 1000 12306 3276 2 0 2 0 0 2632 3276\n   8       3 sda3 12684 147 3271826 51408 10 5 120 1196 0 38716 52600\n   8       4 sda4 482645 22285 91053554 8208412 238064 211714 99408488 83595424 0 2405948 91816104\n   8       5 sda5 40 0 4128 3936 0 0 0 0 0 2352 3936\n   8       6 sda6 60 0 4282 3164 1 0 8 24 0 2740 3188\n  11       0 sr0 0 0 0 0 0 0 0 0 0 0 0\n 179       0 mmcblk0 186 3930 12415 460 1 0 1 4 0 244 464\n 179       1 mmcblk0p1 159 3930 10311 380 1 0 1 4 0 196 384\n"
	newFile.Write([]byte(str))
	defer newFile.Close()
	file, err := os.Open("/tmp/diskstats")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	defer func() {
		err := os.Remove("/tmp/diskstats")
		if err != nil {
			t.Fatal(err)
		}
	}()
	diskStats, err := parseDiskStats(file)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "482645", diskStats["sda4"][0])
	assert.Equal(t, "238064", diskStats["sda4"][2])
	if diskStats["loop"] != nil {
		t.Errorf("don't want diskstats loop")
	}
}
