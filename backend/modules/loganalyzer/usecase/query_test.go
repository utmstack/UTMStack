package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
)

type stubRepo struct {
	existing *domain.SavedQuery
	saved    *domain.SavedQuery
}

func (s *stubRepo) Save(_ context.Context, q *domain.SavedQuery) error {
	s.saved = q
	return nil
}

func (s *stubRepo) FindByID(context.Context, uint64) (*domain.SavedQuery, error) {
	return s.existing, nil
}

func (s *stubRepo) List(context.Context, dto.QueryFilter) ([]domain.SavedQuery, int64, error) {
	return nil, 0, nil
}

func (s *stubRepo) Delete(context.Context, uint64) error { return nil }

// The update path writes the whole row from the request body, which carries no
// tenant (the field is json:"-") and no creation time. Both have to survive the
// round trip: reads are tenant-scoped, so a blanked tenant does not corrupt the
// query, it hides it from the person who just edited it.
func TestUpdateKeepsWhatTheRequestCannotCarry(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	repo := &stubRepo{existing: &domain.SavedQuery{
		ID:        7,
		TenantID:  "8f1c1b8e-0000-4000-8000-000000000001",
		Owner:     "someone@example.com",
		CreatedAt: created,
		Name:      "before",
	}}
	u := NewQueryUsecase(repo)

	incoming := &domain.SavedQuery{ID: 7, Name: "after", Dataset: "logs"}
	if _, err := u.Update(context.Background(), incoming, "editor@example.com"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if repo.saved.TenantID != repo.existing.TenantID {
		t.Errorf("tenant = %q, want %q", repo.saved.TenantID, repo.existing.TenantID)
	}
	if !repo.saved.CreatedAt.Equal(created) {
		t.Errorf("createdAt = %v, want %v", repo.saved.CreatedAt, created)
	}
	// The owner is the one thing the body may override; left empty it stays.
	if repo.saved.Owner != "someone@example.com" {
		t.Errorf("owner = %q, want the original", repo.saved.Owner)
	}
	if repo.saved.Name != "after" {
		t.Errorf("name = %q, want the edit to apply", repo.saved.Name)
	}
}
