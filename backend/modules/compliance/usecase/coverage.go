package usecase

import (
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type ruleTag struct {
	Name       string   `yaml:"name"`
	Compliance []string `yaml:"compliance"`
}

type CoverageIndex struct {
	rulesDir  string // overlay root holding system/ and user/
	mu        sync.RWMutex
	byControl map[string][]string
}

func NewCoverageIndex(rulesDir string) *CoverageIndex {
	return &CoverageIndex{rulesDir: rulesDir, byControl: map[string][]string{}}
}

func (c *CoverageIndex) Load() error {
	idx := map[string][]string{}
	for _, sub := range []string{SystemSubdir, UserSubdir} {
		files, err := scanYAML(filepath.Join(c.rulesDir, sub), false)
		if err != nil {
			return err
		}
		for _, f := range files {
			if !f.enabled {
				continue
			}
			var list []ruleTag
			if err := yaml.Unmarshal(f.data, &list); err != nil || len(list) == 0 {
				continue
			}
			r := list[0]
			for _, cid := range r.Compliance {
				idx[cid] = append(idx[cid], r.Name)
			}
		}
	}
	c.mu.Lock()
	c.byControl = idx
	c.mu.Unlock()
	return nil
}

func (c *CoverageIndex) Rules(controlID string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	src := c.byControl[controlID]
	out := make([]string, len(src))
	copy(out, src)
	return out
}
