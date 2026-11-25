package connectors

import (
	"context"
	"time"
)

// FloodBucket is how many open, un-deduplicated alerts one rule raised for one tenant from one data source.
type FloodBucket struct {
	TenantID   string
	RuleName   string
	DataSource string
	Count      int64
}

// FloodRepository counts the alerts the flood guard looks at.
type FloodRepository interface {
	// FloodBuckets groups the alerts of the whole instance raised within window by tenant, rule and
	// data source. The tenant is part of the grouping: it is what keeps one tenant's volume from
	// adding to another's.
	FloodBuckets(ctx context.Context, window time.Duration) ([]FloodBucket, error)
}

// FloodRule is a correlation rule as the flood guard needs to see it.
type FloodRule struct {
	RelPath string
	Name    string
	Active  bool
}

// FloodRuleSwitch finds and disables the correlation rules of a tenant.
type FloodRuleSwitch interface {
	// FindByName returns the tenant's rules whose name matches name, ignoring case.
	FindByName(ctx context.Context, tenantID, name string) ([]FloodRule, error)
	// Deactivate turns the rule off and says whether this call is the one that changed it.
	Deactivate(ctx context.Context, tenantID, relPath string) (bool, error)
}

// FloodNotifier tells a tenant, in its notification bell, that a rule was disabled.
type FloodNotifier interface {
	Warn(ctx context.Context, tenantID, message string) error
}
