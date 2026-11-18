package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
)

// ScheduledScan is opt-in. A sweep enumerates files into the shared bounded
// scan queue; its completion does not mean every queued analysis has finished.
type ScheduledScan struct {
	Enabled            bool     `json:"enabled"`
	Paths              []string `json:"paths"`
	IntervalMinutes    int      `json:"interval_minutes"`
	RunOnStart         bool     `json:"run_on_start"`
	MaxFiles           int      `json:"max_files"`
	MaxEntries         int      `json:"max_entries"`
	MaxDepth           int      `json:"max_depth"`
	MaxDurationMinutes int      `json:"max_duration_minutes"`
}

func (s *ScheduledScan) UnmarshalJSON(data []byte) error {
	type plain ScheduledScan
	var p plain
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return fmt.Errorf("scheduled_scan: %w", err)
	}
	*s = ScheduledScan(p)
	return nil
}

func (s ScheduledScan) Effective() ScheduledScan {
	if s.IntervalMinutes == 0 {
		s.IntervalMinutes = 240
	}
	if s.MaxFiles == 0 {
		s.MaxFiles = 100000
	}
	if s.MaxEntries == 0 {
		s.MaxEntries = 200000
	}
	if s.MaxDepth == 0 {
		s.MaxDepth = 64
	}
	if s.MaxDurationMinutes == 0 {
		s.MaxDurationMinutes = 60
	}
	s.Paths = append([]string(nil), s.Paths...)
	return s
}

func (s ScheduledScan) Validate() error {
	for _, b := range []struct {
		key    string
		n, max int
	}{
		{"interval_minutes", s.IntervalMinutes, 43200},
		{"max_files", s.MaxFiles, 500000},
		{"max_entries", s.MaxEntries, 1000000},
		{"max_depth", s.MaxDepth, 128},
		{"max_duration_minutes", s.MaxDurationMinutes, 1440},
	} {
		if b.n < 0 || b.n > b.max {
			return fmt.Errorf("scheduled_scan.%s must be from 0 to %d (0 = default)", b.key, b.max)
		}
	}
	if len(s.Paths) > 128 {
		return fmt.Errorf("scheduled_scan.paths accepts at most 128 paths")
	}
	if s.Enabled && len(s.Paths) == 0 {
		return fmt.Errorf("scheduled_scan.paths is required when enabled")
	}
	for _, p := range s.Paths {
		if len(p) > 4096 || !filepath.IsAbs(p) {
			return fmt.Errorf("scheduled_scan.paths must contain absolute paths of at most 4096 bytes")
		}
	}
	return nil
}
