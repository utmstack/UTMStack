//go:build linux || edr_canary_policy

package ransomware

import (
	"errors"
	"testing"
)

func TestCanaryOwnerPolicyNewFileAndExistingAdmission(t *testing.T) {
	for _, create := range []bool{false, true} {
		t.Run(map[bool]string{false: "existing", true: "new"}[create], func(t *testing.T) {
			file := canaryOwnerState{uid: 123, mode: 0640}
			if create {
				file.uid = 0
			}
			calls := 0
			err := checkCanaryOwnership(123, create, func() (uint32, error) { return 123, nil }, func() (canaryOwnerState, error) { return file, nil }, func(uid uint32) error { calls++; file.uid = uid; return nil })
			if err != nil || file.uid != 123 || file.mode != 0640 {
				t.Fatalf("owner/mode admission failed: %+v %v", file, err)
			}
			want := 0
			if create {
				want = 1
			}
			if calls != want {
				t.Fatalf("Fchown calls=%d want%d", calls, want)
			}
		})
	}
	calls := 0
	err := checkCanaryOwnership(123, false, func() (uint32, error) { return 123, nil }, func() (canaryOwnerState, error) { return canaryOwnerState{uid: 0, mode: 0644}, nil }, func(uint32) error { calls++; return nil })
	if err == nil || calls != 0 {
		t.Fatal("existing wrong-owner inode was admitted or migrated", err, calls)
	}
}
func TestCanaryOwnerPolicyErrorsAndOwnershipRaceRefuseAdmission(t *testing.T) {
	for _, stage := range []string{"initial_file_stat", "chown", "final_file_stat", "owner_verify", "mode_verify", "parent_stat", "parent_changed"} {
		t.Run(stage, func(t *testing.T) {
			calls, chowns := 0, 0
			sentinel := errors.New("controlled metadata failure")
			err := checkCanaryOwnership(123, true,
				func() (uint32, error) {
					if stage == "parent_stat" {
						return 0, sentinel
					}
					if stage == "parent_changed" {
						return 124, nil
					}
					return 123, nil
				},
				func() (canaryOwnerState, error) {
					calls++
					if stage == "initial_file_stat" && calls == 1 || stage == "final_file_stat" && calls == 2 {
						return canaryOwnerState{}, sentinel
					}
					owner, mode := uint32(0), uint32(0644)
					if chowns > 0 {
						owner = 123
					}
					if stage == "owner_verify" && calls == 2 {
						owner = 0
					}
					if stage == "mode_verify" && calls == 2 {
						mode = 0666
					}
					return canaryOwnerState{uid: owner, mode: mode}, nil
				},
				func(uint32) error {
					chowns++
					if stage == "chown" {
						return sentinel
					}
					return nil
				})
			if err == nil {
				t.Fatal("metadata failure/race was admitted", stage)
			}
			if (stage == "initial_file_stat" || stage == "final_file_stat" || stage == "chown" || stage == "parent_stat") && !errors.Is(err, sentinel) {
				t.Fatalf("underlying error lost: %v", err)
			}
			if stage == "initial_file_stat" && chowns != 0 {
				t.Fatal("changed owner without initial metadata")
			}
		})
	}
}
func TestCanaryOwnerPolicyRootAndSharedDirectoriesRemainRootOwned(t *testing.T) {
	file := canaryOwnerState{uid: 0, mode: 0644}
	err := checkCanaryOwnership(0, true, func() (uint32, error) { return 0, nil }, func() (canaryOwnerState, error) { return file, nil }, func(uint32) error {
		t.Error("matching root ownership required Fchown")
		return errors.New("Fchown unavailable")
	})
	if err != nil || file.uid != 0 || file.mode != 0644 {
		t.Fatal("root directory owner/mode changed", file, err)
	}
}
func TestCanaryOwnerPolicyMatchingNewOwnerNeedsNoChownPrivilege(t *testing.T) {
	file := canaryOwnerState{uid: 123, mode: 0640}
	err := checkCanaryOwnership(123, true, func() (uint32, error) { return 123, nil }, func() (canaryOwnerState, error) { return file, nil }, func(uint32) error {
		t.Error("matching new ownership required Fchown")
		return errors.New("no chown capability")
	})
	if err != nil {
		t.Fatalf("matching creation rejected without chown privilege: %v", err)
	}
}

func TestCanaryOwnerPolicyRefusesReadOnlyUmaskWithoutChmod(t *testing.T) {
	for _, create := range []bool{false, true} {
		calls := 0
		err := checkCanaryOwnership(123, create, func() (uint32, error) { return 123, nil }, func() (canaryOwnerState, error) { return canaryOwnerState{uid: 123, mode: 0444}, nil }, func(uint32) error { calls++; return nil })
		if err == nil || calls != 0 {
			t.Fatalf("non-owner-writable mode admitted or ownership changed: create=%v err=%v calls=%d", create, err, calls)
		}
	}
}
