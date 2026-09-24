package repository

import (
	"testing"

	"github.com/utmstack/utmstack/backend/modules/adaudit/domain"
)

func str(s string) *string { return &s }
func uid(n uint32) *uint32 { return &n }

// The unique index on (tenant, machine, uid) cannot tell rows apart when the
// uid is missing — NULLs never collide — so an account seen without a uid was
// inserted again on every observation. Only a fully keyed account may take the
// conflict-based path.
func TestSplitByIdentity(t *testing.T) {
	keyed := domain.ADUser{Source: "linux", MachineID: str("m-1"), UIDNumber: uid(1000), Hostname: str("h"), Username: str("root")}
	machineWithoutUID := domain.ADUser{Source: "linux", MachineID: str("m-1"), Hostname: str("h"), Username: str("root")}
	uidWithoutMachine := domain.ADUser{Source: "linux", UIDNumber: uid(1000), Hostname: str("h"), Username: str("root")}
	blankMachine := domain.ADUser{Source: "linux", MachineID: str(""), UIDNumber: uid(1000), Hostname: str("h"), Username: str("root")}
	windows := domain.ADUser{Source: "windows", SID: str("S-1-5-21-1")}
	unlabelled := domain.ADUser{SID: str("S-1-5-21-2")}

	win, linuxKeyed, byAccount := splitByIdentity([]domain.ADUser{keyed, machineWithoutUID, uidWithoutMachine, blankMachine, windows, unlabelled})

	if len(linuxKeyed) != 1 || linuxKeyed[0].UIDNumber == nil || *linuxKeyed[0].UIDNumber != 1000 || *linuxKeyed[0].MachineID != "m-1" {
		t.Fatalf("only the account with machine and uid is keyed, got %+v", linuxKeyed)
	}
	if len(byAccount) != 3 {
		t.Fatalf("machine without uid, uid without machine and a blank machine are matched by account, got %d", len(byAccount))
	}
	if len(win) != 2 {
		t.Fatalf("windows accounts keep the SID path, got %d", len(win))
	}
	if win[1].Source != "windows" {
		t.Fatalf("an unlabelled account defaults to windows, got %q", win[1].Source)
	}
}
