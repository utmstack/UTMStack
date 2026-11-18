package service

import (
	"context"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"testing"
)

func TestManagementSafetyIncludesEveryConfiguredPlane(t *testing.T) {
	c := config.Default()
	c.Server = "https://192.0.2.1:9000"
	c.SigMirror = "https://[2001:db8::1]/private/feed"
	c.Blocklist.MirrorBaseURL = "https://198.51.100.2:9001"
	addresses, err := ManagementAddresses(context.Background(), c)
	if err != nil || len(addresses) != 3 {
		t.Fatalf("%v %v", addresses, err)
	}
	c.Blocklist.MirrorBaseURL = "https://0.0.0.0"
	if _, err := ManagementAddresses(context.Background(), c); err == nil {
		t.Fatal("unsafe management endpoint accepted")
	}
}
