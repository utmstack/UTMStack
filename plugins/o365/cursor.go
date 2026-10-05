package main

import "time"

// cursorKeyPrefix namespaces this plugin's keys in the shared bucket.
const cursorKeyPrefix = "o365."

func o365CursorKey(group *ModuleGroup) string {
	return cursorKeyPrefix + group.Key()
}

type cursorPayload struct {
	WindowEnd       time.Time `json:"windowEnd"`
	RiskLastUpdated time.Time `json:"riskLastUpdated"`
}
