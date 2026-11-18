package service

import (
	"reflect"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestPermissionBackgroundAdmissionBudget(t *testing.T) {
	for _, tc := range []struct {
		name, platform, mode string
		configured, threads  int
		enabled, watcher     bool
		want, status         int
	}{
		{"open reserves one", "linux", "permission_open", 2, 2, true, true, 1, 1},
		{"exec reserves one", "linux", "permission_exec", 4, 4, true, true, 3, 3},
		{"derived concurrency", "linux", "permission_open", 0, 2, true, true, 1, 1},
		{"explicit lower setting retained", "linux", "permission_open", 1, 4, true, true, 1, 1},
		{"hard ceiling retained", "linux", "permission_open", 16, 32, true, true, 15, 15},
		{"single engine slot", "linux", "permission_open", 1, 1, true, true, 1, 0},
		{"notification only", "linux", "notify", 2, 2, true, true, 2, 0},
		{"permission disabled", "linux", "permission_open", 2, 2, false, true, 2, 0},
		{"watcher disabled", "linux", "permission_open", 2, 2, true, false, 2, 0},
		{"windows unchanged", "windows", "permission_open", 2, 2, true, true, 2, 0},
		{"darwin unchanged", "darwin", "permission_open", 2, 2, true, true, 2, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.Enabled = true
			cfg.ScanConcurrency = tc.configured
			cfg.LinuxProtection.Fanotify = config.LinuxFanotify{Enabled: tc.enabled, Mode: tc.mode, FailClosed: true, DeadlineMS: 1500, Workers: 2, MaxPending: 32}
			watcher := tc.watcher
			cfg.Sensors.FileWatcher = &watcher
			got, status := backgroundScanConfigForPlatform(cfg, tc.threads, tc.platform)
			if got.ScanConcurrency != tc.want || status != tc.status {
				t.Fatalf("background=%d status=%d; want %d/%d", got.ScanConcurrency, status, tc.want, tc.status)
			}
			got.ScanConcurrency = cfg.ScanConcurrency
			if !reflect.DeepEqual(got, cfg) {
				t.Fatal("background reservation changed another policy field")
			}
			if cfg.LinuxProtection.Fanotify.DeadlineMS != 1500 || !cfg.LinuxProtection.Fanotify.FailClosed || cfg.ScanConcurrency != tc.configured {
				t.Fatal("reservation relaxed deadline/fail policy or changed owner configuration")
			}
		})
	}
}
