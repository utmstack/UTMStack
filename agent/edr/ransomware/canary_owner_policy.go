//go:build linux || edr_canary_policy

package ransomware

import (
	"errors"
	"fmt"
)

type canaryOwnerState struct{ uid, mode uint32 }

// checkCanaryOwnership accepts only owner-writable decoys in their pinned
// directory's ownership domain. Only a freshly exclusive-created descriptor may
// change owner; existing or adopted files are validated without migration.
func checkCanaryOwnership(parent uint32, create bool, parentNow func() (uint32, error), fileNow func() (canaryOwnerState, error), setOwner func(uint32) error) error {
	before, err := fileNow()
	if err != nil {
		return fmt.Errorf("canary initial ownership metadata: %w", err)
	}
	if before.mode&0200 == 0 {
		return errors.New("canary mode does not permit its directory owner to write; mode is not changed")
	}
	if !create && before.uid != parent {
		return errors.New("existing canary owner differs from pinned directory owner; no ownership migration performed")
	}
	if create && before.uid != parent {
		if err = setOwner(parent); err != nil {
			return fmt.Errorf("new canary descriptor ownership: %w", err)
		}
	}
	after, err := fileNow()
	if err != nil {
		return fmt.Errorf("canary final ownership metadata: %w", err)
	}
	if after.uid != parent {
		return errors.New("canary descriptor owner verification failed")
	}
	if after.mode != before.mode {
		return errors.New("canary mode changed during ownership admission")
	}
	current, err := parentNow()
	if err != nil {
		return fmt.Errorf("canary directory final ownership metadata: %w", err)
	}
	if current != parent {
		return errors.New("canary pinned directory owner changed during admission")
	}
	return nil
}
