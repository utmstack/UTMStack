package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

// These reads are written as SQL, so the boundary that matters is what ends up
// in the statement text: everything a caller supplies has to arrive as a bound
// parameter instead.
func TestNothingTheCallerSuppliedReachesTheStatement(t *testing.T) {
	r := &chIngestionStatsRepository{}
	ctx := authz.WithTenantID(context.Background(), "8f1c1b8e-0000-4000-8000-00000000000a")

	where, args, err := r.predicate(ctx, connectors.IngestionStatsQuery{
		From:       time.Now().Add(-time.Hour),
		To:         time.Now(),
		Type:       "enqueue_success'; DROP TABLE statistics; --",
		DataSource: "fw-01' OR 1=1 --",
	})
	if err != nil {
		t.Fatalf("predicate: %v", err)
	}

	for _, injected := range []string{"DROP TABLE", "OR 1=1", "enqueue_success", "fw-01"} {
		if strings.Contains(where, injected) {
			t.Errorf("%q reached the statement: %s", injected, where)
		}
	}
	if len(args) != 5 {
		t.Errorf("got %d bound parameters, want tenant + both bounds + type + source", len(args))
	}
	if !strings.Contains(where, "tenantId = ?") {
		t.Errorf("the tenant is not part of the predicate: %s", where)
	}
	if args[0] != "8f1c1b8e-0000-4000-8000-00000000000a" {
		t.Errorf("the first parameter is %v, want the caller's tenant", args[0])
	}
}

// A read with no tenant would be a read of every tenant's ingestion.
func TestWithoutATenantItRefusesOrReadsItsOwnInstall(t *testing.T) {
	r := &chIngestionStatsRepository{}
	_, _, err := r.predicate(context.Background(), connectors.IngestionStatsQuery{})

	if tenancy.Enabled() {
		if err == nil {
			t.Error("it answered without a tenant while multitenancy is on")
		}
		return
	}
	if err != nil {
		t.Fatalf("single-tenant install was refused: %v", err)
	}
}

// The bucket size is the one thing that cannot be a bound parameter, so it is
// built from a whitelist rather than from the request.
func TestTheBucketSizeIsNeverTheCallersText(t *testing.T) {
	if got := bucketExpr("1h"); got != "toStartOfInterval(`@timestamp`, INTERVAL 1 HOUR)" {
		t.Errorf("1h -> %s", got)
	}
	if got := bucketExpr("15m"); !strings.Contains(got, "INTERVAL 15 MINUTE") {
		t.Errorf("15m -> %s", got)
	}
	if got := bucketExpr("7d"); !strings.Contains(got, "INTERVAL 7 DAY") {
		t.Errorf("7d -> %s", got)
	}
	if got := bucketExpr("1M"); !strings.Contains(got, "INTERVAL 1 MONTH") {
		t.Errorf("1M -> %s", got)
	}

	for _, hostile := range []string{"1 HOUR) FROM statistics --", "", "abc", "0h", "-1d", "1y"} {
		got := bucketExpr(hostile)
		if !strings.Contains(got, "INTERVAL 1 HOUR)") || strings.Contains(got, hostile) && hostile != "" {
			t.Errorf("%q -> %s, want the hourly fallback", hostile, got)
		}
	}
}

// Ingestion can be broken down by where it came from or what kind of record it
// was, and by nothing else — the column is spliced into the statement.
func TestOnlyTheTwoKnownBreakdownsAreAccepted(t *testing.T) {
	r := &chIngestionStatsRepository{}
	ctx := authz.WithTenantID(context.Background(), "8f1c1b8e-0000-4000-8000-00000000000a")

	for _, field := range []string{"tenantId", "count", "`@timestamp`", "1, (SELECT 1)", ""} {
		if _, _, err := r.TotalsByField(ctx, field, connectors.IngestionStatsQuery{}, 10); err == nil {
			t.Errorf("grouping by %q was accepted", field)
		}
	}
}
