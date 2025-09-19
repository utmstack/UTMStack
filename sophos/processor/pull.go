package processor

import (
	"sync"
	"time"

	"github.com/utmstack/config-client-go/types"
)

var (
	nextKeys   = make(map[int]string)
	nextKeysMu sync.RWMutex
)

func PullLogs(group types.ModuleGroup, startTime time.Time) error {
	nextKeysMu.RLock()
	prevKey := nextKeys[group.ModuleID]
	nextKeysMu.RUnlock()

	agent := getSophosCentralProcessor(group)

	logs, newNextKey, logErr := agent.getLogs(startTime.Unix(), prevKey, group)
	if logErr != nil {
		return logErr
	}

	nextKeys[group.ModuleID] = newNextKey

	sendErr := SendToLogstash(logs)
	if sendErr != nil {
		return sendErr
	}

	return nil
}
