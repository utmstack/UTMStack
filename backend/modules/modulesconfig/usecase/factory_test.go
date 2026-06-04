package usecase

import (
	"context"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
)

type stubKind struct{ name string }

func (s stubKind) Name() string { return s.name }
func (s stubKind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return []domain.ModuleConfigurationKey{{GroupID: groupID, ConfKey: "k", ConfDataType: domain.ConfTypeText}}
}
func (s stubKind) CheckRequirements(_ context.Context, _ int64) ([]domain.ModuleRequirement, error) {
	return nil, nil
}
func (s stubKind) ValidateConfiguration(_ context.Context, _ *domain.UtmModule, _ []domain.UtmModuleGroupConfiguration) error {
	return nil
}

func TestFactoryRegisterAndLookup(t *testing.T) {
	f := NewModuleFactory()
	if f.Has("SOPHOS") {
		t.Fatalf("expected empty factory to not have SOPHOS")
	}
	f.Register(stubKind{name: "SOPHOS"})
	got, ok := f.Get("SOPHOS")
	if !ok {
		t.Fatalf("expected SOPHOS to be registered")
	}
	if got.Name() != "SOPHOS" {
		t.Fatalf("expected name SOPHOS, got %s", got.Name())
	}
}

func TestFactoryRegisterOverwrites(t *testing.T) {
	f := NewModuleFactory()
	f.Register(stubKind{name: "AZURE"})
	f.Register(stubKind{name: "AZURE"}) // second registration wins silently
	if !f.Has("AZURE") {
		t.Fatalf("expected AZURE registered")
	}
}
