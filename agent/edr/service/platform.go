package service

import (
	"errors"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

var errRuntimeUnverified = errors.New("operating-system feed runtime is unverified")

type combinedHealth []interface{ Health() error }

func (checks combinedHealth) Mode() string {
	for _, h := range checks {
		if m, ok := h.(interface{ Mode() string }); ok {
			return m.Mode()
		}
	}
	return ""
}

func (checks combinedHealth) Health() error {
	var unverified bool
	for _, h := range checks {
		if err := h.Health(); err != nil {
			if errors.Is(err, errRuntimeUnverified) {
				unverified = true
			} else {
				return err
			}
		}
	}
	if unverified {
		return errRuntimeUnverified
	}
	return nil
}

type unverifiedFeed struct{}

func (unverifiedFeed) Health() error { return errRuntimeUnverified }

type sensorState struct {
	Supported bool   `json:"supported"`
	Requested bool   `json:"requested"`
	State     string `json:"state"`
	Error     string `json:"error,omitempty"`
	Mode      string `json:"mode,omitempty"`
}

// runtime_unverified is intentional: starting a goroutine or accepting a
// registration request is not proof that the Windows operating system feed runs.
func platformSensors(cfg config.EDRConfig, platform string, health map[string]interface{ Health() error }) map[string]sensorState {
	requested := map[string]bool{
		"file_watcher": cfg.Sensors.FileWatcherOn(), "process_guard": cfg.Sensors.ProcessGuardOn(),
		"amsi": cfg.Sensors.AMSIOn(), "behavioral": cfg.Sensors.BehavioralOn(),
		"ransomware": cfg.Ransomware.Enabled, "network_enforcement": cfg.Blocklist.Enabled,
		"file_prevention":     cfg.LinuxProtection.Fanotify.Enabled && cfg.LinuxProtection.Effective().Fanotify.Mode != "notify",
		"file_access_monitor": cfg.LinuxProtection.Fanotify.Enabled,
		"dns_observation":     cfg.Blocklist.Enabled,
		"network_event_audit": cfg.Blocklist.Enabled,
	}
	out := make(map[string]sensorState, len(requested))
	for name, want := range requested {
		supported := platform == "windows" || (platform == "linux" && (name == "file_watcher" || name == "process_guard" || name == "behavioral" || name == "file_prevention" || name == "file_access_monitor" || name == "ransomware" && cfg.LinuxProtection.Fanotify.Enabled || name == "network_enforcement" && cfg.LinuxProtection.NetworkEnforcement))
		if (name == "file_prevention" || name == "file_access_monitor") && platform != "linux" {
			supported = false
		}
		if (name == "dns_observation" || name == "network_event_audit") && platform != "windows" {
			supported = false
		}
		st := sensorState{Supported: supported, Requested: want, State: "disabled"}
		if want && cfg.Enabled {
			if !supported {
				st.State = "unsupported"
			} else {
				st.State = "runtime_unverified"
				if (name == "file_prevention" || name == "file_access_monitor") && !cfg.Sensors.FileWatcherOn() {
					st.State = "degraded"
					st.Error = "file_watcher sensor must be enabled for configured fanotify protection"
					out[name] = st
					continue
				}
				if h := health[name]; h != nil {
					if m, ok := h.(interface{ Mode() string }); ok {
						st.Mode = m.Mode()
					}
					if err := h.Health(); errors.Is(err, errRuntimeUnverified) {
						st.State = "runtime_unverified"
					} else if err != nil {
						st.State = "degraded"
						st.Error = err.Error()
					} else {
						st.State = "active"
					}
				}
			}
		}
		out[name] = st
	}
	return out
}
