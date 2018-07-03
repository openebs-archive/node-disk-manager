/*
Copyright 2015 The Prometheus Authors.
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
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	diskType       = "disk"
	diskSubsystem  = ""
	diskSectorSize = 512
)

var (
	previousDiskStats = map[string][]string{}
	diskUUID          = map[string]string{}
	ignoredDevices    = "^(ram|loop|fd|(s|h|v|xv)d[a-z]\\d+n\\d+p)\\d+$"
)

type typedFactorDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
	factor    float64
}

type diskstatsCollector struct {
	ignoredDevicesPattern *regexp.Regexp
	descs                 []typedFactorDesc
}

func init() {
	getDiskStats()
	registerCollector("diskstats", defaultEnabled, NewDiskstatsCollector)
}

func (d *typedFactorDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	if d.factor != 0 {
		value *= d.factor
	}
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}

func (c *diskstatsCollector) Update(ch chan<- prometheus.Metric) error {
	procDiskStats := procFilePath("diskstats")
	diskStats, err := getDiskStats()
	if err != nil {
		return fmt.Errorf("couldn't get diskstats: %s", err)
	}

	for dev, stats := range diskStats {
		if c.ignoredDevicesPattern.MatchString(dev) {
			continue
		}

		if len(stats) != len(c.descs) {
			return fmt.Errorf("invalid line for %s for %s", procDiskStats, dev)
		}

		for i, value := range stats {
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value %s in diskstats: %s", value, err)
			}
			if diskUUID[dev] != "" {
				ch <- c.descs[i].mustNewConstMetric(v, diskUUID[dev])
			}
		}
	}
	return nil
}

func getDiskStats() (map[string]map[int]string, error) {
	file, err := os.Open(procFilePath("diskstats"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseDiskStats(file)
}

func parseDiskStats(r io.Reader) (map[string]map[int]string, error) {
	var (
		diskStats = map[string]map[int]string{}
		scanner   = bufio.NewScanner(r)
	)
	for scanner.Scan() {
		/*
			These are the info from /proc/diskstats file
			content of a single line is like this -
			8 0 sda 28987 5911 2163548 1947012 11137 17181 699690 269512 0 182512 2216872
			Field 1 - #major number
			Field 2 - #minor mumber
			Field 3 - #device name
			Field 4 - #reads completed successfully
			Field 5 - #reads merged
			Field 6 - #sectors read
			Field 7 - #time spent reading (ms)
			Field 8 - #writes completed
			Field 9 - #writes merged
			Field 10 - #sectors written
			Field 11 - #time spent writing (ms)
			Field 12 - #I/Os currently in progress
			Field 13 - #time spent doing I/Os (ms)
			Field14 - #weighted time spent doing I/Os (ms)
		*/
		parts := strings.Fields(scanner.Text())
		if len(parts) < 4 {
			// strip major, minor and dev
			return nil, fmt.Errorf("invalid line in %s: %s", procFilePath("diskstats"), scanner.Text())
		}
		dev := parts[2]   //To get device name
		disk := parts[3:] // strip major, minor and dev
		diskStats[dev] = map[int]string{}
		index := 0
		previousDisk := previousDiskStats[dev]
		var readLatency, writeLatency float64 = 0, 0
		var avgReadSize, avgWriteSize float64 = 0, 0
		var previousTotalBlockRead, previousTotalBlockWrite float64
		var previousTotalRead, previousTotalWrite float64
		var previousTotalReadTime, previousTotalWriteTime float64
		var totalBlockRead, totalBlockWrite float64
		var totalRead, totalWrite float64
		var totalReadTime, totalWriteTime float64
		var fsectorSize float64

		if previousDisk != nil {
			sectorSize, err := getSectorSize(dev)
			if err != nil {
				sectorSize = "512"
			}
			fsectorSize, _ = strconv.ParseFloat(sectorSize, 64)

			totalRead, _ = strconv.ParseFloat(disk[0], 64)
			totalBlockRead, _ = strconv.ParseFloat(disk[2], 64)
			totalReadTime, _ = strconv.ParseFloat(disk[3], 64)
			totalWrite, _ = strconv.ParseFloat(disk[4], 64)
			totalBlockWrite, _ = strconv.ParseFloat(disk[6], 64)
			totalWriteTime, _ = strconv.ParseFloat(disk[7], 64)

			previousTotalRead, _ = strconv.ParseFloat(previousDisk[0], 64)
			previousTotalBlockRead, _ = strconv.ParseFloat(previousDisk[2], 64)
			previousTotalReadTime, _ = strconv.ParseFloat(previousDisk[3], 64)
			previousTotalWrite, _ = strconv.ParseFloat(previousDisk[4], 64)
			previousTotalBlockWrite, _ = strconv.ParseFloat(previousDisk[6], 64)
			previousTotalWriteTime, _ = strconv.ParseFloat(previousDisk[7], 64)

			if (totalRead - previousTotalRead) != 0 {
				avgReadSize = (totalBlockRead - previousTotalBlockRead) / (totalRead - previousTotalRead)
				avgReadSize *= fsectorSize
			}
			if (totalRead - previousTotalRead) != 0 {
				avgWriteSize = (totalBlockWrite - previousTotalBlockWrite) / (totalWrite - previousTotalWrite)
				avgWriteSize *= fsectorSize
			}
			if (totalBlockRead - previousTotalBlockRead) != 0 {
				readLatency = (totalReadTime - previousTotalReadTime) / (totalBlockRead - previousTotalBlockRead)
				readLatency /= fsectorSize
			}
			if (totalBlockWrite - previousTotalBlockWrite) != 0 {
				writeLatency = (totalWriteTime - previousTotalWriteTime) / (totalBlockWrite - previousTotalBlockWrite)
				writeLatency /= fsectorSize
			}
		}
		previousDiskStats[dev] = disk
		diskStats[dev][index] = disk[0] // metrics for #reads completed successfully
		index++
		diskStats[dev][index] = fmt.Sprintf("%f", totalBlockRead*fsectorSize) // metrics for #sectors read * sector size
		index++
		diskStats[dev][index] = disk[4] // metrics for #writes completed successfully
		index++
		diskStats[dev][index] = fmt.Sprintf("%f", totalBlockWrite*fsectorSize) // metrics for #sectors write * sector size
		index++
		diskStats[dev][index] = fmt.Sprintf("%f", avgReadSize)
		index++
		diskStats[dev][index] = fmt.Sprintf("%f", avgWriteSize)
		index++
		diskStats[dev][index] = fmt.Sprintf("%f", readLatency)
		index++
		diskStats[dev][index] = fmt.Sprintf("%f", writeLatency)
	}
	return diskStats, scanner.Err()
}

func getSectorSize(deviceName string) (string, error) {
	sectorSize := ""
	file, err := os.Open("/sys/block/" + deviceName + "/queue/hw_sector_size")
	if err != nil {
		return sectorSize, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sectorSize = scanner.Text()
	}
	return sectorSize, err
}
