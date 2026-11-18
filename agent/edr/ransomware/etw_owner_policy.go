//go:build windows || edr_etw_policy

package ransomware

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

const maxTraceOwnerRecords = 16
const maxTraceOwnerBytes = 4096

type traceConsumerState struct {
	mu      sync.Mutex
	failure error
	closing bool
}

func (s *traceConsumerState) noteError(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	if s.failure == nil {
		s.failure = err
	}
	s.mu.Unlock()
}
func (s *traceConsumerState) Err() error { s.mu.Lock(); defer s.mu.Unlock(); return s.failure }
func validNativePropertySize(payload uint16, claim uint32) bool {
	return claim > 0 && claim <= uint32(payload)
}

type traceOwnerRecord struct {
	Version  int    `json:"version"`
	PID      uint32 `json:"pid"`
	Creation int64  `json:"creation_filetime"`
	Name     string `json:"name"`
	GUID     string `json:"session_guid"`
}

func traceOwnerName(pid uint32, creation int64) string {
	return fmt.Sprintf("UTMStackEDR-RansomFile-PID%d-Birth%d", pid, creation)
}
func canonicalTraceGUID(value string) bool {
	if len(value) != 38 || value[0] != '{' || value[37] != '}' {
		return false
	}
	for i := 1; i < 37; i++ {
		if i == 9 || i == 14 || i == 19 || i == 24 {
			if value[i] != '-' {
				return false
			}
			continue
		}
		c := value[i]
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return value != "{00000000-0000-0000-0000-000000000000}"
}
func (r traceOwnerRecord) validate() error {
	if r.Version != 1 || r.PID == 0 || r.Creation <= 0 || r.Name != traceOwnerName(r.PID, r.Creation) || !canonicalTraceGUID(r.GUID) {
		return errors.New("unverifiable ransomware ETW ownership record")
	}
	return nil
}
func (r traceOwnerRecord) filename() string { return "trace-" + strings.Trim(r.GUID, "{}") + ".json" }

type traceOwnerProcess uint8

const (
	traceProcessUnknown traceOwnerProcess = iota
	traceProcessSameAlive
	traceProcessGone
	traceProcessReused
)

type traceOwnerSnapshot struct {
	Known, Exists bool
	Name, GUID    string
	Handle        uint64
}
type traceOrphanAction uint8

const (
	traceLeaveActive traceOrphanAction = iota
	traceRemoveAbsentRecord
	traceStopOwnedOrphan
)

// A persisted record is durable intent, not evidence that StartTrace succeeded.
// A reused PID proves only the old owner gone; it never permits process action.
func decideTraceOrphan(record traceOwnerRecord, process traceOwnerProcess, native traceOwnerSnapshot) (traceOrphanAction, error) {
	if err := record.validate(); err != nil {
		return traceLeaveActive, err
	}
	if process == traceProcessUnknown || !native.Known {
		return traceLeaveActive, errors.New("ransomware ETW owner or native query is unknown")
	}
	if native.Exists && (native.Name != record.Name || native.GUID != record.GUID || native.Handle == 0) {
		return traceLeaveActive, errors.New("ransomware ETW native ownership does not match protected record")
	}
	if process == traceProcessSameAlive {
		return traceLeaveActive, nil
	}
	if process != traceProcessGone && process != traceProcessReused {
		return traceLeaveActive, errors.New("unsupported ransomware ETW owner state")
	}
	if !native.Exists {
		return traceRemoveAbsentRecord, nil
	}
	return traceStopOwnedOrphan, nil
}
