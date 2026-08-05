package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/modules/billing/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

func init() { gin.SetMode(gin.TestMode) }

const customerTenant = "8f1c1b8e-0000-4000-8000-000000000001"

type fakeLicense struct{ lic domain.License }

func (f fakeLicense) Current() domain.License { return f.lic }
func (f fakeLicense) Replace([]byte) (domain.License, error) {
	return f.lic, nil
}

func enterprise() domain.License {
	return domain.License{
		Edition:          domain.EditionEnterprise,
		MSSP:             true,
		IngestGBPerMonth: 5000,
		Type:             "online",
		ExpiresAt:        time.Now().UTC().AddDate(1, 0, 0),
	}
}

func get(t *testing.T, tenant string) licenseView {
	t.Helper()

	h := NewLicenseHandler(fakeLicense{lic: enterprise()})

	rec := httptest.NewRecorder()
	e := gin.New()
	e.GET("/license", func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Set("user_login", "someone")
		c.Set("tenant_id", tenant)
		h.Get(c)
	})
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/license", nil))

	var out licenseView
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return out
}

// The instance's own tenant sees the whole thing: on an install that is one
// tenant, that is everyone, so nothing that used to be shown disappears.
func TestOwnTenantSeesTheTerms(t *testing.T) {
	got := get(t, authz.DefaultTenantID)

	if got.IngestGBPerMonth == nil || *got.IngestGBPerMonth != 5000 {
		t.Errorf("ingestGbPerMonth = %v, want 5000", got.IngestGBPerMonth)
	}
	if got.Type != "online" {
		t.Errorf("type = %q, want online", got.Type)
	}
	if got.ExpiresAt == nil {
		t.Error("expiresAt was withheld from the instance's own tenant")
	}
}

// A customer on a managed instance has no business knowing what their provider
// contracted or when it runs out — but still needs to know what the product
// offers them.
func TestACustomerSeesOnlyWhatDecidesFeatures(t *testing.T) {
	got := get(t, customerTenant)

	if got.Edition != domain.EditionEnterprise || !got.MSSP {
		t.Errorf("edition/mssp = %q/%v, want them kept", got.Edition, got.MSSP)
	}
	if got.IngestGBPerMonth != nil {
		t.Errorf("ingestGbPerMonth = %v, want it withheld", *got.IngestGBPerMonth)
	}
	if got.Type != "" {
		t.Errorf("type = %q, want it withheld", got.Type)
	}
	if got.ExpiresAt != nil {
		t.Errorf("expiresAt = %v, want it withheld", *got.ExpiresAt)
	}
}
