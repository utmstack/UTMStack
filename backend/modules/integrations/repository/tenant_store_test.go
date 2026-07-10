package repository

import (
	"os"
	"strings"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
)

func TestTenantStoreRoundtrip(t *testing.T) {
	dir := t.TempDir()
	s := NewTenantStore(dir)

	want := []domain.Tenant{
		{Name: "t1", Config: map[string]string{"access_key": "AK", "secret": "SK"}},
		{Name: "t2", Config: map[string]string{"access_key": "AK2"}},
	}
	if err := s.Save("AWS_IAM_USER", want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := s.Load("AWS_IAM_USER")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Name != want[i].Name {
			t.Fatalf("[%d] name got %q want %q", i, got[i].Name, want[i].Name)
		}
		for k, v := range want[i].Config {
			if got[i].Config[k] != v {
				t.Fatalf("[%d] %s got %q want %q", i, k, got[i].Config[k], v)
			}
		}
	}

	data, err := os.ReadFile(s.path("AWS_IAM_USER"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	txt := string(data)
	if !strings.Contains(txt, "plugins:") || !strings.Contains(txt, "aws:") || !strings.Contains(txt, "tenants:") {
		t.Fatalf("unexpected file layout:\n%s", txt)
	}
}
