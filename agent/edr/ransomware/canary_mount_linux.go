//go:build linux

package ransomware

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func (d *canaryDir) mountKey() (string, error) {
	f, err := os.Open(fmt.Sprintf("/proc/self/fdinfo/%d", d.file.Fd()))
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil || len(b) > 4096 {
		return "", errors.New("canary mount identity metadata unavailable")
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "mnt_id:") {
			key := strings.TrimSpace(strings.TrimPrefix(line, "mnt_id:"))
			if id, err := strconv.ParseUint(key, 10, 64); err == nil && id != 0 {
				return key, nil
			}
		}
	}
	return "", errors.New("canary mount identity unavailable")
}
