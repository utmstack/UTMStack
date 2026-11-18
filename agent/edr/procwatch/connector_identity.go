package procwatch

// connectorBirthPredates bounds /proc's floored BOOTTIME start tick against a
// MONOTONIC connector event. A clock bracket includes suspend offset and read
// uncertainty. Ambiguity must remain a current snapshot, never an original exec.
func connectorBirthPredates(event uint64, ticks int64, hz, monoBefore, boot, monoAfter, subscriptionOffsetLower uint64) bool {
	if event == 0 || ticks <= 0 || hz == 0 || hz > 1_000_000 || monoBefore > monoAfter || boot < monoBefore || event > monoAfter {
		return false
	}
	n := uint64(ticks) + 1
	q, r := n/hz, n%hz
	const billion = uint64(1_000_000_000)
	const max = ^uint64(0)
	if q > max/billion {
		return false
	}
	upper := q*billion + (r*billion+hz-1)/hz
	if upper < q*billion {
		return false
	}
	var offsetLower uint64
	if boot > monoAfter {
		offsetLower = boot - monoAfter
	}
	// A current offset must never lend authority to a pre-suspend queued event.
	// The lower bound captured before subscription remains valid across suspend.
	if offsetLower > subscriptionOffsetLower {
		offsetLower = subscriptionOffsetLower
	}
	if event > max-offsetLower {
		return false
	}
	return upper <= event+offsetLower
}
