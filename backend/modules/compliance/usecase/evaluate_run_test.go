package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

const evalTenant = "ce66672c-e36d-4761-a8c8-90058fee1a24"

// One evidence check per source and one limit, the shape the shipped controls use.
const evalControl = `id: si-3
name: Malicious Code Protection
scope: data
strategy: %s
checks:
  - key: windows-activity
    name: Windows activity
    dataType: wineventlog
    rule: MIN_HITS_REQUIRED
    ruleValue: 1
  - key: linux-activity
    name: Linux activity
    dataType: linux
    rule: MIN_HITS_REQUIRED
    ruleValue: 1
  - key: windows-disabled
    name: Windows protection disabled
    dataType: wineventlog
    rule: THRESHOLD_MAX
    ruleValue: 0
`

func evalFixture(t *testing.T, body string) (*evaluator, *domain.Control) {
	t.Helper()
	dir := t.TempDir()
	sys := filepath.Join(dir, "system")
	if err := os.MkdirAll(sys, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sys, "si-3.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cs := NewControlStore(sys, filepath.Join(dir, "user"))
	if err := cs.Load(); err != nil {
		t.Fatal(err)
	}
	e := &evaluator{controls: cs, coverage: NewCoverageIndex(filepath.Join(dir, "rules"))}
	c, ok := cs.Get(authz.WithTenantID(context.Background(), evalTenant), "si-3")
	if !ok {
		t.Fatal("control did not load")
	}
	return e, c
}

func outcomes(c *domain.Control, by map[string]domain.CheckOutcome) map[checkSig]checkOutcome {
	res := map[checkSig]checkOutcome{}
	for _, ch := range c.Checks {
		o, ok := by[ch.Key]
		if !ok {
			o = domain.CheckNotApplicable
		}
		res[sigOf(ch)] = checkOutcome{outcome: o}
	}
	return res
}

func TestControlRowLimitsAndEvidence(t *testing.T) {
	const (
		P  = domain.CheckPassed
		F  = domain.CheckFailed
		NA = domain.CheckNotApplicable
	)
	cases := []struct {
		name     string
		strategy string
		by       map[string]domain.CheckOutcome
		want     domain.ComplianceStatus
	}{
		{"any: evidence on one source, limit kept", "ANY",
			map[string]domain.CheckOutcome{"windows-activity": P, "linux-activity": NA, "windows-disabled": P}, domain.StatusCompliant},
		{"any: an exceeded limit fails despite evidence", "ANY",
			map[string]domain.CheckOutcome{"windows-activity": P, "linux-activity": P, "windows-disabled": F}, domain.StatusNonCompliant},
		{"any: a kept limit does not stand in for missing evidence", "ANY",
			map[string]domain.CheckOutcome{"windows-activity": F, "linux-activity": NA, "windows-disabled": P}, domain.StatusNonCompliant},
		{"any: evidence from another source is enough", "ANY",
			map[string]domain.CheckOutcome{"windows-activity": F, "linux-activity": P, "windows-disabled": P}, domain.StatusCompliant},
		{"any: no source sent", "ANY",
			map[string]domain.CheckOutcome{}, domain.StatusNotEvaluated},
		{"all: an exceeded limit fails", "ALL",
			map[string]domain.CheckOutcome{"windows-activity": P, "linux-activity": P, "windows-disabled": F}, domain.StatusNonCompliant},
		{"all: unchanged, a missing source is still at risk", "ALL",
			map[string]domain.CheckOutcome{"windows-activity": P, "linux-activity": NA, "windows-disabled": P}, domain.StatusAtRisk},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e, c := evalFixture(t, fmt.Sprintf(evalControl, tc.strategy))
			ctx := authz.WithTenantID(context.Background(), evalTenant)
			row := e.controlRow(ctx, "si-3", outcomes(c, tc.by), nil)
			if row.EngineStatus != tc.want {
				t.Fatalf("status %s, want %s (evidence %q)", row.EngineStatus, tc.want, row.Evidence)
			}
		})
	}
}

// Only-limit controls (au-9: no audit log cleared) pass when every source they
// apply to kept the limit, and are not evaluated when the tenant sends none.
func TestControlRowOnlyLimits(t *testing.T) {
	e, c := evalFixture(t, `id: si-3
name: limits only
scope: data
strategy: ANY
checks:
  - key: windows-disabled
    name: Windows protection disabled
    dataType: wineventlog
    rule: THRESHOLD_MAX
    ruleValue: 0
`)
	ctx := authz.WithTenantID(context.Background(), evalTenant)
	row := e.controlRow(ctx, "si-3", outcomes(c, map[string]domain.CheckOutcome{"windows-disabled": domain.CheckPassed}), nil)
	if row.EngineStatus != domain.StatusCompliant {
		t.Fatalf("status %s, want COMPLIANT (%q)", row.EngineStatus, row.Evidence)
	}
	row = e.controlRow(ctx, "si-3", outcomes(c, map[string]domain.CheckOutcome{"windows-disabled": domain.CheckFailed}), nil)
	if row.EngineStatus != domain.StatusNonCompliant {
		t.Fatalf("status %s, want NON_COMPLIANT (%q)", row.EngineStatus, row.Evidence)
	}
	row = e.controlRow(ctx, "si-3", outcomes(c, map[string]domain.CheckOutcome{}), nil)
	if row.EngineStatus != domain.StatusNotEvaluated {
		t.Fatalf("status %s, want NOT_EVALUATED (%q)", row.EngineStatus, row.Evidence)
	}
}
