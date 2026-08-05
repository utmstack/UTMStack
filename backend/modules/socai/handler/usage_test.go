package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/pkg/authz"
)

func init() { gin.SetMode(gin.TestMode) }

const usageTenant = "8f1c1b8e-0000-4000-8000-000000000001"

func getUsage(t *testing.T, limit int, used int64, limitErr, usedErr error) (int, usageResponse) {
	t.Helper()

	h := NewUsageHandler(
		func(context.Context, string) (int, error) { return limit, limitErr },
		func(context.Context, string) (int64, error) { return used, usedErr },
	)

	rec := httptest.NewRecorder()
	e := gin.New()
	e.GET("/usage", h.Usage)

	req := httptest.NewRequest(http.MethodGet, "/usage", nil)
	req = req.WithContext(authz.WithTenantID(req.Context(), usageTenant))
	e.ServeHTTP(rec, req)

	var out usageResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestUsageReportsWhatIsLeft(t *testing.T) {
	code, got := getUsage(t, 500, 312, nil, nil)

	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if got.Limit != 500 || got.Used != 312 {
		t.Errorf("limit/used = %d/%d, want 500/312", got.Limit, got.Used)
	}
	if got.Remaining == nil || *got.Remaining != 188 {
		t.Errorf("remaining = %v, want 188", got.Remaining)
	}
	if !got.ResetsAt.After(time.Now().UTC()) {
		t.Errorf("resetsAt = %s, want it in the future", got.ResetsAt)
	}
}

// Counting past the limit is how the gate stays safe under retries, so what is
// left is reported as none rather than as a negative number.
func TestUsageNeverReportsNegativeRemaining(t *testing.T) {
	_, got := getUsage(t, 500, 520, nil, nil)

	if got.Remaining == nil || *got.Remaining != 0 {
		t.Errorf("remaining = %v, want 0", got.Remaining)
	}
}

// No limit means there is nothing to have left.
func TestUsageOmitsRemainingWithoutALimit(t *testing.T) {
	_, got := getUsage(t, 0, 42, nil, nil)

	if got.Remaining != nil {
		t.Errorf("remaining = %v, want it absent", *got.Remaining)
	}
	if got.Used != 42 {
		t.Errorf("used = %d, want it still reported", got.Used)
	}
}

// Unlike the gate, this one does not fail open: a wrong number here would be
// read as fact, and no AI call depends on the answer.
func TestUsageFailsClosed(t *testing.T) {
	if code, _ := getUsage(t, 0, 0, errors.New("db down"), nil); code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 when the limit cannot be read", code)
	}
	if code, _ := getUsage(t, 500, 0, nil, errors.New("db down")); code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 when the count cannot be read", code)
	}
}
