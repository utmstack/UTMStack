package socai

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/utmstack/utmstack/backend/pkg/authz"
)

func init() { gin.SetMode(gin.TestMode) }

const quotaTenant = "8f1c1b8e-0000-4000-8000-000000000001"

// run sends one request through the gate and reports the status and whether the
// handler behind it was reached.
func run(t *testing.T, q *AIQuota, tenant string) (int, bool) {
	t.Helper()

	rec := httptest.NewRecorder()
	e := gin.New()
	reached := false
	e.POST("/chat", q.Gate(), func(c *gin.Context) {
		reached = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/chat", nil)
	if tenant != "" {
		req = req.WithContext(authz.WithTenantID(req.Context(), tenant))
	}
	e.ServeHTTP(rec, req)

	return rec.Code, reached
}

func quota(limit int, used *int64) *AIQuota {
	return &AIQuota{
		LimitOf: func(context.Context, string) (int, error) { return limit, nil },
		Consume: func(context.Context, string) (int64, error) { *used++; return *used, nil },
	}
}

func TestQuotaAllowsUpToTheLimit(t *testing.T) {
	var used int64
	q := quota(3, &used)

	for i := 1; i <= 3; i++ {
		if code, reached := run(t, q, quotaTenant); !reached || code != http.StatusOK {
			t.Fatalf("request %d: status %d, reached %v; want it served", i, code, reached)
		}
	}
}

func TestQuotaRefusesPastTheLimit(t *testing.T) {
	var used int64
	q := quota(3, &used)

	for range 3 {
		run(t, q, quotaTenant)
	}

	code, reached := run(t, q, quotaTenant)
	if reached {
		t.Error("the fourth request reached the handler")
	}
	if code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", code)
	}
}

// Zero is how a limit is lifted, and it must not cost a count either.
func TestNoLimitDoesNotEvenCount(t *testing.T) {
	var used int64
	q := quota(0, &used)

	if code, reached := run(t, q, quotaTenant); !reached || code != http.StatusOK {
		t.Fatalf("status %d, reached %v; want it served", code, reached)
	}
	if used != 0 {
		t.Errorf("consumed %d, want nothing counted when there is no limit", used)
	}
}

// A tenant locked out because a query failed costs more than an AI call that
// should have been refused.
func TestQuotaFailsOpen(t *testing.T) {
	broken := &AIQuota{
		LimitOf: func(context.Context, string) (int, error) { return 0, errors.New("db down") },
		Consume: func(context.Context, string) (int64, error) { return 0, nil },
	}
	if code, reached := run(t, broken, quotaTenant); !reached || code != http.StatusOK {
		t.Errorf("status %d, reached %v; want it served when the limit cannot be read", code, reached)
	}

	uncountable := &AIQuota{
		LimitOf: func(context.Context, string) (int, error) { return 5, nil },
		Consume: func(context.Context, string) (int64, error) { return 0, errors.New("db down") },
	}
	if code, reached := run(t, uncountable, quotaTenant); !reached || code != http.StatusOK {
		t.Errorf("status %d, reached %v; want it served when the count fails", code, reached)
	}
}

// An install that does not meter AI gets a gate that does nothing.
func TestNilQuotaIsAPassThrough(t *testing.T) {
	var q *AIQuota
	if code, reached := run(t, q, quotaTenant); !reached || code != http.StatusOK {
		t.Errorf("status %d, reached %v; want it served", code, reached)
	}
}

// The internal route the plugin calls is the same gate the user-facing ones
// use, so automatic analysis and a person in the UI spend the same allowance
// under the same rules.
func TestInternalConsumeIsTheSameGate(t *testing.T) {
	var used int64
	q := quota(2, &used)

	consume := func() int {
		rec := httptest.NewRecorder()
		e := gin.New()
		e.POST("/consume", q.Gate(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

		req := httptest.NewRequest(http.MethodPost, "/consume", nil)
		req = req.WithContext(authz.WithTenantID(req.Context(), quotaTenant))
		e.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := consume(); code != http.StatusNoContent {
		t.Fatalf("first consume = %d, want 204", code)
	}
	if code := consume(); code != http.StatusNoContent {
		t.Fatalf("second consume = %d, want 204", code)
	}
	if code := consume(); code != http.StatusTooManyRequests {
		t.Errorf("third consume = %d, want 429", code)
	}

	// And a user-facing request now finds the allowance already spent.
	if code, reached := run(t, q, quotaTenant); reached || code != http.StatusTooManyRequests {
		t.Errorf("chat status = %d, reached %v; want it refused on the same counter", code, reached)
	}
}
