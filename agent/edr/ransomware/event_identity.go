package ransomware

import "time"

// A current Windows process born after an ETW event cannot be its originator.
// Missing event timestamps are insufficient authority for numeric PID response.
func creationPredatesEvent(filetime int64, event time.Time) bool {
	if filetime <= 0 || event.IsZero() {
		return false
	}
	const epoch int64 = 116444736000000000
	if filetime < epoch {
		return false
	}
	eventTicks := event.Unix()*10000000 + int64(event.Nanosecond()/100)
	return filetime-epoch <= eventTicks
}
