package usecase

import (
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// ruleTag is the minimal slice of a correlation rule the compliance module reads:
// its name and the control ids it claims to cover. Rules are stored as
// single-element YAML lists (the cel-plugin format), so we parse a []ruleTag and
// take the first element.
type ruleTag struct {
	Name       string   `yaml:"name"`
	Compliance []string `yaml:"compliance"`
}

// CoverageIndex maps a control id → the names of enabled correlation rules that
// tag it (via their `compliance:` block), built by scanning the eventprocessing
// rules overlay. It powers the coverage and activity report dimensions.
type CoverageIndex struct {
	rulesDir  string // overlay root holding system/ and user/
	mu        sync.RWMutex
	byControl map[string][]string
}

func NewCoverageIndex(rulesDir string) *CoverageIndex {
	return &CoverageIndex{rulesDir: rulesDir, byControl: map[string][]string{}}
}

// Load rescans the rules overlay and rebuilds the control → [ruleName] index.
// Only enabled rules (no .disabled suffix) contribute coverage.
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

// Rules returns the names of enabled rules covering the given control id.
func (c *CoverageIndex) Rules(controlID string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	src := c.byControl[controlID]
	out := make([]string, len(src))
	copy(out, src)
	return out
}
