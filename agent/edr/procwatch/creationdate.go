package procwatch

import (
	"strconv"
	"time"
)

// WMI represents CreationDate to microseconds, while GetProcessTimes uses
// 100ns FILETIME ticks since 1601. Match only complete CIM timestamps; a missing
// or wildcard timestamp cannot bind an old creation event to a current PID.
func matchesCreationDate(cim string, filetime int64) bool {
	if len(cim) != 25 || cim[14] != '.' || (cim[21] != '+' && cim[21] != '-') || filetime <= 0 {
		return false
	}
	for i, b := range []byte(cim) {
		if i == 14 || i == 21 {
			continue
		}
		if b < '0' || b > '9' {
			return false
		}
	}
	offset, err := strconv.Atoi(cim[22:])
	if err != nil {
		return false
	}
	if cim[21] == '-' {
		offset = -offset
	}
	created, err := time.ParseInLocation("20060102150405.000000", cim[:21], time.FixedZone("CIM", offset*60))
	if err != nil {
		return false
	}
	const filetimeEpochMicroseconds int64 = 11644473600000000
	return filetime/10-filetimeEpochMicroseconds == created.UnixMicro()
}
