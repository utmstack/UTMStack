package utils

import (
	"encoding/csv"
	"os"

	"gopkg.in/yaml.v2"
)

func ReadYaml(url string, result interface{}) {
	f, err := os.Open(url)
	if err != nil {
		h.Error("Could not open file: %v", err)
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	if err := d.Decode(result); err != nil {
		h.Error("Could not decode YAML: %v", err)
	}
}

func ReadCSV(url string) [][]string {
	f, err := os.Open(url)
	if err != nil {
		h.Error("Could not open file: %v", err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	result, err := r.ReadAll()
	if err != nil {
		h.Error("Could not read CSV: %v", err)
	}
	return result
}
