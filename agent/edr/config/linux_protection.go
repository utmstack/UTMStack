package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

// Native permission and network controls are opt-in. A Linux host keeps its
// post-arrival baseline until an administrator selects explicit protected scope.
type LinuxProtection struct {
	ProcessEvents      string        `json:"process_events"`
	NetworkEnforcement bool          `json:"network_enforcement"`
	Fanotify           LinuxFanotify `json:"fanotify"`
}
type LinuxFanotify struct {
	Enabled    bool     `json:"enabled"`
	Mounts     []string `json:"mounts"`
	Roots      []string `json:"roots"`
	Mode       string   `json:"mode"`
	FailClosed bool     `json:"fail_closed"`
	DeadlineMS int      `json:"deadline_ms"`
	Workers    int      `json:"workers"`
	MaxPending int      `json:"max_pending"`
}

func (l *LinuxProtection) UnmarshalJSON(b []byte) error {
	type plain LinuxProtection
	var p plain
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return fmt.Errorf("linux_protection: %w", err)
	}
	*l = LinuxProtection(p)
	return nil
}
func (l LinuxProtection) Effective() LinuxProtection {
	if l.ProcessEvents == "" {
		l.ProcessEvents = "auto"
	}
	f := &l.Fanotify
	if f.Mode == "" {
		f.Mode = "notify"
	}
	if f.DeadlineMS == 0 {
		f.DeadlineMS = 1500
	}
	if f.Workers == 0 {
		f.Workers = 2
	}
	if f.MaxPending == 0 {
		f.MaxPending = 32
	}
	f.Mounts = append([]string(nil), f.Mounts...)
	f.Roots = append([]string(nil), f.Roots...)
	return l
}
func (l LinuxProtection) Validate() error {
	e := l.Effective()
	switch e.ProcessEvents {
	case "auto", "connector", "poll":
	default:
		return fmt.Errorf("linux_protection.process_events must be auto, connector or poll")
	}
	f := e.Fanotify
	switch f.Mode {
	case "notify", "permission_exec", "permission_open":
	default:
		return fmt.Errorf("linux_protection.fanotify.mode must be notify, permission_exec or permission_open")
	}
	for _, v := range []struct {
		key         string
		n, min, max int
	}{{"deadline_ms", f.DeadlineMS, 100, 5000}, {"workers", f.Workers, 1, 8}, {"max_pending", f.MaxPending, 1, 64}} {
		if v.n < v.min || v.n > v.max {
			return fmt.Errorf("linux_protection.fanotify.%s must be %d..%d (zero selects default)", v.key, v.min, v.max)
		}
	}
	if len(f.Mounts) > 16 || len(f.Roots) > 64 {
		return fmt.Errorf("linux_protection.fanotify scope exceeds 16 mounts or 64 roots")
	}
	if f.MaxPending < f.Workers {
		return fmt.Errorf("fanotify.max_pending must be at least workers")
	}
	if f.Enabled && (len(f.Mounts) == 0 || len(f.Roots) == 0) {
		return fmt.Errorf("enabled fanotify requires explicit mounts and roots")
	}
	for _, list := range [][]string{f.Mounts, f.Roots} {
		for _, p := range list {
			if len(p) > 4096 || !path.IsAbs(p) || strings.ContainsRune(p, 0) || path.Clean(p) != p {
				return fmt.Errorf("fanotify scope requires clean absolute Linux paths of at most 4096 bytes")
			}
			for _, pseudo := range []string{"/proc", "/sys", "/dev", "/run"} {
				if p == pseudo || strings.HasPrefix(p, pseudo+"/") {
					return fmt.Errorf("fanotify pseudo-filesystem scopes are unavailable")
				}
			}
		}
	}
	for _, root := range f.Roots {
		inside := false
		for _, mount := range f.Mounts {
			if mount == "/" || root == mount || strings.HasPrefix(root, mount+"/") {
				inside = true
				break
			}
		}
		if !inside {
			return fmt.Errorf("fanotify root %q is outside configured mounts", root)
		}
	}
	return nil
}
