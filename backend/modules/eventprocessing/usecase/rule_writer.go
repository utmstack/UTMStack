package usecase

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// writeRuleFile marshals a rule to YAML and writes it atomically to path
// (creating parent directories as needed). The write goes to a sibling temp
// file which is then renamed into place, so a reader never sees a partial file.
func writeRuleFile(path string, rule Rule) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal([]Rule{rule})
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func readRuleFile(path string) (Rule, error) {
	var rule Rule
	data, err := os.ReadFile(path)
	if err != nil {
		return rule, err
	}
	var list []Rule
	if err := yaml.Unmarshal(data, &list); err != nil {
		return rule, err
	}
	if len(list) == 0 {
		return rule, fmt.Errorf("rule file %s contains no rules", path)
	}
	return list[0], nil
}

// renameRuleFile moves a rule file (used to toggle the .disabled suffix). It is
// a no-op when src == dst.
func renameRuleFile(src, dst string) error {
	if src == dst {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

// removeRuleFile deletes a rule file, tolerating a missing file.
func removeRuleFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
