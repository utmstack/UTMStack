// Package modulecfg delivers the configuration of an integration to its container. The backend keeps
// each integration's module groups in a NATS key-value bucket, encrypted; the container watches its
// key, so it gets the current value when it starts and every change after that, with no shared
// folder between the two.
package modulecfg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Bucket is where the backend writes the configuration, one key per integration (aws, azure, ...).
const Bucket = "integration-config"

// document starts every configuration the backend writes. A cipher that does not recognise its input
// hands it back unchanged, so a payload that does not start like this was not decrypted.
const document = "plugins:"

// Decrypter is the part of the shared cipher a watcher needs.
type Decrypter interface {
	Decrypt(string) (string, error)
}

// retryDelay is how long to wait for NATS or the bucket to exist. A variable so tests do not wait.
var retryDelay = 5 * time.Second

// Open decrypts a stored value into the YAML document it carries.
func Open(dec Decrypter, stored string) ([]byte, error) {
	plain, err := dec.Decrypt(stored)
	if err != nil {
		return nil, fmt.Errorf("decrypting the configuration: %w", err)
	}

	if !strings.HasPrefix(strings.TrimSpace(plain), document) {
		return nil, errors.New("the configuration did not decrypt: wrong key, or not written by the backend")
	}

	return []byte(plain), nil
}

// Watch calls onChange with the configuration stored under key: right away with the current value,
// and again on every change. A nil value means the integration has no configuration (the key does
// not exist or was removed). It blocks until ctx is done, waiting for NATS and for the bucket to
// exist, and reconnecting when the watch is lost. A value that cannot be read is reported to
// onError and the previous configuration stays in force.
func Watch(ctx context.Context, natsURL, key string, dec Decrypter, onChange func(yaml []byte), onError func(error)) error {
	if natsURL == "" {
		return errors.New("NATS_URL is required to read the module configuration")
	}

	nc, err := nats.Connect(natsURL,
		nats.MaxReconnects(-1),
		nats.RetryOnFailedConnect(true),
		nats.Name("modulecfg-"+key),
	)
	if err != nil {
		return fmt.Errorf("connecting to NATS: %w", err)
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		return fmt.Errorf("creating the JetStream context: %w", err)
	}

	for ctx.Err() == nil {
		if err := watchOnce(ctx, js, key, dec, onChange, onError); err != nil && ctx.Err() == nil {
			onError(err)
		}

		select {
		case <-ctx.Done():
		case <-time.After(retryDelay):
		}
	}

	return ctx.Err()
}

func watchOnce(ctx context.Context, js jetstream.JetStream, key string, dec Decrypter, onChange func([]byte), onError func(error)) error {
	kv, err := js.KeyValue(ctx, Bucket)
	if err != nil {
		return fmt.Errorf("opening the %s bucket: %w", Bucket, err)
	}

	w, err := kv.Watch(ctx, key)
	if err != nil {
		return fmt.Errorf("watching %s: %w", key, err)
	}
	defer func() { _ = w.Stop() }()

	seen := false

	for {
		select {
		case <-ctx.Done():
			return nil
		case e, ok := <-w.Updates():
			if !ok {
				return errors.New("the watch ended")
			}

			// A nil entry marks the end of the stored values the watch starts with. With none
			// before it, the key does not exist.
			if e == nil {
				if !seen {
					onChange(nil)
					seen = true
				}

				continue
			}

			seen = true

			switch e.Operation() {
			case jetstream.KeyValueDelete, jetstream.KeyValuePurge:
				onChange(nil)
			default:
				yaml, err := Open(dec, string(e.Value()))
				if err != nil {
					onError(fmt.Errorf("%s: %w", key, err))

					continue
				}

				onChange(yaml)
			}
		}
	}
}
