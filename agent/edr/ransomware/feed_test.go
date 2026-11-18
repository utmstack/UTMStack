package ransomware

import "testing"

func TestConsider_KeepsCanaryTamper(t *testing.T) {
	isCanary := func(p string) bool { return p == `C:\u\00__a.xlsx` }
	ex := func(string) bool { return false }
	ev := FileEvent{PID: 40, Path: `C:\u\00__a.xlsx`, Op: OpWrite}
	if !Consider(ev, isCanary, ex, 9) {
		t.Fatal("canary write must be considered")
	}
	for _, op := range []FileOp{OpWrite, OpRename, OpDelete, OpSetInfo} {
		if !Consider(FileEvent{PID: 40, Path: `C:\u\00__a.xlsx`, Op: op}, isCanary, ex, 9) {
			t.Fatalf("modifying op %d on a canary must be considered", op)
		}
	}
}

func TestConsider_CanaryOpenIsNotTamper(t *testing.T) {
	// Regression (Windows Server 2022 run): Kernel-File "Create" is emitted for
	// every open, including read-only opens and attribute queries. Opening a
	// decoy is not tampering and must never reach the scorer.
	isCanary := func(p string) bool { return p == `C:\00__accounts.xlsx` }
	if Consider(FileEvent{PID: 40, Path: `C:\00__accounts.xlsx`, Op: OpCreate}, isCanary, nil, 9) {
		t.Fatal("an open of a canary must not be considered")
	}
}

func TestConsider_DropsSystemAndIdle(t *testing.T) {
	// The System process (4) issues the cache manager's lazy-writer flushes of a
	// write another process already made, and Idle (0) is not a process. Neither
	// can be the culprit nor be suspended/killed.
	isCanary := func(string) bool { return true }
	for _, pid := range []int{0, 4} {
		if Consider(FileEvent{PID: pid, Path: `C:\u\00__a.xlsx`, Op: OpWrite}, isCanary, nil, 9) {
			t.Fatalf("pid %d must be dropped", pid)
		}
	}
}

func TestConsider_CanaryKeptEvenWhenExcluded(t *testing.T) {
	isCanary := func(p string) bool { return p == `C:\u\Downloads\00__a.xlsx` }
	ex := func(p string) bool { return p == `C:\u\Downloads\00__a.xlsx` } // canary dir is also excluded
	if !Consider(FileEvent{PID: 7, Path: `C:\u\Downloads\00__a.xlsx`, Op: OpWrite}, isCanary, ex, 9) {
		t.Fatal("a canary tamper under an excluded dir must still be considered")
	}
}

func TestConsider_DropsSelfExcludedAndReads(t *testing.T) {
	isCanary := func(string) bool { return false }
	ex := func(p string) bool { return p == `C:\Program Files\UTMStack\edr` }
	// self PID dropped
	if Consider(FileEvent{PID: 9, Path: `C:\u\a.txt`, Op: OpWrite}, isCanary, ex, 9) {
		t.Error("self PID must be dropped")
	}
	// excluded path dropped
	if Consider(FileEvent{PID: 5, Path: `C:\Program Files\UTMStack\edr`, Op: OpWrite}, isCanary, ex, 9) {
		t.Error("excluded path must be dropped")
	}
	// non-mutating op dropped
	if Consider(FileEvent{PID: 5, Path: `C:\u\a.txt`, Op: OpCreate}, isCanary, ex, 9) {
		t.Error("OpCreate alone (no write/rename/delete) is not a ransomware signal in v1")
	}
}
