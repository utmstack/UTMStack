package ransomware

import "testing"

func TestConsider_KeepsCanaryTamper(t *testing.T) {
	isCanary := func(p string) bool { return p == `C:\u\00__a.xlsx` }
	ex := func(string) bool { return false }
	ev := FileEvent{PID: 40, Path: `C:\u\00__a.xlsx`, Op: OpWrite}
	if !Consider(ev, isCanary, ex, 9) {
		t.Fatal("canary write must be considered")
	}
	// A canary touch is kept even for a non-mutating op (OpCreate), proving the
	// canary branch — not the mutating-op switch — is what keeps it.
	if !Consider(FileEvent{PID: 40, Path: `C:\u\00__a.xlsx`, Op: OpCreate}, isCanary, ex, 9) {
		t.Fatal("canary touch must be kept even for a non-mutating op")
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
