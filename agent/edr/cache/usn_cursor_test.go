package cache

import (
	"path/filepath"
	"testing"
)

func TestUSNCursorRoundTrip(t *testing.T) {
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	c.SaveUSN("C:", 999, 12345)
	jid, next := c.LoadUSN("C:")
	if jid != 999 || next != 12345 {
		t.Fatalf("LoadUSN = %d,%d", jid, next)
	}
	// unknown volume returns zeros
	if j, n := c.LoadUSN("D:"); j != 0 || n != 0 {
		t.Fatalf("unknown volume = %d,%d", j, n)
	}
}
