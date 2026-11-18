//go:build darwin

package engine

import "testing"

func TestNativeDarwinHostProbe(t *testing.T) {
	ram, cores, err := profileHost()
	if err != nil || ram == 0 || cores < 1 {
		t.Fatalf("native host memory/core probe failed: RAM=%d cores=%d error=%v", ram, cores, err)
	}
}
