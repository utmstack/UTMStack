//go:build windows || edr_etw_policy

package ransomware

import (
	"errors"
	"sync"
	"testing"
)

func TestTraceOwnershipRequiresExactNativeAndDeadOwner(t *testing.T) {
	r := traceOwnerRecord{Version: 1, PID: 123, Creation: 456, GUID: "{12345678-1234-4123-8123-123456789ABC}"}
	r.Name = traceOwnerName(r.PID, r.Creation)
	n := traceOwnerSnapshot{Known: true, Exists: true, Name: r.Name, GUID: r.GUID, Handle: 9}
	for _, tc := range []struct {
		name    string
		process traceOwnerProcess
		native  traceOwnerSnapshot
		want    traceOrphanAction
		bad     bool
	}{
		{"dead-exact", traceProcessGone, n, traceStopOwnedOrphan, false},
		{"reused-old-owner", traceProcessReused, n, traceStopOwnedOrphan, false},
		{"active-exact", traceProcessSameAlive, n, traceLeaveActive, false},
		{"unknown-owner", traceProcessUnknown, n, traceLeaveActive, true},
		{"unknown-query", traceProcessGone, traceOwnerSnapshot{}, traceLeaveActive, true},
		{"absent-after-crash-before-start", traceProcessGone, traceOwnerSnapshot{Known: true}, traceRemoveAbsentRecord, false},
		{"active-before-start", traceProcessSameAlive, traceOwnerSnapshot{Known: true}, traceLeaveActive, false},
		{"foreign-name", traceProcessGone, traceOwnerSnapshot{Known: true, Exists: true, Name: r.Name + "other", GUID: r.GUID, Handle: 9}, traceLeaveActive, true},
		{"foreign-guid", traceProcessGone, traceOwnerSnapshot{Known: true, Exists: true, Name: r.Name, GUID: "{AAAAAAAA-1234-4123-8123-123456789ABC}", Handle: 9}, traceLeaveActive, true},
		{"missing-handle", traceProcessGone, traceOwnerSnapshot{Known: true, Exists: true, Name: r.Name, GUID: r.GUID}, traceLeaveActive, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			action, err := decideTraceOrphan(r, tc.process, tc.native)
			if action != tc.want || (err != nil) != tc.bad {
				t.Fatalf("action=%d err=%v", action, err)
			}
		})
	}
}

func TestTraceOwnershipRejectsTamperedRecord(t *testing.T) {
	r := traceOwnerRecord{Version: 1, PID: 123, Creation: 456, GUID: "{12345678-1234-4123-8123-123456789ABC}"}
	r.Name = traceOwnerName(r.PID, r.Creation)
	for _, change := range []func(*traceOwnerRecord){func(r *traceOwnerRecord) { r.Version++ }, func(r *traceOwnerRecord) { r.PID = 0 }, func(r *traceOwnerRecord) { r.Creation = 0 }, func(r *traceOwnerRecord) { r.Name = "UTMStackEDR-RansomFile-PID123" }, func(r *traceOwnerRecord) { r.GUID = "../../foreign" }, func(r *traceOwnerRecord) { r.GUID = "{00000000-0000-0000-0000-000000000000}" }} {
		copy := r
		change(&copy)
		if copy.validate() == nil {
			t.Fatalf("accepted %+v", copy)
		}
	}
}

func TestTraceConsumerHealthConcurrentFirstFailure(t *testing.T) {
	var state traceConsumerState
	first := errors.New("first native failure")
	state.noteError(first)
	var group sync.WaitGroup
	for i := 0; i < 32; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 100; j++ {
				state.noteError(errors.New("later failure"))
				if state.Err() != first {
					t.Error("native health lost its first failure")
				}
			}
		}()
	}
	group.Wait()
	if state.Err() != first {
		t.Fatal("missing native failure")
	}
}
func TestNativePropertySizeNeverExceedsActualPayload(t *testing.T) {
	for _, tc := range []struct {
		payload uint16
		claim   uint32
		want    bool
	}{{0, 0, false}, {8, 0, false}, {8, 4, true}, {8, 8, true}, {8, 9, false}, {65535, 65535, true}, {65535, 65536, false}, {8, ^uint32(0), false}} {
		if got := validNativePropertySize(tc.payload, tc.claim); got != tc.want {
			t.Fatalf("payload=%d claim=%d got=%v", tc.payload, tc.claim, got)
		}
	}
}
