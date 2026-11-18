package signatures

import (
	"encoding/json"
	"os"
	"testing"
)

// This optional installed-feed check reports actual native part coverage;
// accepting a database's syntax is not evidence that this stream handles it.
func TestNativeACInstalledFeedCoverage(t *testing.T) {
	dir := os.Getenv("EDR_NATIVE_AC_DB_PATH")
	if dir == "" {
		t.Skip("requires a retained signature database")
	}
	d, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	type sample struct {
		Name    string `json:"rule"`
		Pattern int    `json:"pattern"`
		Reason  string `json:"reason"`
	}
	type row struct {
		Root          int            `json:"root"`
		Active        bool           `json:"native_stream"`
		Registrations int            `json:"registrations"`
		Patterns      int            `json:"patterns"`
		Unmodeled     int            `json:"unmodeled"`
		Reasons       map[string]int `json:"reasons,omitempty"`
		Samples       []sample       `json:"samples,omitempty"`
	}
	rows := make([]row, maxTargets)
	seen := [maxTargets]map[int]bool{}
	for root := range rows {
		rows[root] = row{Root: root, Active: d.nativeAC[root] != nil, Reasons: map[string]int{}}
		seen[root] = map[int]bool{}
	}
	for i := range d.rules {
		r := &d.rules[i]
		if !acRegistration(d, r) {
			continue
		}
		for _, id := range r.patterns {
			p := &d.patterns[id]
			if p.holder != nil {
				continue
			}
			row := &rows[r.target]
			row.Registrations++
			if seen[r.target][id] {
				continue
			}
			seen[r.target][id] = true
			row.Patterns++
			modeled := p.graph != nil && p.graph.native != nil || p.simple && !p.nativeAC && len(p.prefix) >= 2
			if !modeled {
				row.Unmodeled++
				reason := "missing native part model"
				if p.graph != nil && p.graph.nativeError != "" {
					reason = p.graph.nativeError
				}
				row.Reasons[reason]++
				if len(row.Samples) < 12 {
					row.Samples = append(row.Samples, sample{r.name, id, reason})
				}
			}
		}
	}
	for _, row := range rows {
		if row.Registrations > 0 {
			t.Logf("root%d: active=%v registrations=%d patterns=%d unmodeled=%d reasons=%v", row.Root, row.Active, row.Registrations, row.Patterns, row.Unmodeled, row.Reasons)
			if !row.Active || row.Unmodeled != 0 {
				t.Errorf("root%d has ordinary AC registrations outside the native part stream", row.Root)
			}
		}
	}
	if out := os.Getenv("EDR_NATIVE_AC_COVERAGE_OUT"); out != "" {
		data, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(out, append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
