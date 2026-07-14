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
	Events []idOnly `json:"events"`
}

func scanLines(path string, match func(line string) bool) json.RawMessage {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	for _, line := range strings.Split(string(data), "\n") {
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

func findAlertByEventID(path, eventID string) json.RawMessage {
	return scanLines(path, func(line string) bool {
		var probe alertProbe
		if err := json.Unmarshal([]byte(line), &probe); err != nil {
			return false
		}
		for _, e := range probe.Events {
			if e.ID == eventID {
				return true
			}
		}
		return false
	})
}
