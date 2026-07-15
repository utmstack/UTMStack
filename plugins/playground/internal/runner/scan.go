package runner

import (
	"encoding/json"
	"os"
	"strings"
)

type idOnly struct {
	ID string `json:"id"`
}

type alertProbe struct {
	ID     string   `json:"id"`
	Events []idOnly `json:"events"`
}

func scanLines(path string, match func(line string) bool) json.RawMessage {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if match(line) {
			return json.RawMessage(line)
		}
	}
	return nil
}

func findByID(path, id string) json.RawMessage {
	return scanLines(path, func(line string) bool {
		var probe idOnly
		if err := json.Unmarshal([]byte(line), &probe); err != nil {
			return false
		}
		return probe.ID == id
	})
}

func findAlertsByEventID(path, eventID string, seenIDs map[string]struct{}) []json.RawMessage {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var out []json.RawMessage
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var probe alertProbe
		if err := json.Unmarshal([]byte(line), &probe); err != nil {
			continue
		}

		matched := false
		for _, e := range probe.Events {
			if e.ID == eventID {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		if probe.ID != "" {
			if _, seen := seenIDs[probe.ID]; seen {
				continue
			}
			seenIDs[probe.ID] = struct{}{}
		}

		out = append(out, json.RawMessage(line))
	}
	return out
}
