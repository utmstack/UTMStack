package usecase

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

type enabledFileYAML struct {
	Enabled []string `yaml:"enabled"`
}

type enabledSet map[string]bool

func readEnabled(dir string) (enabledSet, error) {
	data, err := os.ReadFile(filepath.Join(dir, EnabledFileName))
	if os.IsNotExist(err) {
		return enabledSet{}, nil
	}
	if err != nil {
		return nil, err
	}
	var content enabledFileYAML
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, err
	}
	out := make(enabledSet, len(content.Enabled))
	for _, rel := range content.Enabled {
		out[rel] = true
	}
	return out, nil
}

func writeEnabled(dir string, set enabledSet) error {
	list := make([]string, 0, len(set))
	for rel, on := range set {
		if on {
			list = append(list, rel)
		}
	}

	sort.Strings(list)

	data, err := yaml.Marshal(enabledFileYAML{Enabled: list})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, EnabledFileName)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
