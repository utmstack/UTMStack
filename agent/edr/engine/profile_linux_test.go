package engine

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestEffectiveLinuxRAM(t *testing.T) {
	const host = uint64(16 << 30)
	for _, tc := range []struct {
		name, v2, v1 string
		want         uint64
	}{
		{"no-cgroup", "", "", host},
		{"v2-limit", "1073741824\n", "", 1 << 30},
		{"v2-unlimited", "max\n", "", host},
		{"v1-limit", "", "2147483648\n", 2 << 30},
		{"v1-unlimited-sentinel", "", "9223372036854771712\n", host},
		{"invalid", "not-a-limit", "-1", host},
		{"zero", "0", "0", host},
		{"larger-than-host", "34359738368", "", host},
		{"minimum-of-present-limits", "2147483648", "1073741824", 1 << 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.v2 != "" {
				if err := os.WriteFile(filepath.Join(root, "memory.max"), []byte(tc.v2), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.v1 != "" {
				path := filepath.Join(root, "memory", "memory.limit_in_bytes")
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(tc.v1), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if got := effectiveLinuxRAM(host, root); got != tc.want {
				t.Fatalf("effectiveLinuxRAM() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestLinuxHostProbeHonorsCgroup(t *testing.T) {
	data, err := os.ReadFile("/sys/fs/cgroup/memory.max")
	if err != nil {
		t.Skip("no cgroup v2 memory limit")
	}
	limit, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil || limit == 0 || limit >= 2*gib {
		t.Skip("not running under a small finite cgroup v2 memory limit")
	}
	ram, cores, err := profileHost()
	if err != nil || ram == 0 || ram > limit || cores < 1 {
		t.Fatalf("profileHost: ram=%d limit=%d cores=%d err=%v", ram, limit, cores, err)
	}
	tuning := PlanTuning(config.Default())
	if tuning.MaxThreads != 1 || tuning.MaxFileSizeMB != 25 || tuning.MaxScanSizeMB != 25 {
		t.Fatalf("small cgroup must receive small input budgets: %+v", tuning)
	}
}
