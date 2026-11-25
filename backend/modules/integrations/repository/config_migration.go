package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/nats-io/nats.go/jetstream"
	"gopkg.in/yaml.v3"
)

// MigrateFiles moves the module configuration the backend used to keep in system_plugins_<name>.yaml
// files into NATS, and returns how many integrations it imported. A file whose integration already
// has a value in NATS is left to that value, which is newer. A file is renamed to <file>.migrated only
// after what was stored has been read back and matches, so a failure leaves it where it was and the
// next start tries again. One integration failing does not stop the others.
func (s *NATSConfigStore) MigrateFiles(ctx context.Context, dir string) (int, error) {
	var (
		imported int
		errs     []error
		seen     = map[string]bool{}
	)

	for _, key := range pluginNames {
		if seen[key] {
			continue
		}

		seen[key] = true

		ok, err := s.migrateOne(ctx, dir, key)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))

			continue
		}

		if ok {
			imported++
		}
	}

	return imported, errors.Join(errs...)
}

func (s *NATSConfigStore) migrateOne(ctx context.Context, dir, key string) (bool, error) {
	path := filepath.Join(dir, "system_plugins_"+key+".yaml")

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	var pf pluginsFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return false, fmt.Errorf("%s is not valid YAML: %w", path, err)
	}

	if pf.Plugins == nil {
		pf.Plugins = map[string]pluginEntry{}
	}

	_, revision, err := s.read(ctx, key)
	if err != nil {
		return false, err
	}

	imported := false

	if revision == 0 {
		want := pf.Plugins[key].Tenants

		if err := s.write(ctx, key, pf, want, 0); err != nil && !errors.Is(err, jetstream.ErrKeyExists) && !isConflict(err) {
			return false, fmt.Errorf("storing it in NATS: %w", err)
		}

		// Read it back: the file is only set aside once NATS holds the same thing.
		back, _, err := s.read(ctx, key)
		if err != nil {
			return false, fmt.Errorf("reading it back from NATS: %w", err)
		}

		if !reflect.DeepEqual(normalize(back.Plugins[key].Tenants), normalize(want)) {
			return false, errors.New("what NATS holds does not match the file; the file was kept")
		}

		imported = true
	}

	if err := os.Rename(path, path+".migrated"); err != nil {
		return imported, fmt.Errorf("setting the file aside: %w", err)
	}

	_ = os.Remove(path + ".lock")

	return imported, nil
}

// normalize makes "no groups" compare equal whether it came back as nil or empty.
func normalize[T any](in []T) []T {
	if len(in) == 0 {
		return nil
	}

	return in
}
