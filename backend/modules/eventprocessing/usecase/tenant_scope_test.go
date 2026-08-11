package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

const (
	tenantA = "8f1c1b8e-0000-4000-8000-00000000000a"
	tenantB = "8f1c1b8e-0000-4000-8000-00000000000b"
)

func ctxFor(tenant string) context.Context {
	return authz.WithTenantID(context.Background(), tenant)
}

func ruleOf(tenant string, system bool) *domain.StoredRule {
	sr := &domain.StoredRule{RelPath: "custom/rule.yaml", System: system}
	sr.TenantId = tenant
	return sr
}

// A rule belongs to the tenant that wrote it. One customer reading, editing or
// deleting another's would be the whole point of multitenancy failing, and the
// identity here is a file path — easy to guess.
func TestATenantCannotSeeAnothersRule(t *testing.T) {
	if visible(ruleOf(tenantB, false), tenantA) {
		t.Error("tenant A can see a rule owned by tenant B")
	}
}

// The shipped catalog and anything global belong to everybody: matching on the
// tenant alone would hide the rules every customer is meant to use.
func TestTheShippedAndGlobalRulesAreVisibleToEveryone(t *testing.T) {
	if !visible(ruleOf("", true), tenantA) {
		t.Error("a system rule is hidden from a tenant")
	}
	if !visible(ruleOf("", false), tenantA) {
		t.Error("a rule with no tenant is hidden from a tenant")
	}
	if !visible(ruleOf(tenantA, false), tenantA) {
		t.Error("a tenant cannot see its own rule")
	}
}

// On-prem runs with no tenant in context and must keep seeing everything.
func TestWithoutATenantEverythingIsVisible(t *testing.T) {
	for _, sr := range []*domain.StoredRule{ruleOf(tenantA, false), ruleOf(tenantB, false), ruleOf("", true)} {
		if !visible(sr, "") {
			t.Errorf("rule %q hidden when no tenant is acting", sr.TenantId)
		}
	}
}

func pipelineOf(tenant string, system bool) *domain.Pipeline {
	return &domain.Pipeline{RelPath: "custom/p.yaml", System: system, TenantID: tenant}
}

func TestPipelineVisibilityFollowsTheSameRule(t *testing.T) {
	if visiblePipeline(pipelineOf(tenantB, false), tenantA) {
		t.Error("tenant A can see a pipeline owned by tenant B")
	}
	if !visiblePipeline(pipelineOf("", true), tenantA) {
		t.Error("a system pipeline is hidden from a tenant")
	}
	if !visiblePipeline(pipelineOf(tenantA, false), tenantA) {
		t.Error("a tenant cannot see its own pipeline")
	}
}

// The store already refuses to rewrite a shipped rule; what this covers is the
// other half — that a tenant is refused another tenant's rule, and told "not
// found" rather than "forbidden" so the answer reveals nothing.
func TestWritingAnotherTenantsRuleReportsNotFound(t *testing.T) {
	uc := &correlationRuleUsecase{store: fakeRuleStore{rule: ruleOf(tenantB, false)}}

	err := uc.writable(ctxFor(tenantA), "custom/rule.yaml")
	if !errors.Is(err, domain.ErrCorrelationRuleNotFound) {
		t.Errorf("err = %v, want ErrCorrelationRuleNotFound", err)
	}
}

func TestWritingAShippedRuleIsRefused(t *testing.T) {
	uc := &correlationRuleUsecase{store: fakeRuleStore{rule: ruleOf("", true)}}

	err := uc.writable(ctxFor(tenantA), "custom/rule.yaml")
	if !errors.Is(err, domain.ErrCorrelationRuleSystemOwner) {
		t.Errorf("err = %v, want ErrCorrelationRuleSystemOwner", err)
	}
}

func TestATenantMayWriteItsOwnRule(t *testing.T) {
	uc := &correlationRuleUsecase{store: fakeRuleStore{rule: ruleOf(tenantA, false)}}

	if err := uc.writable(ctxFor(tenantA), "custom/rule.yaml"); err != nil {
		t.Errorf("a tenant was refused its own rule: %v", err)
	}
}

// fakeRuleStore answers one rule for any path; the rest of the port is unused
// by these tests and panics rather than pretending to work.
type fakeRuleStore struct {
	connectors.RuleRepository
	rule *domain.StoredRule
}

func (f fakeRuleStore) FindByRelPath(string) *domain.StoredRule { return f.rule }
