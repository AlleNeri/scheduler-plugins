package optimizedpreemption

import (
	"os"
	"encoding/csv"
	"encoding/json"
	"strconv"
)

type CsvBin struct {
	index uint
	memory int64
	cpu int64
	labels []string
}

type CsvPod struct {
	index uint
	memory int64
	cpu int64
	// labels []string
	bin uint
	priority int32
	affinity []string
	antiaffinity []string
}

type CsvRecord interface {
	getRecord() []string
}

func (bin CsvBin) getRecord() []string {
	if jsonString, err := json.Marshal(bin.labels); err != nil || len(bin.labels) == 0 {
		return []string{"bin", strconv.FormatUint(uint64(bin.index), 10), strconv.FormatInt(bin.memory, 10), strconv.FormatInt(bin.cpu, 10), "[]", "", "", "", ""}
	} else {
		return []string{"bin", strconv.FormatUint(uint64(bin.index), 10), strconv.FormatInt(bin.memory, 10), strconv.FormatInt(bin.cpu, 10), string(jsonString), "", "", "", ""}
	}
}

func (pod CsvPod) getRecord() []string {
	var affinity, antiaffinity string

	jsonAffinity, err := json.Marshal(pod.affinity)
	if err != nil || len(pod.affinity) == 0 {
		affinity = "[]"
	} else {
		affinity = string(jsonAffinity)
	}

	jsonAnitaffinity, err := json.Marshal(pod.antiaffinity)
	if err != nil || len(pod.antiaffinity) == 0 {
		antiaffinity = "[]"
	} else {
		antiaffinity = string(jsonAnitaffinity)
	}

	return []string{"pod", strconv.FormatUint(uint64(pod.index), 10), strconv.FormatInt(pod.memory, 10), strconv.FormatInt(pod.cpu, 10), "", strconv.FormatUint(uint64(pod.bin), 10), strconv.FormatInt(int64(pod.priority), 10), affinity, antiaffinity}
}

func printCsv(records []CsvRecord, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	// Header
	if err := csvWriter.Write([]string{"type", "index", "ram", "cpu", "label", "where", "priority", "affinity", "anti_affinity"}); err != nil {
		return err
	}

	// Records
	for _, record := range records {
		if err := csvWriter.Write(record.getRecord()); err != nil {
			return err
		}
	}

	return nil
}
