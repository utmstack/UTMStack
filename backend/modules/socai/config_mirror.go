package socai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/socai/repository"
)

const (
	// ConfigBucket and ConfigKey are where the SOC AI service reads the configuration: one document
	// with the instance default and the configuration of every tenant that has its own.
	ConfigBucket = "soc-ai-config"
	ConfigKey    = "config"

	// resync repairs the bucket if it was ever lost or changed by hand. The configuration itself lives in
	// the store; what is in NATS is a copy of it.
	resync = 5 * time.Minute
)

type encrypter interface {
	Encrypt(string) (string, error)
}

// ConfigMirror keeps the SOC AI service's copy of the configuration in NATS up to date. The settings
// page still writes to the backend's store; the mirror publishes the whole document, encrypted, each
// time it changes and now and then in between, so a service that starts later or a bucket that was lost
// ends up with the current configuration.
type ConfigMirror struct {
	store  *repository.ConfigStore
	cipher encrypter
	nc     *nats.Conn
	kv     jetstream.KeyValue
	wake   chan struct{}
}

// NewConfigMirror connects to NATS and creates the bucket if it does not exist. NATS may still be
// starting with the backend, so it waits up to wait for it.
func NewConfigMirror(ctx context.Context, url string, store *repository.ConfigStore, cipher encrypter, wait time.Duration) (*ConfigMirror, error) {
	nc, err := nats.Connect(url, nats.MaxReconnects(-1), nats.RetryOnFailedConnect(true), nats.Name("backend-socai-config"))
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()

		return nil, fmt.Errorf("creating the JetStream context: %w", err)
	}

	deadline := time.Now().Add(wait)

	for {
		kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{Bucket: ConfigBucket, History: 5, Storage: jetstream.FileStorage})
		if err == nil {
			m := &ConfigMirror{store: store, cipher: cipher, nc: nc, kv: kv, wake: make(chan struct{}, 1)}
			store.SetOnChange(m.Notify)

			return m, nil
		}

		if time.Now().After(deadline) || ctx.Err() != nil {
			nc.Close()

			return nil, fmt.Errorf("creating the %s bucket: %w", ConfigBucket, err)
		}

		time.Sleep(time.Second)
	}
}

// Notify asks for a publish soon. It never blocks, and asking twice before it runs is asking once.
func (m *ConfigMirror) Notify() {
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

// Sync publishes the current configuration, or removes the key when there is none.
func (m *ConfigMirror) Sync(ctx context.Context) error {
	snap, err := m.store.Snapshot()
	if err != nil {
		return err
	}

	if snap == nil || (snap.Default == nil && len(snap.Tenants) == 0) {
		if err := m.kv.Purge(ctx, ConfigKey); err != nil && err != jetstream.ErrKeyNotFound {
			return err
		}

		return nil
	}

	raw, err := json.Marshal(snap)
	if err != nil {
		return err
	}

	sealed, err := m.cipher.Encrypt(string(raw))
	if err != nil {
		return err
	}

	_, err = m.kv.Put(ctx, ConfigKey, []byte(sealed))

	return err
}

// Run publishes now, on every change and every few minutes, until ctx is done.
func (m *ConfigMirror) Run(ctx context.Context) {
	defer m.nc.Close()

	tick := time.NewTicker(resync)
	defer tick.Stop()

	for {
		sctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		if err := m.Sync(sctx); err != nil {
			_ = catcher.Error("soc-ai: cannot publish the configuration to NATS", err, nil)
		}
		cancel()

		select {
		case <-ctx.Done():
			return
		case <-m.wake:
		case <-tick.C:
		}
	}
}
