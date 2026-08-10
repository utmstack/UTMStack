package repository

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	tenantA = "8f1c1b8e-0000-4000-8000-00000000000a"
	tenantB = "8f1c1b8e-0000-4000-8000-00000000000b"
)

func ctxFor(tenant string) context.Context {
	return authz.WithTenantID(context.Background(), tenant)
}

func groups(names ...string) []domain.ConfigGroup {
	out := make([]domain.ConfigGroup, 0, len(names))
	for _, n := range names {
		out = append(out, domain.ConfigGroup{Name: n, Config: map[string]string{"clientId": n}})
	}
	return out
}

func TestAConfigGroupSurvivesTheRoundTrip(t *testing.T) {
	s := NewConfigStore(t.TempDir())
	ctx := ctxFor(tenantA)

	want := []domain.ConfigGroup{{
		Name:        "Azure Prod",
		Description: "production subscription",
		Config:      map[string]string{"clientId": "abc", "clientSecret": "shh"},
	}}
	if err := s.Save(ctx, "AZURE", want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Load(ctx, "AZURE")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("loaded %d groups, want 1", len(got))
	}
	if got[0].Name != "Azure Prod" || got[0].Description != "production subscription" {
		t.Errorf("group came back as %+v", got[0])
	}
	if got[0].Config["clientSecret"] != "shh" {
		t.Errorf("credential lost: %v", got[0].Config)
	}
}

// These files hold customer credentials. One tenant reading another's would be
// the whole point of multitenancy failing at the last step.
func TestATenantCannotSeeAnothersGroups(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Save(ctxFor(tenantA), "AZURE", groups("only-a")); err != nil {
		t.Fatal(err)
	}

	got, err := s.Load(ctxFor(tenantB), "AZURE")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("tenant B sees %d of tenant A's groups", len(got))
	}
}

// One file holds every tenant's section, so a write for one must not be a
// rewrite of the file that drops the others.
func TestSavingForOneTenantLeavesTheOthersAlone(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Save(ctxFor(tenantA), "AZURE", groups("a1", "a2")); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(ctxFor(tenantB), "AZURE", groups("b1")); err != nil {
		t.Fatal(err)
	}

	a, err := s.Load(ctxFor(tenantA), "AZURE")
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 2 {
		t.Errorf("tenant A has %d groups after tenant B saved, want 2", len(a))
	}
}

// The name only has to be unique inside a tenant — "production" is what
// everybody calls theirs.
func TestTwoTenantsMayUseTheSameGroupName(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Upsert(ctxFor(tenantA), "AZURE", domain.ConfigGroup{
		Name: "production", Config: map[string]string{"clientId": "for-a"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Upsert(ctxFor(tenantB), "AZURE", domain.ConfigGroup{
		Name: "production", Config: map[string]string{"clientId": "for-b"}}); err != nil {
		t.Fatal(err)
	}

	a, _ := s.Load(ctxFor(tenantA), "AZURE")
	b, _ := s.Load(ctxFor(tenantB), "AZURE")
	if len(a) != 1 || len(b) != 1 {
		t.Fatalf("groups: A=%d B=%d, want 1 each", len(a), len(b))
	}
	if a[0].Config["clientId"] != "for-a" || b[0].Config["clientId"] != "for-b" {
		t.Errorf("one tenant overwrote the other: A=%q B=%q",
			a[0].Config["clientId"], b[0].Config["clientId"])
	}
}

func TestUpsertReplacesTheGroupWithTheSameName(t *testing.T) {
	s := NewConfigStore(t.TempDir())
	ctx := ctxFor(tenantA)

	for _, secret := range []string{"old", "new"} {
		if err := s.Upsert(ctx, "AZURE", domain.ConfigGroup{
			Name: "prod", Config: map[string]string{"clientSecret": secret}}); err != nil {
			t.Fatal(err)
		}
	}

	got, _ := s.Load(ctx, "AZURE")
	if len(got) != 1 {
		t.Fatalf("upsert created %d groups, want 1", len(got))
	}
	if got[0].Config["clientSecret"] != "new" {
		t.Errorf("stored secret is %q, want the updated one", got[0].Config["clientSecret"])
	}
}

// A request with no tenant must not be served from, or written into, the
// section that belongs to nobody — that is how credentials leak between
// customers or vanish into an unreachable entry.
func TestWithoutATenantEveryOperationIsRefused(t *testing.T) {
	s := NewConfigStore(t.TempDir())
	ctx := context.Background()

	if _, err := s.Load(ctx, "AZURE"); !errors.Is(err, tenancy.ErrNoTenant) {
		t.Errorf("Load returned %v, want ErrNoTenant", err)
	}
	if err := s.Save(ctx, "AZURE", groups("x")); !errors.Is(err, tenancy.ErrNoTenant) {
		t.Errorf("Save returned %v, want ErrNoTenant", err)
	}
	if err := s.Upsert(ctx, "AZURE", domain.ConfigGroup{Name: "x"}); !errors.Is(err, tenancy.ErrNoTenant) {
		t.Errorf("Upsert returned %v, want ErrNoTenant", err)
	}
	if err := s.Delete(ctx, "AZURE", "x"); !errors.Is(err, tenancy.ErrNoTenant) {
		t.Errorf("Delete returned %v, want ErrNoTenant", err)
	}
}

func TestDeletingAGroupThatIsNotThereSaysSo(t *testing.T) {
	s := NewConfigStore(t.TempDir())
	ctx := ctxFor(tenantA)
	if err := s.Save(ctx, "AZURE", groups("kept")); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(ctx, "AZURE", "never-existed"); !errors.Is(err, domain.ErrConfigGroupNotFound) {
		t.Errorf("Delete returned %v, want ErrConfigGroupNotFound", err)
	}

	got, _ := s.Load(ctx, "AZURE")
	if len(got) != 1 {
		t.Errorf("a failed delete changed the file: %d groups left", len(got))
	}
}

func TestDeletingTheLastGroupRemovesTheTenantSection(t *testing.T) {
	s := NewConfigStore(t.TempDir())

	if err := s.Save(ctxFor(tenantA), "AZURE", groups("only")); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(ctxFor(tenantB), "AZURE", groups("b1")); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctxFor(tenantA), "AZURE", "only"); err != nil {
		t.Fatal(err)
	}

	all, err := s.LoadAllTenants(context.Background(), "AZURE")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Fatalf("file holds %d tenant sections, want only the one with groups", len(all))
	}
	if len(all[0].Groups) != 1 || all[0].Groups[0].Name != "b1" {
		t.Errorf("the surviving section is %+v", all[0])
	}
}

// The datasource sweep belongs to no tenant's request and must see everyone.
func TestLoadAllTenantsSeesEverySection(t *testing.T) {
	s := NewConfigStore(t.TempDir())
	if err := s.Save(ctxFor(tenantA), "AZURE", groups("a1", "a2")); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(ctxFor(tenantB), "AZURE", groups("b1")); err != nil {
		t.Fatal(err)
	}

	all, err := s.LoadAllTenants(context.Background(), "AZURE")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("sweep sees %d tenants, want 2", len(all))
	}
	total := 0
	for _, tc := range all {
		total += len(tc.Groups)
	}
	if total != 3 {
		t.Errorf("sweep sees %d groups, want 3", total)
	}
}

// The puller plugins find the file by name; AWS_IAM_USER is the one that does
// not simply lowercase.
func TestTheFileIsNamedForThePluginThatReadsIt(t *testing.T) {
	dir := t.TempDir()
	s := NewConfigStore(dir)

	if err := s.Save(ctxFor(tenantA), "AWS_IAM_USER", groups("x")); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(dir + "/system_plugins_aws.yaml"); err != nil {
		t.Errorf("the aws plugin will not find its config: %v", err)
	}
}

func TestWithFileLockWaits(t *testing.T) {
	dir := t.TempDir()
	target := dir + "/f.yaml"

	// pre-hold the flock via a peer fd
	peer, err := os.OpenFile(target+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := syscall.Flock(int(peer.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- withFileLock(target, func() error { return nil })
	}()

	select {
	case <-done:
		t.Fatal("withFileLock returned while flock was held")
	case <-time.After(120 * time.Millisecond):
	}

	if err := syscall.Flock(int(peer.Fd()), syscall.LOCK_UN); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	_ = peer.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("withFileLock: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("withFileLock did not return after flock released")
	}
}

// TestWithFileLockKernelReleasesOnClose proves the crash-recovery property: if
// the peer holding the flock closes its fd (equivalent to process death), the
// kernel releases the lock and a new acquirer proceeds without any TTL/steal.
func TestWithFileLockKernelReleasesOnClose(t *testing.T) {
	dir := t.TempDir()
	target := dir + "/f.yaml"

	peer, err := os.OpenFile(target+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := syscall.Flock(int(peer.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock: %v", err)
	}
	// simulate crash: close fd without explicit unlock
	_ = peer.Close()

	start := time.Now()
	if err := withFileLock(target, func() error { return nil }); err != nil {
		t.Fatalf("withFileLock: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("lock not released on close: %v", elapsed)
	}
}
