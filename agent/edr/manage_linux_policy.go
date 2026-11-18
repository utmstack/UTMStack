package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func setLinuxProtection(cfg *config.EDRConfig, key, val string) error {
	next := cfg.LinuxProtection
	var err error
	switch key {
	case "linux_protection.process_events":
		next.ProcessEvents = val
	case "linux_protection.network_enforcement":
		next.NetworkEnforcement, err = parseBool(val)
	case "linux_protection.fanotify.enabled":
		next.Fanotify.Enabled, err = parseBool(val)
	case "linux_protection.fanotify.mounts":
		err = json.Unmarshal([]byte(val), &next.Fanotify.Mounts)
	case "linux_protection.fanotify.roots":
		err = json.Unmarshal([]byte(val), &next.Fanotify.Roots)
	case "linux_protection.fanotify.mode":
		next.Fanotify.Mode = val
	case "linux_protection.fanotify.fail_closed":
		next.Fanotify.FailClosed, err = parseBool(val)
	case "linux_protection.fanotify.deadline_ms":
		next.Fanotify.DeadlineMS, err = strconv.Atoi(val)
	case "linux_protection.fanotify.workers":
		next.Fanotify.Workers, err = strconv.Atoi(val)
	case "linux_protection.fanotify.max_pending":
		next.Fanotify.MaxPending, err = strconv.Atoi(val)
	default:
		return fmt.Errorf("unknown key: %s", key)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", key, err)
	}
	if err = next.Validate(); err != nil {
		return err
	}
	cfg.LinuxProtection = next
	return nil
}
