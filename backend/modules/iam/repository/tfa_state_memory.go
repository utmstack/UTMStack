package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
)

type memTfaStateRepository struct {
	mu      sync.Mutex
	entries map[string]*domain.TfaSetupState
	ttl     time.Duration
}

func NewInMemoryTfaStateRepository(ttl time.Duration) connectors.TfaStateRepository {
	return &memTfaStateRepository{
		entries: make(map[string]*domain.TfaSetupState),
		ttl:     ttl,
	}
}

func tfaKey(purpose string, userID uint64, method string) string {
	return fmt.Sprintf("%s:%d:%s", purpose, userID, method)
}

func (r *memTfaStateRepository) Get(_ context.Context, purpose string, userID uint64, method string) (*domain.TfaSetupState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := tfaKey(purpose, userID, method)
	entry, ok := r.entries[key]
	if !ok {
		return nil, nil
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		delete(r.entries, key)
		return nil, nil
	}
	clone := *entry
	return &clone, nil
}

func (r *memTfaStateRepository) Put(_ context.Context, state *domain.TfaSetupState) error {
	if state == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if state.ExpiresAt.IsZero() {
		state.ExpiresAt = time.Now().Add(r.ttl)
	}
	clone := *state
	r.entries[tfaKey(state.Purpose, state.UserID, state.Method)] = &clone
	return nil
}

func (r *memTfaStateRepository) Delete(_ context.Context, purpose string, userID uint64, method string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, tfaKey(purpose, userID, method))
	return nil
}

func (r *memTfaStateRepository) MarkVerified(_ context.Context, purpose string, userID uint64, method string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := tfaKey(purpose, userID, method)
	entry, ok := r.entries[key]
	if !ok {
		return nil
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		delete(r.entries, key)
		return nil
	}
	entry.Verified = true
	return nil
}
