/*


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

package logplug

import (
	"os"
	"encoding/csv"
	"encoding/json"
	"strconv"
)

type CsvBin struct {
	memory int64
	cpu int64
	labels []string
}

type CsvPod struct {
	memory int64
	cpu int64
	labels []string
	bin int
	priority int32
	affinity []string
	antiaffinity []string
}

type CsvRecord interface {
	getRecord(index int) []string
}

func (bin CsvBin) getRecord(index int) []string {
	if jsonString, err := json.Marshal(bin.labels); err != nil {
		return []string{"bin", strconv.Itoa(index), strconv.FormatInt(bin.memory, 10), strconv.FormatInt(bin.cpu, 10), "[]"}
	} else {
		return []string{"bin", strconv.Itoa(index), strconv.FormatInt(bin.memory, 10), strconv.FormatInt(bin.cpu, 10), string(jsonString)}
	}
}

func (pod CsvPod) getRecord(index int) []string {
	var labels, affinity, antiaffinity string

	jsonLabels, err := json.Marshal(pod.labels)
	if err != nil {
		labels = "[]"
	} else {
		labels = string(jsonLabels)
	}

	jsonAffinity, err := json.Marshal(pod.affinity)
	if err != nil {
		affinity = "[]"
	} else {
		affinity = string(jsonAffinity)
	}

	jsonAnitaffinity, err := json.Marshal(pod.antiaffinity)
	if err != nil {
		antiaffinity = "[]"
	} else {
		antiaffinity = string(jsonAnitaffinity)
	}

	return []string{"pod", strconv.Itoa(index), strconv.FormatInt(pod.memory, 10), strconv.FormatInt(pod.cpu, 10), labels, strconv.Itoa(pod.bin), strconv.FormatInt(int64(pod.priority), 10), affinity, antiaffinity}
}

func printCsv(records []CsvRecord) error {
	file, err := os.Create("cluster.csv")
	if err != nil {
		return err
	}
	defer file.Close()
	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	for i, record := range records {
		if err := csvWriter.Write(record.getRecord(i)); err != nil {
			return err
		}
	}

	return nil
}
