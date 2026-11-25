package repository

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

const testKey = "a-test-encryption-key"

func natsURL(t *testing.T) string {
	t.Helper()

	srv, err := server.NewServer(&server.Options{Port: -1, JetStream: true, StoreDir: t.TempDir(), NoLog: true, NoSigs: true})
	if err != nil {
		t.Fatal(err)
	}

	go srv.Start()

	if !srv.ReadyForConnections(10 * time.Second) {
		t.Fatal("NATS did not start")
	}

	// Accepting connections comes before JetStream answers API calls.
	for deadline := time.Now().Add(20 * time.Second); !srv.JetStreamEnabled(); {
		if time.Now().After(deadline) {
			t.Fatal("JetStream did not start")
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Cleanup(srv.Shutdown)

	return srv.ClientURL()
}

func natsStore(t *testing.T, url string) *NATSConfigStore {
	t.Helper()

	c, err := secret.NewCipher(testKey)
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewNATSConfigStore(context.Background(), url, c, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(s.Close)

	return s
}

func TestNATSStoreRoundTrip(t *testing.T) {
	s := natsStore(t, natsURL(t))
	ctx := ctxFor(tenantA)

	want := []domain.ConfigGroup{{Name: "Azure Prod", Description: "d", Config: map[string]string{"clientId": "abc", "clientSecret": "shh"}}}
	if err := s.Save(ctx, "AZURE", want); err != nil {
		t.Fatal(err)
	}

	got, err := s.Load(ctx, "AZURE")
	if err != nil || len(got) != 1 || got[0].Name != "Azure Prod" || got[0].Config["clientSecret"] != "shh" {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestNATSStoreKeepsTenantsApartAndSavingOneLeavesTheOthers(t *testing.T) {
	s := natsStore(t, natsURL(t))

	if err := s.Save(ctxFor(tenantA), "AZURE", groups("a1", "a2")); err != nil {
		t.Fatal(err)
	}

	if err := s.Save(ctxFor(tenantB), "AZURE", groups("b1")); err != nil {
		t.Fatal(err)
	}

	a, _ := s.Load(ctxFor(tenantA), "AZURE")
	b, _ := s.Load(ctxFor(tenantB), "AZURE")

	if len(a) != 2 || len(b) != 1 {
		t.Errorf("tenant A has %d groups and tenant B %d, want 2 and 1", len(a), len(b))
	}

	all, err := s.LoadAllTenants(context.Background(), "AZURE")
	if err != nil || len(all) != 2 {
		t.Errorf("all tenants: %d %v", len(all), err)
	}
}

func TestNATSStoreUpsertAndDelete(t *testing.T) {
	s := natsStore(t, natsURL(t))
	ctx := ctxFor(tenantA)

	for _, c := range []string{"one", "two"} {
		if err := s.Upsert(ctx, "GCP", domain.ConfigGroup{Name: "g", Config: map[string]string{"v": c}}); err != nil {
			t.Fatal(err)
		}
	}

	got, _ := s.Load(ctx, "GCP")
	if len(got) != 1 || got[0].Config["v"] != "two" {
		t.Fatalf("upserting the same name must replace it: %+v", got)
	}

	if err := s.Delete(ctx, "GCP", "g"); err != nil {
		t.Fatal(err)
	}

	if got, _ := s.Load(ctx, "GCP"); len(got) != 0 {
		t.Errorf("group still there: %+v", got)
	}

	if err := s.Delete(ctx, "GCP", "g"); err != domain.ErrConfigGroupNotFound {
		t.Errorf("deleting a missing group: %v", err)
	}
}

func TestNATSStoreNeverHoldsAReadableConfiguration(t *testing.T) {
	url := natsURL(t)
	s := natsStore(t, url)

	if err := s.Save(ctxFor(tenantA), "AWS_IAM_USER", []domain.ConfigGroup{{Name: "prod", Config: map[string]string{"aws_secret_access_key": "TOPSECRET", "region": "us-east-1"}}}); err != nil {
		t.Fatal(err)
	}

	nc, _ := nats.Connect(url)
	defer nc.Close()

	js, _ := jetstream.New(nc)
	kv, err := js.KeyValue(context.Background(), ConfigBucket)
	if err != nil {
		t.Fatal(err)
	}

	e, err := kv.Get(context.Background(), "aws")
	if err != nil {
		t.Fatal(err)
	}

	raw := string(e.Value())
	for _, leak := range []string{"TOPSECRET", "us-east-1", "plugins", "prod", tenantA} {
		if strings.Contains(raw, leak) {
			t.Errorf("the stored value shows %q", leak)
		}
	}
}

func TestNATSStoreReadingWithAnotherKeyFailsInsteadOfReturningGarbage(t *testing.T) {
	url := natsURL(t)
	s := natsStore(t, url)

	if err := s.Save(ctxFor(tenantA), "AZURE", groups("a")); err != nil {
		t.Fatal(err)
	}

	other, _ := secret.NewCipher("another-key")

	s2, err := NewNATSConfigStore(context.Background(), url, other, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()

	if _, err := s2.Load(ctxFor(tenantA), "AZURE"); err == nil {
		t.Error("a wrong key must be an error")
	}

	// And it must not overwrite what it could not read.
	if err := s2.Upsert(ctxFor(tenantA), "AZURE", domain.ConfigGroup{Name: "x"}); err == nil {
		t.Error("a write that could not read the current value must fail")
	}
}

func TestNATSStoreConcurrentWritesDoNotLoseEachOther(t *testing.T) {
	url := natsURL(t)
	a, b := natsStore(t, url), natsStore(t, url) // two replicas of the backend

	var wg sync.WaitGroup

	for i, s := range []*NATSConfigStore{a, b} {
		for n := 0; n < 5; n++ {
			wg.Add(1)

			go func(s *NATSConfigStore, name string) {
				defer wg.Done()

				if err := s.Upsert(ctxFor(tenantA), "SOPHOS", domain.ConfigGroup{Name: name, Config: map[string]string{"k": "v"}}); err != nil {
					t.Error(err)
				}
			}(s, string(rune('a'+i))+string(rune('0'+n)))
		}
	}

	wg.Wait()

	got, err := a.Load(ctxFor(tenantA), "SOPHOS")
	if err != nil || len(got) != 10 {
		t.Fatalf("%d groups survived of 10 (%v)", len(got), err)
	}
}

const legacyFile = `plugins:
    azure:
        tenants:
            - id: 8f1c1b8e-0000-4000-8000-00000000000a
              groups:
                - name: legacy
                  config:
                    clientId: from-the-file
`

func writeLegacy(t *testing.T, dir, key, body string) string {
	t.Helper()

	p := filepath.Join(dir, "system_plugins_"+key+".yaml")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	return p
}

func TestMigrationMovesALeftoverFileIntoNATS(t *testing.T) {
	s := natsStore(t, natsURL(t))
	dir := t.TempDir()
	path := writeLegacy(t, dir, "azure", legacyFile)

	n, err := s.MigrateFiles(context.Background(), dir)
	if err != nil || n != 1 {
		t.Fatalf("imported %d, err %v", n, err)
	}

	got, err := s.Load(ctxFor(tenantA), "AZURE")
	if err != nil || len(got) != 1 || got[0].Name != "legacy" || got[0].Config["clientId"] != "from-the-file" {
		t.Fatalf("after migration: %+v %v", got, err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("the file must be set aside")
	}

	if _, err := os.Stat(path + ".migrated"); err != nil {
		t.Error("the file must be kept as .migrated")
	}
}

func TestMigrationIsIdempotent(t *testing.T) {
	s := natsStore(t, natsURL(t))
	dir := t.TempDir()
	writeLegacy(t, dir, "azure", legacyFile)

	if _, err := s.MigrateFiles(context.Background(), dir); err != nil {
		t.Fatal(err)
	}

	if n, err := s.MigrateFiles(context.Background(), dir); err != nil || n != 0 {
		t.Errorf("second run imported %d, err %v", n, err)
	}
}

func TestMigrationNeverOverwritesWhatNATSAlreadyHolds(t *testing.T) {
	s := natsStore(t, natsURL(t))
	dir := t.TempDir()
	path := writeLegacy(t, dir, "azure", legacyFile)

	if err := s.Save(ctxFor(tenantA), "AZURE", []domain.ConfigGroup{{Name: "newer", Config: map[string]string{"clientId": "from-nats"}}}); err != nil {
		t.Fatal(err)
	}

	n, err := s.MigrateFiles(context.Background(), dir)
	if err != nil || n != 0 {
		t.Fatalf("imported %d, err %v", n, err)
	}

	got, _ := s.Load(ctxFor(tenantA), "AZURE")
	if len(got) != 1 || got[0].Name != "newer" {
		t.Errorf("NATS value was overwritten: %+v", got)
	}

	if _, err := os.Stat(path + ".migrated"); err != nil {
		t.Error("the stale file must still be set aside")
	}
}

func TestMigrationKeepsAFileItCannotRead(t *testing.T) {
	s := natsStore(t, natsURL(t))
	dir := t.TempDir()
	bad := writeLegacy(t, dir, "gcp", "plugins: [not: valid")
	good := writeLegacy(t, dir, "azure", legacyFile)

	n, err := s.MigrateFiles(context.Background(), dir)
	if err == nil {
		t.Error("the broken file must be reported")
	}

	if n != 1 {
		t.Errorf("the good file must still be imported, imported %d", n)
	}

	if _, err := os.Stat(bad); err != nil {
		t.Error("the broken file must stay where it was")
	}

	if _, err := os.Stat(good + ".migrated"); err != nil {
		t.Error("the good file must be set aside")
	}
}

func TestMigrationWithNoFilesDoesNothing(t *testing.T) {
	s := natsStore(t, natsURL(t))

	if n, err := s.MigrateFiles(context.Background(), t.TempDir()); err != nil || n != 0 {
		t.Errorf("imported %d, err %v", n, err)
	}
}
