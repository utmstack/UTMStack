package socai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/utmstack/utmstack/backend/modules/socai/repository"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

func natsForTest(t *testing.T) string {
	t.Helper()

	srv, err := server.NewServer(&server.Options{Port: -1, JetStream: true, StoreDir: t.TempDir(), NoLog: true, NoSigs: true})
	if err != nil {
		t.Fatal(err)
	}

	go srv.Start()

	if !srv.ReadyForConnections(10 * time.Second) {
		t.Fatal("NATS did not start")
	}

	t.Cleanup(srv.Shutdown)

	for deadline := time.Now().Add(20 * time.Second); !srv.JetStreamEnabled(); {
		if time.Now().After(deadline) {
			t.Fatal("JetStream did not start")
		}

		time.Sleep(50 * time.Millisecond)
	}

	return srv.ClientURL()
}

type reader struct {
	kv     jetstream.KeyValue
	cipher *secret.Cipher
}

func readerFor(t *testing.T, url string, c *secret.Cipher) reader {
	t.Helper()

	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(nc.Close)

	js, _ := jetstream.New(nc)

	kv, err := js.KeyValue(context.Background(), ConfigBucket)
	if err != nil {
		t.Fatal(err)
	}

	return reader{kv: kv, cipher: c}
}

func (r reader) raw(t *testing.T) (string, bool) {
	t.Helper()

	e, err := r.kv.Get(context.Background(), ConfigKey)
	if err != nil {
		return "", false
	}

	return string(e.Value()), true
}

func (r reader) snapshot(t *testing.T) repository.Snapshot {
	t.Helper()

	raw, ok := r.raw(t)
	if !ok {
		t.Fatal("no configuration in NATS")
	}

	plain, err := r.cipher.Decrypt(raw)
	if err != nil {
		t.Fatal(err)
	}

	var s repository.Snapshot
	if err := json.Unmarshal([]byte(plain), &s); err != nil {
		t.Fatalf("what NATS holds does not decrypt to the configuration: %v (%q)", err, plain[:min(40, len(plain))])
	}

	return s
}

func setup(t *testing.T) (*ConfigMirror, *repository.ConfigStore, reader) {
	t.Helper()

	url := natsForTest(t)
	c, err := secret.NewCipher("a-test-key")
	if err != nil {
		t.Fatal(err)
	}

	store := repository.NewConfigStore(t.TempDir())

	m, err := NewConfigMirror(context.Background(), url, store, c, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(m.nc.Close)

	return m, store, readerFor(t, url, c)
}

func TestTheConfigurationIsPublishedEncryptedAndWholeWithEachTenantsOwn(t *testing.T) {
	m, store, r := setup(t)

	if err := store.Save("", &repository.FileConfig{Provider: "openai", Model: "gpt", URL: "https://x/v1/chat/completions",
		APIKey: "ENC-KEY", AutoAnalyze: true}); err != nil {
		t.Fatal(err)
	}

	if err := store.Save("t1", &repository.FileConfig{Provider: "anthropic", Model: "claude", URL: "https://a/v1/messages",
		CustomHeaders: map[string]string{"id": "ENC-ID"}, AllowIncidents: true}); err != nil {
		t.Fatal(err)
	}

	if err := m.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}

	raw, _ := r.raw(t)
	for _, leak := range []string{"openai", "claude", "ENC-KEY", "provider", "t1"} {
		if strings.Contains(raw, leak) {
			t.Errorf("the value in NATS shows %q", leak)
		}
	}

	s := r.snapshot(t)
	if s.Default == nil || s.Default.Model != "gpt" || s.Default.APIKey != "ENC-KEY" || !s.Default.AutoAnalyze {
		t.Errorf("default %+v", s.Default)
	}

	if own := s.Tenants["t1"]; own.Provider != "anthropic" || own.CustomHeaders["id"] != "ENC-ID" || !own.AllowIncidents {
		t.Errorf("tenant t1 %+v", own)
	}
}

func TestEveryChangeInTheStoreReachesNATSWithoutAnyoneAsking(t *testing.T) {
	m, store, r := setup(t)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go m.Run(ctx)

	wait := func(what string, ok func(repository.Snapshot) bool) {
		t.Helper()

		for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(50 * time.Millisecond) {
			if raw, found := r.raw(t); found {
				plain, _ := r.cipher.Decrypt(raw)

				var s repository.Snapshot
				if json.Unmarshal([]byte(plain), &s) == nil && ok(s) {
					return
				}
			}
		}

		t.Fatalf("NATS never showed: %s", what)
	}

	if err := store.Save("", &repository.FileConfig{Provider: "openai", Model: "m1", URL: "u"}); err != nil {
		t.Fatal(err)
	}

	wait("the default", func(s repository.Snapshot) bool { return s.Default != nil && s.Default.Model == "m1" })

	if err := store.Save("t9", &repository.FileConfig{Provider: "openai", Model: "own", URL: "u"}); err != nil {
		t.Fatal(err)
	}

	wait("the tenant's own", func(s repository.Snapshot) bool { return s.Tenants["t9"].Model == "own" })

	if err := store.Delete("t9"); err != nil {
		t.Fatal(err)
	}

	wait("the tenant going back to the default", func(s repository.Snapshot) bool { _, has := s.Tenants["t9"]; return !has && s.Default != nil })
}

func TestWithNothingConfiguredThereIsNothingInNATS(t *testing.T) {
	m, _, r := setup(t)

	if err := m.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, found := r.raw(t); found {
		t.Error("a key with nothing in it would read as a configuration")
	}
}

func TestAKeyThatNoLongerHasAConfigurationIsRemoved(t *testing.T) {
	m, store, r := setup(t)

	if err := store.Save("", &repository.FileConfig{Provider: "openai", Model: "m", URL: "u"}); err != nil {
		t.Fatal(err)
	}

	if err := m.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, found := r.raw(t); !found {
		t.Fatal("it should be there first")
	}

	// The instance default cannot be deleted through the API; an operator removing the file by hand can.
	if err := store.Save("", &repository.FileConfig{}); err != nil {
		t.Fatal(err)
	}

	if err := m.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, found := r.raw(t); found {
		t.Error("an empty configuration must not stay published")
	}
}
