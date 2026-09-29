package schema

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/utmstack/UTMStack/shared/fs"
)

// windowsEventChannelsTemplate is written as raw text, not marshalled, because
// yaml.Marshal would drop the comments documenting the file for the client.
const windowsEventChannelsTemplate = `# UTMStack - Additional Windows Event Channels
#
# Add extra Windows event log channels under "channels", one per line.
#
# The agent's built-in system channels (Security, System, Application,
# PowerShell, Windows Defender, Windows Firewall, WinLogon, ForwardedEvents)
# are always collected and cannot be disabled from this file.
#
# Channels already collected by default are ignored, so duplicates are harmless.
#
# Changes are applied within a few seconds. No service restart is required.
#
# Warning: high-volume channels can saturate the agent buffer and cause event
# loss on all channels, including Security.
#
# Many Windows channels are disabled by default. Subscribing to one succeeds,
# but no events arrive until the channel itself is enabled:
#   wevtutil gl <channel>           check the "enabled" line
#   wevtutil sl <channel> /e:true   enable it
#
# Note: "/Analytic" and "/Debug" channels do not support real-time subscription
# in Windows and will fail.
#
# Example:
# channels:
#   - Microsoft-Windows-Sysmon/Operational

channels: []
`

type windowsEventChannels struct {
	Channels []string `yaml:"channels"`
}

// ResolveWindowsEventChannels returns defaults plus the extra channels listed in
// path, creating that file from a documented template when it is missing.
// It never returns fewer channels than defaults: an unreadable or malformed file
// degrades to defaults and a non-nil error. skipped describes ignored entries.
func ResolveWindowsEventChannels(path string, defaults []string) (channels, skipped []string, err error) {
	if err := ensureWindowsEventChannels(path); err != nil {
		channels, skipped = mergeWindowsEventChannels(defaults, nil)
		return channels, skipped, err
	}

	// Removing the "channels" key but keeping the comments leaves the decoder at
	// EOF. That is an empty list, not a broken file.
	var cnf windowsEventChannels
	if err := fs.ReadYAML(path, &cnf); err != nil && !errors.Is(err, io.EOF) {
		channels, skipped = mergeWindowsEventChannels(defaults, nil)
		return channels, skipped, fmt.Errorf("error reading %s: %w", path, err)
	}

	channels, skipped = mergeWindowsEventChannels(defaults, cnf.Channels)
	return channels, skipped, nil
}

func ensureWindowsEventChannels(path string) error {
	if fs.Exists(path) {
		return nil
	}
	if err := fs.WriteString(path, windowsEventChannelsTemplate); err != nil {
		return fmt.Errorf("error creating %s: %w", path, err)
	}
	return nil
}

// mergeWindowsEventChannels appends custom to defaults, dropping blank and
// duplicate entries. Windows channel names are case-insensitive, so matching is
// too. Every default is always present, and first.
func mergeWindowsEventChannels(defaults, custom []string) (channels, skipped []string) {
	channels = make([]string, 0, len(defaults)+len(custom))
	seen := make(map[string]struct{}, len(defaults)+len(custom))

	add := func(channel string) bool {
		key := strings.ToLower(channel)
		if _, duplicate := seen[key]; duplicate {
			return false
		}
		seen[key] = struct{}{}
		channels = append(channels, channel)
		return true
	}

	for _, channel := range defaults {
		add(channel)
	}

	for _, entry := range custom {
		channel := strings.TrimSpace(entry)
		if channel == "" {
			skipped = append(skipped, "blank entry ignored")
			continue
		}
		if !add(channel) {
			skipped = append(skipped, fmt.Sprintf("%q is already collected, ignored", channel))
		}
	}

	return channels, skipped
}
