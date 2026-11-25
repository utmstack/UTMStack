package utils_test

import (
	"errors"
	"testing"

	"github.com/utmstack/UTMStack/installer/utils"
)

func TestResolveMainIPFallbacks(t *testing.T) {
	fail := func() (string, error) { return "", errors.New("network unreachable") }
	okOnline := func() (string, error) { return "192.0.2.10", nil }
	okAirGap := func() (string, error) { return "198.51.100.20", nil }

	t.Run("online succeeds", func(t *testing.T) {
		ip, err := utils.ResolveMainIP(true, "203.0.113.1", okOnline, okAirGap)
		if err != nil || ip != "192.0.2.10" {
			t.Fatalf("got (%q, %v), want (192.0.2.10, nil)", ip, err)
		}
	})

	t.Run("online fails falls back to airgap interface", func(t *testing.T) {
		ip, err := utils.ResolveMainIP(true, "203.0.113.1", fail, okAirGap)
		if err != nil || ip != "198.51.100.20" {
			t.Fatalf("got (%q, %v), want (198.51.100.20, nil)", ip, err)
		}
	})

	t.Run("airgap boot race falls back to persisted MainServer", func(t *testing.T) {
		ip, err := utils.ResolveMainIP(false, "203.0.113.1", fail, fail)
		if err != nil || ip != "203.0.113.1" {
			t.Fatalf("got (%q, %v), want (203.0.113.1, nil)", ip, err)
		}
	})

	t.Run("all sources empty returns error without panicking", func(t *testing.T) {
		ip, err := utils.ResolveMainIP(false, "", fail, fail)
		if err == nil || ip != "" {
			t.Fatalf("got (%q, %v), want error", ip, err)
		}
	})
}
