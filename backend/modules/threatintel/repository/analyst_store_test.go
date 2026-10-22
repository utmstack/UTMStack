package repository

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/utils/tests"

	"github.com/utmstack/utmstack/backend/modules/threatintel/domain"
)

// DryRun: exercises the query-building and tenancy scoping without a database.
// The Where clauses are what enforce isolation; a dry run still renders and
// validates them.
func TestAnalystStoreIsolationAndSaveDryRun(t *testing.T) {
	db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	s := NewAnalystStore(db)
	ctx := context.Background()
	conv := &domain.AnalystConversation{
		ID: "conv_a1", TenantID: "t1", UserID: "u1",
		Title: "First", Turns: `[{"id":"t1"}]`, History: "[]",
		UpdatedAt: time.Now(),
	}

	if err := s.Create(ctx, conv); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := s.List(ctx, "t1", "u1"); err != nil {
		t.Fatalf("list: %v", err)
	}
	if _, err := s.Get(ctx, "t1", "u1", "conv_a1"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if err := s.Save(ctx, conv); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := s.Delete(ctx, "t1", "u1", "conv_a1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
