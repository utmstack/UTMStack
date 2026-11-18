//go:build windows

package netblock

import (
	"strings"
	"sync"

	"golang.org/x/sys/windows"
)

var (
	devMapOnce sync.Once
	devMap     map[string]string
)

func buildDevMap() {
	devMap = map[string]string{}
	var buf [1024]uint16
	for c := byte('A'); c <= 'Z'; c++ {
		letter := string(rune(c)) + ":"
		lp, err := windows.UTF16PtrFromString(letter)
		if err != nil {
			continue
		}
		n, err := windows.QueryDosDevice(lp, &buf[0], uint32(len(buf)))
		if err != nil || n == 0 {
			continue
		}
		if dev := windows.UTF16ToString(buf[:n]); dev != "" {
			devMap[strings.ToLower(dev)] = letter
		}
	}
}

// NormalizeDevicePath rewrites "\Device\HarddiskVolumeN\..." → "C:\...".
func NormalizeDevicePath(p string) string {
	p = stripQ(p)
	if strings.HasPrefix(strings.ToLower(p), `\device\`) {
		devMapOnce.Do(buildDevMap)
		low := strings.ToLower(p)
		for dev, letter := range devMap {
			if strings.HasPrefix(low, dev+`\`) {
				return letter + p[len(dev):]
			}
		}
	}
	return p
}
