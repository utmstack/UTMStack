package usecase

import (
	"context"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/appconfig/domain"
)

// A tenant that never customized branding must NOT see the master's brand.
// Bug: master edits branding → tenants without their own row inherited it via
// GetByKey's default-tenant fallback. Fix: read via GetOwn.
type fakeRepo struct{ own *domain.Config }

func (f *fakeRepo) List(context.Context) ([]domain.Config, error)  { return nil, nil }
func (f *fakeRepo) Save(context.Context, *domain.Config) error     { return nil }
func (f *fakeRepo) GetByKey(context.Context, string) (*domain.Config, error) {
	return &domain.Config{Key: brandingConfigKey, Value: `{"enabled":true,"productName":"MasterCo"}`}, nil
}
func (f *fakeRepo) GetOwn(context.Context, string) (*domain.Config, error) { return f.own, nil }
func (f *fakeRepo) CountValueContains(context.Context, string, string) (int, error) {
	return 0, nil
}

func TestBrandingDoesNotInheritMaster(t *testing.T) {
	s := NewBranding(&fakeRepo{own: nil})
	got, err := s.Get(context.Background())
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ProductName != defaultProductName {
		t.Errorf("productName = %q, want defaults (%q) — leaked master brand", got.ProductName, defaultProductName)
	}
	if got.Enabled {
		t.Error("enabled = true, want false — leaked master brand")
	}
}

// SetAsset must return the URL that used to occupy the slot so the caller can
// decide whether to garbage-collect the old file. Without this, an admin
// re-upload silently orphans the previous file and its cleanup nukes files
// other tenants still point to.
func TestSetAssetReturnsPreviousURL(t *testing.T) {
	repo := &fakeRepo{own: &domain.Config{Key: brandingConfigKey, Value: `{"logoURL":"/uploads/branding/logo-old.png"}`}}
	s := NewBranding(repo)
	_, previous, err := s.SetAsset(context.Background(), "alex", AssetLogo, "/uploads/branding/logo-new.png")
	if err != nil {
		t.Fatalf("SetAsset: %v", err)
	}
	if previous != "/uploads/branding/logo-old.png" {
		t.Errorf("previous = %q, want the URL that was in the slot before", previous)
	}
}

func TestSetAssetPreviousIsEmptyWhenSlotWasUnset(t *testing.T) {
	s := NewBranding(&fakeRepo{own: nil})
	_, previous, err := s.SetAsset(context.Background(), "alex", AssetLogo, "/uploads/branding/logo-first.png")
	if err != nil {
		t.Fatalf("SetAsset: %v", err)
	}
	if previous != "" {
		t.Errorf("previous = %q, want empty — nothing to GC on first upload", previous)
	}
}

// IsBrandingAssetReferenced routes through the repo count so callers only delete
// files no tenant still points to.
type countingRepo struct {
	fakeRepo
	count int
}

func (c *countingRepo) CountValueContains(context.Context, string, string) (int, error) {
	return c.count, nil
}

func TestIsReferencedTracksCount(t *testing.T) {
	s := NewBranding(&countingRepo{count: 2})
	ref, err := s.IsBrandingAssetReferenced(context.Background(), "/uploads/branding/logo-x.png")
	if err != nil || !ref {
		t.Fatalf("want referenced=true, got %v err=%v", ref, err)
	}

	s = NewBranding(&countingRepo{count: 0})
	ref, err = s.IsBrandingAssetReferenced(context.Background(), "/uploads/branding/logo-x.png")
	if err != nil || ref {
		t.Fatalf("want referenced=false, got %v err=%v", ref, err)
	}
}
