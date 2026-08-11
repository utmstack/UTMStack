package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/storage/domain"
)

type fakeStore struct {
	retentions map[domain.Dataset]domain.Retention
	ready      bool
	reloaded   bool
	reloadErr  error
	adopted    []domain.Retention
	flat       []domain.Retention
}

func newFakeStore() *fakeStore {
	return &fakeStore{retentions: map[domain.Dataset]domain.Retention{
		domain.DatasetLogs:   {Dataset: domain.DatasetLogs, KeepDays: 730},
		domain.DatasetAlerts: {Dataset: domain.DatasetAlerts, KeepDays: 730},
		domain.DatasetStats:  {Dataset: domain.DatasetStats, KeepDays: 1095},
	}}
}

func (f *fakeStore) Retention(_ context.Context, d domain.Dataset) (domain.Retention, error) {
	return f.retentions[d], nil
}

func (f *fakeStore) SetRetention(_ context.Context, r domain.Retention) error {
	if f.retentions[r.Dataset].Tiered() && !r.Tiered() {
		return domain.ErrTieringPermanent
	}
	f.flat = append(f.flat, r)
	f.retentions[r.Dataset] = r
	return nil
}

func (f *fakeStore) AdoptTiering(_ context.Context, r domain.Retention) error {
	if !f.ready {
		return errors.New("the volume does not exist")
	}
	f.adopted = append(f.adopted, r)
	f.retentions[r.Dataset] = r
	return nil
}

func (f *fakeStore) Usage(context.Context) ([]domain.Usage, error)  { return nil, nil }
func (f *fakeStore) Health(context.Context) (domain.Health, error)  { return domain.Health{}, nil }
func (f *fakeStore) ColdStorageReady(context.Context) (bool, error) { return f.ready, nil }
func (f *fakeStore) ReloadConfig(context.Context) error {
	f.reloaded = true
	if f.reloadErr != nil {
		return f.reloadErr
	}
	f.ready = true
	return nil
}

type fakeConfig struct {
	tiering domain.Tiering
	written []domain.ObjectStore
	undone  bool
}

func (f *fakeConfig) Read() (domain.Tiering, error) { return f.tiering, nil }
func (f *fakeConfig) Write(o domain.ObjectStore) (func() error, error) {
	previous := f.tiering
	f.written = append(f.written, o)
	f.tiering = domain.Tiering{Configured: true, Endpoint: o.Endpoint}
	return func() error { f.tiering = previous; f.undone = true; return nil }, nil
}

func newService() (*service, *fakeStore, *fakeConfig) {
	st, cfg := newFakeStore(), &fakeConfig{}
	return &service{store: st, config: cfg}, st, cfg
}

// A cold retention on an instance with no object storage would be a TTL naming
// a volume that does not exist, which fails at the table rather than at the
// person who typed it.
func TestMovingToColdStorageNeedsColdStorage(t *testing.T) {
	s, _, _ := newService()

	_, err := s.SetRetention(context.Background(), domain.Retention{
		Dataset: domain.DatasetLogs, KeepDays: 365, ColdDays: 30,
	})
	if !errors.Is(err, domain.ErrTieringRequired) {
		t.Fatalf("err = %v, want cold storage required", err)
	}
}

func TestATieredRetentionAdoptsThePolicy(t *testing.T) {
	s, st, _ := newService()
	st.ready = true

	got, err := s.SetRetention(context.Background(), domain.Retention{
		Dataset: domain.DatasetLogs, KeepDays: 365, ColdDays: 30,
	})
	if err != nil {
		t.Fatalf("SetRetention: %v", err)
	}
	if len(st.adopted) != 1 {
		t.Fatalf("adopted %d times, want the policy adopted once", len(st.adopted))
	}
	if len(st.flat) != 0 {
		t.Error("a tiered retention went through the flat path")
	}
	if got.ColdDays != 30 || got.KeepDays != 365 {
		t.Errorf("got %+v, want what the table ended up with", got)
	}
}

// Rewriting a tiered TTL as a flat one drops the move clause silently, so the
// store refuses and the refusal has to survive up here.
func TestTurningColdStorageOffIsRefused(t *testing.T) {
	s, st, _ := newService()
	st.ready = true
	st.retentions[domain.DatasetLogs] = domain.Retention{
		Dataset: domain.DatasetLogs, KeepDays: 365, ColdDays: 30,
	}

	_, err := s.SetRetention(context.Background(), domain.Retention{
		Dataset: domain.DatasetLogs, KeepDays: 365,
	})
	if !errors.Is(err, domain.ErrTieringPermanent) {
		t.Fatalf("err = %v, want the tiering refused", err)
	}
}

func TestARetentionThatMakesNoSenseNeverReachesTheStore(t *testing.T) {
	s, st, _ := newService()
	st.ready = true

	_, err := s.SetRetention(context.Background(), domain.Retention{
		Dataset: domain.DatasetLogs, KeepDays: 30, ColdDays: 90,
	})
	if err == nil {
		t.Fatal("it was accepted")
	}
	if len(st.adopted)+len(st.flat) != 0 {
		t.Error("the store was asked anyway")
	}
}

// Enabling writes the file and then makes the store read it: a file nobody
// reloaded is a bucket that does not exist yet.
func TestEnablingWritesThenReloads(t *testing.T) {
	s, st, cfg := newService()

	got, err := s.EnableTiering(context.Background(), domain.ObjectStore{
		Endpoint:  "https://s3.example.com/utmstack-cold",
		AccessKey: "AKIA",
		SecretKey: "s3cr3t",
	})
	if err != nil {
		t.Fatalf("EnableTiering: %v", err)
	}
	if len(cfg.written) != 1 {
		t.Fatalf("wrote %d files", len(cfg.written))
	}
	if !st.reloaded {
		t.Error("the store was never told to read its configuration again")
	}
	if !got.Ready || !got.Configured {
		t.Errorf("got %+v, want it configured and in use", got)
	}
	if got.Endpoint != "https://s3.example.com/utmstack-cold/" {
		t.Errorf("endpoint = %q, want it normalized", got.Endpoint)
	}
}

// Parts already moved carry the bucket they went to. Pointing the store
// somewhere else leaves them unreadable, and nothing says so until a query
// reaches that far back.
func TestTheBucketCannotChangeOnceRecordsLiveInIt(t *testing.T) {
	s, st, cfg := newService()
	st.ready = true
	cfg.tiering = domain.Tiering{Configured: true, Endpoint: "https://s3.example.com/old/"}
	st.retentions[domain.DatasetLogs] = domain.Retention{
		Dataset: domain.DatasetLogs, KeepDays: 365, ColdDays: 30,
	}

	_, err := s.EnableTiering(context.Background(), domain.ObjectStore{
		Endpoint:  "https://s3.example.com/new/",
		AccessKey: "AKIA",
		SecretKey: "s3cr3t",
	})
	if !errors.Is(err, domain.ErrEndpointLocked) {
		t.Fatalf("err = %v, want the bucket locked", err)
	}
	if len(cfg.written) != 0 {
		t.Error("it wrote the new bucket anyway")
	}
}

// Rotating the keys is the same bucket, and has to keep working.
func TestCredentialsCanBeRotated(t *testing.T) {
	s, st, cfg := newService()
	st.ready = true
	cfg.tiering = domain.Tiering{Configured: true, Endpoint: "https://s3.example.com/cold/"}
	st.retentions[domain.DatasetLogs] = domain.Retention{
		Dataset: domain.DatasetLogs, KeepDays: 365, ColdDays: 30,
	}

	if _, err := s.EnableTiering(context.Background(), domain.ObjectStore{
		Endpoint:  "https://s3.example.com/cold/",
		AccessKey: "AKIA-NEW",
		SecretKey: "rotated",
	}); err != nil {
		t.Fatalf("rotation refused: %v", err)
	}
	if len(cfg.written) != 1 {
		t.Error("the new credentials were not written")
	}
}

// Tiering answers what is true, not what was asked for: a file naming a bucket
// the store has not picked up is not cold storage yet.
func TestAWrittenBucketIsNotColdStorageUntilTheStoreHasIt(t *testing.T) {
	s, _, cfg := newService()
	cfg.tiering = domain.Tiering{Configured: true, Endpoint: "https://s3.example.com/cold/"}

	got, err := s.Tiering(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !got.Configured {
		t.Error("the configured bucket was hidden")
	}
	if got.Ready {
		t.Error("it reported cold storage the store does not have")
	}
}

// The store writes a probe object into the bucket while reading the file, so a
// wrong endpoint or key is refused right here. The file has to come back out:
// left behind, it is what the server would load on its next start.
func TestABucketTheStoreRefusesIsTakenBackOut(t *testing.T) {
	s, st, cfg := newService()
	st.reloadErr = errors.New("Not found address of host: minio.internal")

	_, err := s.EnableTiering(context.Background(), domain.ObjectStore{
		Endpoint:  "http://minio.internal:9000/utmstack-cold/",
		AccessKey: "AKIA",
		SecretKey: "s3cr3t",
	})
	if !errors.Is(err, domain.ErrColdRefused) {
		t.Fatalf("err = %v, want the bucket refused", err)
	}
	if !cfg.undone {
		t.Error("the rejected configuration was left on disk")
	}
	if !strings.Contains(err.Error(), "minio.internal") {
		t.Errorf("err = %v, want it to carry what the store said", err)
	}
}
