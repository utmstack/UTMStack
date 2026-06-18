package recovery

import (
	"fmt"
	"time"
)

func doShutdown(timeout time.Duration) error {
	// Step 1+2: Signal shutdown and cancel context.
	Stop()

	// Step 3: Wait for in-flight dispatches to drain.
	_ = WaitDone(timeout)

	// Step 4: Roll back DISPATCHED → PENDING with attempts--.
	disp.mu.Lock()
	db := disp.db
	disp.mu.Unlock()

	if db == nil {
		// Dispatcher was never started — nothing to roll back.
		return nil
	}

	count, err := RollbackDispatchedToPending(db)
	if err != nil {
		return fmt.Errorf("shutdown rollback: %w", err)
	}

	LogRecoveryShutdownRollback(int(count))
	return nil
}
