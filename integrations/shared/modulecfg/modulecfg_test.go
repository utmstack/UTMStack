package modulecfg

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/utmstack/UTMStack/integrations/shared/crypto"
)

const secret = "an-encryption-key"

func init() { retryDelay = 50 * time.Millisecond }

func startNATS(t *testing.T) string {
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

func bucket(t *testing.T, url string) jetstream.KeyValue {
	t.Helper()

	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(nc.Close)

	js, _ := jetstream.New(nc)

	kv, err := js.CreateOrUpdateKeyValue(context.Background(), jetstream.KeyValueConfig{Bucket: Bucket})
	if err != nil {
		t.Fatal(err)
	}

	return kv
}

func seal(t *testing.T, key, doc string) []byte {
	t.Helper()

	enc, err := crypto.NewCipher(key).Encrypt(doc)
	if err != nil {
		t.Fatal(err)
	}

	return []byte(enc)
}

// recorder collects what the watcher reports.
type recorder struct {
	mu     sync.Mutex
	values [][]byte
	errs   []error
	wake   chan struct{}
}

func newRecorder() *recorder { return &recorder{wake: make(chan struct{}, 64)} }

func (r *recorder) change(b []byte) {
	r.mu.Lock()
	r.values = append(r.values, b)
	r.mu.Unlock()
	r.wake <- struct{}{}
}

func (r *recorder) fail(err error) {
	r.mu.Lock()
	r.errs = append(r.errs, err)
	r.mu.Unlock()
	r.wake <- struct{}{}
}

func (r *recorder) snapshot() ([][]byte, []error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([][]byte(nil), r.values...), append([]error(nil), r.errs...)
}

func (r *recorder) waitFor(t *testing.T, n int) {
	t.Helper()

	deadline := time.After(10 * time.Second)

	for {
		r.mu.Lock()
		got := len(r.values) + len(r.errs)
		r.mu.Unlock()

		if got >= n {
			return
		}

		select {
		case <-r.wake:
		case <-deadline:
			t.Fatalf("waited for %d events, got %d", n, got)
		}
	}
}

func run(t *testing.T, url, key string, rec *recorder) context.CancelFunc {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	go func() { _ = Watch(ctx, url, key, crypto.NewCipher(secret), rec.change, rec.fail) }()

	t.Cleanup(cancel)

	return cancel
}

func TestAStartingContainerGetsTheCurrentConfiguration(t *testing.T) {
	url := startNATS(t)
	kv := bucket(t, url)

	if _, err := kv.Put(context.Background(), "aws", seal(t, secret, "plugins:\n  aws:\n    tenants: []\n")); err != nil {
		t.Fatal(err)
	}

	rec := newRecorder()
	run(t, url, "aws", rec)
	rec.waitFor(t, 1)

	if vals, _ := rec.snapshot(); string(vals[0]) != "plugins:\n  aws:\n    tenants: []\n" {
		t.Errorf("got %q", vals[0])
	}
}

func TestChangesArriveWithoutRestarting(t *testing.T) {
	url := startNATS(t)
	kv := bucket(t, url)

	rec := newRecorder()
	run(t, url, "azure", rec)

	rec.waitFor(t, 1)

	if vals, _ := rec.snapshot(); vals[0] != nil {
		t.Fatalf("a key that does not exist must read as no configuration, got %q", vals[0])
	}

	if _, err := kv.Put(context.Background(), "azure", seal(t, secret, "plugins:\n  azure: {}\n")); err != nil {
		t.Fatal(err)
	}

	rec.waitFor(t, 2)

	if vals, _ := rec.snapshot(); string(vals[1]) != "plugins:\n  azure: {}\n" {
		t.Errorf("got %q", vals[1])
	}

	if err := kv.Delete(context.Background(), "azure"); err != nil {
		t.Fatal(err)
	}

	rec.waitFor(t, 3)

	if vals, _ := rec.snapshot(); vals[2] != nil {
		t.Errorf("a removed key must read as no configuration, got %q", vals[2])
	}
}

func TestAWatcherOnlySeesItsOwnKey(t *testing.T) {
	url := startNATS(t)
	kv := bucket(t, url)

	rec := newRecorder()
	run(t, url, "gcp", rec)
	rec.waitFor(t, 1)

	if _, err := kv.Put(context.Background(), "sophos", seal(t, secret, "plugins:\n  sophos: {}\n")); err != nil {
		t.Fatal(err)
	}

	time.Sleep(300 * time.Millisecond)

	rec.mu.Lock()
	defer rec.mu.Unlock()

	if len(rec.values) != 1 {
		t.Errorf("another integration's configuration reached this one: %d values", len(rec.values))
	}
}

func TestAConfigurationWrittenWithAnotherKeyIsRejectedAndTheOldOneStays(t *testing.T) {
	url := startNATS(t)
	kv := bucket(t, url)

	if _, err := kv.Put(context.Background(), "o365", seal(t, secret, "plugins:\n  o365: {}\n")); err != nil {
		t.Fatal(err)
	}

	rec := newRecorder()
	run(t, url, "o365", rec)
	rec.waitFor(t, 1)

	if _, err := kv.Put(context.Background(), "o365", seal(t, "another-key", "plugins:\n  o365: {x: 1}\n")); err != nil {
		t.Fatal(err)
	}

	rec.waitFor(t, 2)

	rec.mu.Lock()
	defer rec.mu.Unlock()

	if len(rec.values) != 1 || len(rec.errs) != 1 {
		t.Errorf("values=%d errors=%d, want the old value kept and one error", len(rec.values), len(rec.errs))
	}
}

func TestItWaitsForTheBucketToExist(t *testing.T) {
	url := startNATS(t)

	rec := newRecorder()
	run(t, url, "crowdstrike", rec)

	rec.waitFor(t, 1) // the failed attempts are reported while the bucket is missing

	kv := bucket(t, url)
	if _, err := kv.Put(context.Background(), "crowdstrike", seal(t, secret, "plugins:\n  crowdstrike: {}\n")); err != nil {
		t.Fatal(err)
	}

	deadline := time.After(10 * time.Second)

	for {
		rec.mu.Lock()
		n := len(rec.values)
		rec.mu.Unlock()

		if n > 0 {
			return
		}

		select {
		case <-rec.wake:
		case <-deadline:
			t.Fatal("the configuration never arrived after the bucket was created")
		}
	}
}

func TestWithoutNATSItRefusesToStart(t *testing.T) {
	if err := Watch(context.Background(), "", "aws", crypto.NewCipher(secret), func([]byte) {}, func(error) {}); err == nil {
		t.Error("expected an error")
	}
}
