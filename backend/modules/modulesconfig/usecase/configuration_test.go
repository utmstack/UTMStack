package usecase

import (
	"context"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/dto"
	"github.com/utmstack/utmstack/backend/pkg/secret"
)

type fakeConfigRepo struct{ saved []domain.UtmModuleGroupConfiguration }

func (r *fakeConfigRepo) ListByGroupID(context.Context, int64) ([]domain.UtmModuleGroupConfiguration, error) {
	return nil, nil
}
func (r *fakeConfigRepo) GetByGroupAndKey(context.Context, int64, string) (*domain.UtmModuleGroupConfiguration, error) {
	return nil, nil
}
func (r *fakeConfigRepo) SaveAll(_ context.Context, c []domain.UtmModuleGroupConfiguration) error {
	r.saved = append([]domain.UtmModuleGroupConfiguration(nil), c...)
	return nil
}

type fakeModuleRepo struct{ module domain.UtmModule }

func (r *fakeModuleRepo) GetByID(context.Context, int64) (*domain.UtmModule, error) {
	m := r.module
	return &m, nil
}
func (r *fakeModuleRepo) GetByServerAndName(context.Context, int64, string) (*domain.UtmModule, error) {
	return nil, domain.ErrModuleNotFound
}
func (r *fakeModuleRepo) List(context.Context, connectors.ModuleListFilter) ([]domain.UtmModule, int64, error) {
	return nil, 0, nil
}
func (r *fakeModuleRepo) Save(_ context.Context, m *domain.UtmModule) error {
	r.module = *m
	return nil
}
func (r *fakeModuleRepo) Categories(context.Context, *int64) ([]string, error) { return nil, nil }
func (r *fakeModuleRepo) CountActiveByName(context.Context, string) (int64, error) {
	return 0, nil
}

func newCipher(t *testing.T) *secret.Cipher {
	t.Helper()
	c, err := secret.NewCipher("test-encryption-key")
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	return c
}

func factoryWith(name string) connectors.ModuleFactory {
	f := NewModuleFactory()
	f.Register(stubKind{name: name})
	return f
}

// Masked-value passthrough: when the UI re-PUTs a sensitive field whose value
// is the mask sentinel, the usecase must drop that row from the write set so
// the stored ciphertext isn't clobbered.
func TestConfigUpdate_DropsMaskedSensitive(t *testing.T) {
	configs := &fakeConfigRepo{}
	modules := &fakeModuleRepo{module: domain.UtmModule{ID: 1, ModuleName: "SOPHOS"}}
	uc := NewConfigUsecase(configs, modules, factoryWith("SOPHOS"), newCipher(t))

	err := uc.Update(context.Background(), dto.UpdateGroupConfigurationRequest{
		ModuleID: 1,
		Keys: []dto.GroupConfigurationItemInput{
			{GroupID: 10, ConfKey: "secret", ConfDataType: domain.ConfTypePassword, ConfValue: domain.MaskedValue},
			{GroupID: 10, ConfKey: "host", ConfDataType: domain.ConfTypeText, ConfValue: "example.com"},
		},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if len(configs.saved) != 1 {
		t.Fatalf("expected 1 saved row (masked dropped), got %d", len(configs.saved))
	}
	if configs.saved[0].ConfKey != "host" {
		t.Fatalf("expected host to survive, got %s", configs.saved[0].ConfKey)
	}
}

// Required-empty rejection: a required field submitted blank must return an
// error and abort the write — no partial save.
func TestConfigUpdate_RejectsRequiredEmpty(t *testing.T) {
	configs := &fakeConfigRepo{}
	modules := &fakeModuleRepo{module: domain.UtmModule{ID: 1, ModuleName: "SOPHOS"}}
	uc := NewConfigUsecase(configs, modules, factoryWith("SOPHOS"), newCipher(t))

	err := uc.Update(context.Background(), dto.UpdateGroupConfigurationRequest{
		ModuleID: 1,
		Keys: []dto.GroupConfigurationItemInput{
			{GroupID: 10, ConfKey: "host", ConfDataType: domain.ConfTypeText, ConfRequired: true, ConfValue: ""},
		},
	})
	if err == nil {
		t.Fatal("expected required-empty error")
	}
	if len(configs.saved) != 0 {
		t.Fatalf("expected no rows saved on validation failure, got %d", len(configs.saved))
	}
}

// Encryption: a non-masked sensitive value gets ciphered before being saved,
// and the saved blob is not the plaintext.
func TestConfigUpdate_EncryptsSensitiveValues(t *testing.T) {
	configs := &fakeConfigRepo{}
	modules := &fakeModuleRepo{module: domain.UtmModule{ID: 1, ModuleName: "SOPHOS"}}
	cipher := newCipher(t)
	uc := NewConfigUsecase(configs, modules, factoryWith("SOPHOS"), cipher)

	const plain = "super-secret-token"
	err := uc.Update(context.Background(), dto.UpdateGroupConfigurationRequest{
		ModuleID: 1,
		Keys: []dto.GroupConfigurationItemInput{
			{GroupID: 10, ConfKey: "token", ConfDataType: domain.ConfTypePassword, ConfValue: plain},
		},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if len(configs.saved) != 1 {
		t.Fatalf("expected 1 saved row, got %d", len(configs.saved))
	}
	if configs.saved[0].ConfValue == plain {
		t.Fatal("expected stored value to be encrypted, got plaintext")
	}
	dec, err := cipher.Decrypt(configs.saved[0].ConfValue)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if dec != plain {
		t.Fatalf("expected round-trip to recover plaintext, got %q", dec)
	}
}

// needs_restart flag flip: SOPHOS is in the allow-list, so an update should
// set needs_restart=true on the parent module.
func TestConfigUpdate_FlipsNeedsRestart(t *testing.T) {
	configs := &fakeConfigRepo{}
	modules := &fakeModuleRepo{module: domain.UtmModule{ID: 1, ModuleName: "SOPHOS", NeedsRestart: false}}
	uc := NewConfigUsecase(configs, modules, factoryWith("SOPHOS"), newCipher(t))

	err := uc.Update(context.Background(), dto.UpdateGroupConfigurationRequest{
		ModuleID: 1,
		Keys: []dto.GroupConfigurationItemInput{
			{GroupID: 10, ConfKey: "host", ConfDataType: domain.ConfTypeText, ConfValue: "example.com"},
		},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !modules.module.NeedsRestart {
		t.Fatal("expected SOPHOS to be flagged needs_restart after update")
	}
}

// Non-restart module: a module outside the allow-list keeps needs_restart=false.
func TestConfigUpdate_DoesNotFlipForNonRestartModule(t *testing.T) {
	configs := &fakeConfigRepo{}
	modules := &fakeModuleRepo{module: domain.UtmModule{ID: 1, ModuleName: "LINUX_LOGS", NeedsRestart: false}}
	uc := NewConfigUsecase(configs, modules, factoryWith("LINUX_LOGS"), newCipher(t))

	err := uc.Update(context.Background(), dto.UpdateGroupConfigurationRequest{
		ModuleID: 1,
		Keys: []dto.GroupConfigurationItemInput{
			{GroupID: 10, ConfKey: "host", ConfDataType: domain.ConfTypeText, ConfValue: "example.com"},
		},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if modules.module.NeedsRestart {
		t.Fatal("expected LINUX_LOGS to keep needs_restart=false")
	}
}
