package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

// sqlSpy is a gorm logger that records every statement it traces — including
// ones built by Raw+Scan, which run through the same Trace call as everything
// else (see gorm's generic processor.Execute). That is what makes this a
// faithful check: it sees exactly what would reach Postgres, not what the
// tenancy callbacks merely attempted to add to a Statement that a Raw query
// had already serialized.
type sqlSpy struct{ queries []string }

func (s *sqlSpy) LogMode(gormlogger.LogLevel) gormlogger.Interface { return s }
func (s *sqlSpy) Info(context.Context, string, ...interface{})    {}
func (s *sqlSpy) Warn(context.Context, string, ...interface{})    {}
func (s *sqlSpy) Error(context.Context, string, ...interface{})   {}
func (s *sqlSpy) Trace(_ context.Context, _ time.Time, fc func() (string, int64), _ error) {
	sql, _ := fc()
	s.queries = append(s.queries, sql)
}

func newSpyDB(t *testing.T) (*gorm.DB, *sqlSpy) {
	t.Helper()
	spy := &sqlSpy{}
	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{DryRun: true, Logger: spy})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	if err := tenancy.Register(db, func() bool { return true }); err != nil {
		t.Fatalf("tenancy.Register: %v", err)
	}
	return db, spy
}

// The bug this guards: List used to build its WHERE clause as a hand-rolled
// SQL string and run it through db.Raw(...).Scan(...). Raw fully serializes
// the query text up front, so the tenancy plugin's Row callback — which adds
// its predicate to the Statement, not to already-built SQL — had nothing left
// to attach to, and every tenant's tagging rules came back for anyone who
// asked.
func TestTagRuleListIsScopedToTheCallersTenant(t *testing.T) {
	db, spy := newSpyDB(t)
	repo := NewAlertTagRuleRepository(db)

	ctx := authz.WithTenantID(context.Background(), "tenant-a")
	if _, _, err := repo.List(ctx, dto.AlertTagRuleFilters{}); err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(spy.queries) == 0 {
		t.Fatal("no SQL was traced; the test isn't observing the real query")
	}
	for _, sql := range spy.queries {
		if !strings.Contains(sql, "alert_tag_rule") {
			continue
		}
		if !strings.Contains(sql, "tenant_id") {
			t.Errorf("query against alert_tag_rule has no tenant_id predicate: %s", sql)
		}
		if !strings.Contains(sql, "tenant-a") {
			t.Errorf("query against alert_tag_rule doesn't bind the caller's tenant: %s", sql)
		}
	}
}

// A request with no tenant in context must fail rather than silently list
// every tenant's rules — same fail-closed contract every other tenant-scoped
// repository gets from the plugin, now that List goes through it.
func TestTagRuleListWithoutTenantFailsClosed(t *testing.T) {
	db, _ := newSpyDB(t)
	repo := NewAlertTagRuleRepository(db)

	if _, _, err := repo.List(context.Background(), dto.AlertTagRuleFilters{}); err == nil {
		t.Fatal("List without a tenant in context returned no error; it would have listed every tenant's rules")
	}
}
