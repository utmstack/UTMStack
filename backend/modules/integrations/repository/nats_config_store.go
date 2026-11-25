package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"gopkg.in/yaml.v3"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
)

// ConfigBucket is the NATS key-value bucket the integrations read their module groups from, one key
// per integration. The containers watch it, so a change reaches them without a shared folder.
const ConfigBucket = "integration-config"

// document is how every stored value starts once decrypted; the containers check for it.
const document = "plugins:"

const maxWriteAttempts = 10

var _ connectors.ConfigRepository = (*NATSConfigStore)(nil)

// NATSConfigStore keeps the module groups of each integration in NATS, encrypted. It is the store the
// integrations' containers read, so the whole document is encrypted, not only the secrets in it.
type NATSConfigStore struct {
	kv     jetstream.KeyValue
	cipher connectors.Cipher
	nc     *nats.Conn
}

// NewNATSConfigStore connects to NATS and creates the bucket if it does not exist. NATS may still be
// starting when the backend is, so it waits up to wait for it.
func NewNATSConfigStore(ctx context.Context, url string, cipher connectors.Cipher, wait time.Duration) (*NATSConfigStore, error) {
	nc, err := nats.Connect(url, nats.MaxReconnects(-1), nats.RetryOnFailedConnect(true), nats.Name("backend-integrations"))
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
		kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
			Bucket:  ConfigBucket,
			History: 5,
			Storage: jetstream.FileStorage,
		})
		if err == nil {
			return &NATSConfigStore{kv: kv, cipher: cipher, nc: nc}, nil
		}

		if time.Now().After(deadline) || ctx.Err() != nil {
			nc.Close()

			return nil, fmt.Errorf("creating the %s bucket: %w", ConfigBucket, err)
		}

		time.Sleep(time.Second)
	}
}

func (s *NATSConfigStore) Close() { s.nc.Close() }

// read returns the stored document and its revision; a key that does not exist is an empty document
// at revision 0.
func (s *NATSConfigStore) read(ctx context.Context, integration string) (pluginsFile, uint64, error) {
	pf := pluginsFile{Plugins: map[string]pluginEntry{}}

	e, err := s.kv.Get(ctx, pluginKey(integration))
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return pf, 0, nil
	}

	if err != nil {
		return pf, 0, err
	}

	plain, err := s.cipher.Decrypt(string(e.Value()))
	if err != nil {
		return pf, 0, fmt.Errorf("decrypting the %s configuration: %w", pluginKey(integration), err)
	}

	if !strings.HasPrefix(strings.TrimSpace(plain), document) {
		return pf, 0, fmt.Errorf("the %s configuration did not decrypt: wrong ENCRYPTION_KEY?", pluginKey(integration))
	}

	if err := yaml.Unmarshal([]byte(plain), &pf); err != nil {
		return pluginsFile{Plugins: map[string]pluginEntry{}}, 0, err
	}

	if pf.Plugins == nil {
		pf.Plugins = map[string]pluginEntry{}
	}

	return pf, e.Revision(), nil
}

// write stores the document if nobody changed it since revision (0 means it must not exist yet).
func (s *NATSConfigStore) write(ctx context.Context, integration string, pf pluginsFile, tenants []domain.TenantConfig, revision uint64) error {
	if pf.Plugins == nil {
		pf.Plugins = map[string]pluginEntry{}
	}

	pf.Plugins[pluginKey(integration)] = pluginEntry{Tenants: tenants}

	data, err := yaml.Marshal(pf)
	if err != nil {
		return err
	}

	sealed, err := s.cipher.Encrypt(string(data))
	if err != nil {
		return err
	}

	if revision == 0 {
		_, err = s.kv.Create(ctx, pluginKey(integration), []byte(sealed))
	} else {
		_, err = s.kv.Update(ctx, pluginKey(integration), []byte(sealed), revision)
	}

	return err
}

func isConflict(err error) bool {
	return errors.Is(err, jetstream.ErrKeyExists) || strings.Contains(fmt.Sprint(err), "wrong last sequence")
}

// modify applies change to the stored tenants and writes the result, trying again when another
// replica of the backend changed the document in between.
func (s *NATSConfigStore) modify(ctx context.Context, integration string, change func([]domain.TenantConfig) ([]domain.TenantConfig, error)) error {
	for attempt := 0; attempt < maxWriteAttempts; attempt++ {
		pf, revision, err := s.read(ctx, integration)
		if err != nil {
			return err
		}

		tenants, err := change(pf.Plugins[pluginKey(integration)].Tenants)
		if err != nil {
			return err
		}

		err = s.write(ctx, integration, pf, tenants, revision)
		if err == nil {
			return nil
		}

		if !isConflict(err) {
			return err
		}
	}

	return fmt.Errorf("the %s configuration kept changing under this write", pluginKey(integration))
}

func (s *NATSConfigStore) Load(ctx context.Context, integration string) ([]domain.ConfigGroup, error) {
	tenant, err := callerTenant(ctx)
	if err != nil {
		return nil, err
	}

	pf, _, err := s.read(ctx, integration)
	if err != nil {
		return nil, err
	}

	for _, tc := range pf.Plugins[pluginKey(integration)].Tenants {
		if tc.ID == tenant {
			return tc.Groups, nil
		}
	}

	return []domain.ConfigGroup{}, nil
}

func (s *NATSConfigStore) LoadAllTenants(ctx context.Context, integration string) ([]domain.TenantConfig, error) {
	pf, _, err := s.read(ctx, integration)
	if err != nil {
		return nil, err
	}

	return pf.Plugins[pluginKey(integration)].Tenants, nil
}

func (s *NATSConfigStore) Save(ctx context.Context, integration string, groups []domain.ConfigGroup) error {
	tenant, err := callerTenant(ctx)
	if err != nil {
		return err
	}

	return s.modify(ctx, integration, func(tenants []domain.TenantConfig) ([]domain.TenantConfig, error) {
		return replaceTenant(tenants, tenant, groups), nil
	})
}

func (s *NATSConfigStore) Upsert(ctx context.Context, integration string, group domain.ConfigGroup) error {
	tenant, err := callerTenant(ctx)
	if err != nil {
		return err
	}

	return s.modify(ctx, integration, func(tenants []domain.TenantConfig) ([]domain.TenantConfig, error) {
		groups := groupsOf(tenants, tenant)

		replaced := false

		for i := range groups {
			if groups[i].Name == group.Name {
				groups[i], replaced = group, true

				break
			}
		}

		if !replaced {
			groups = append(groups, group)
		}

		return replaceTenant(tenants, tenant, groups), nil
	})
}

func (s *NATSConfigStore) Delete(ctx context.Context, integration, groupName string) error {
	tenant, err := callerTenant(ctx)
	if err != nil {
		return err
	}

	return s.modify(ctx, integration, func(tenants []domain.TenantConfig) ([]domain.TenantConfig, error) {
		groups := groupsOf(tenants, tenant)
		out := make([]domain.ConfigGroup, 0, len(groups))
		found := false

		for _, g := range groups {
			if g.Name == groupName {
				found = true

				continue
			}

			out = append(out, g)
		}

		if !found {
			return nil, domain.ErrConfigGroupNotFound
		}

		return replaceTenant(tenants, tenant, out), nil
	})
}
