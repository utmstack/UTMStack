package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
)

type fakeVariableRepo struct {
	byID map[uuid.UUID]*domain.SoarVariable
	saved *domain.SoarVariable
}

func (f *fakeVariableRepo) Save(_ context.Context, v *domain.SoarVariable) error {
	f.saved = v
	f.byID[v.ID] = v
	return nil
}
func (f *fakeVariableRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.SoarVariable, error) {
	v, ok := f.byID[id]
	if !ok {
		return nil, domain.ErrVariableNotFound
	}
	cp := *v
	return &cp, nil
}
func (f *fakeVariableRepo) FindAll(context.Context, dto.VariableFilter) ([]domain.SoarVariable, int64, error) {
	return nil, 0, nil
}
func (f *fakeVariableRepo) FindAllPlain(context.Context) ([]domain.SoarVariable, error) {
	out := make([]domain.SoarVariable, 0, len(f.byID))
	for _, v := range f.byID {
		out = append(out, *v)
	}
	return out, nil
}
func (f *fakeVariableRepo) FindByName(context.Context, string) (*domain.SoarVariable, error) { return nil, nil }
func (f *fakeVariableRepo) FindByNames(context.Context, []string) ([]domain.SoarVariable, error) {
	return nil, nil
}
func (f *fakeVariableRepo) Delete(context.Context, uuid.UUID) error { return nil }

// fakeCipher is a reversible stand-in: Encrypt prefixes, Decrypt strips it.
type fakeCipher struct{}

func (fakeCipher) Encrypt(plaintext string) (string, error) { return "enc:" + plaintext, nil }
func (fakeCipher) Decrypt(ciphertext string) (string, error) {
	if len(ciphertext) < 4 || ciphertext[:4] != "enc:" {
		return "", errors.New("not encrypted")
	}
	return ciphertext[4:], nil
}

// SOAR-F4: flipping IsSecret off without supplying a new value must not hand
// back the old secret in clear.
func TestUpdateCannotUnsecretWithoutANewValue(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC()
	repo := &fakeVariableRepo{byID: map[uuid.UUID]*domain.SoarVariable{
		id: {ID: id, Name: "api_token", Value: "enc:top-secret", IsSecret: true, CreatedAt: now},
	}}
	u := NewVariableUsecase(repo, fakeCipher{})

	_, err := u.Update(context.Background(), dto.UpdateVariableRequest{ID: id, IsSecret: false}, "tester")
	if !errors.Is(err, domain.ErrVariableUnsecretNeedsValue) {
		t.Fatalf("got %v, want ErrVariableUnsecretNeedsValue", err)
	}
	if repo.saved != nil {
		t.Fatalf("variable must not have been saved, got %+v", repo.saved)
	}
}

// Supplying an explicit new value while unsecreting is a legitimate way to
// replace a secret with a plain one — must still work.
func TestUpdateCanUnsecretWithAnExplicitNewValue(t *testing.T) {
	id := uuid.New()
	repo := &fakeVariableRepo{byID: map[uuid.UUID]*domain.SoarVariable{
		id: {ID: id, Name: "api_token", Value: "enc:top-secret", IsSecret: true},
	}}
	u := NewVariableUsecase(repo, fakeCipher{})

	newVal := "now-public"
	resp, err := u.Update(context.Background(), dto.UpdateVariableRequest{ID: id, IsSecret: false, Value: &newVal}, "tester")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.Value != newVal {
		t.Fatalf("response value = %q, want %q", resp.Value, newVal)
	}
	if repo.saved.Value != newVal {
		t.Fatalf("stored value = %q, want %q (plain, since IsSecret=false)", repo.saved.Value, newVal)
	}
}

// Editing a secret variable's other fields while keeping it secret must still
// work (decrypt-then-reencrypt is fine as long as the result stays secret).
func TestUpdateKeepingSecretStillWorks(t *testing.T) {
	id := uuid.New()
	desc := "new description"
	repo := &fakeVariableRepo{byID: map[uuid.UUID]*domain.SoarVariable{
		id: {ID: id, Name: "api_token", Value: "enc:top-secret", IsSecret: true},
	}}
	u := NewVariableUsecase(repo, fakeCipher{})

	resp, err := u.Update(context.Background(), dto.UpdateVariableRequest{ID: id, IsSecret: true, Description: &desc}, "tester")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.Value != "****" {
		t.Fatalf("response value = %q, want masked", resp.Value)
	}
	if repo.saved.Value != "enc:top-secret" {
		t.Fatalf("stored value = %q, want unchanged ciphertext", repo.saved.Value)
	}
}

// ACC-F3-shaped bug class for SOAR-F3: a secret must stay masked in the
// execution-history line for every origin, not just flow-triggered rows —
// the viewer role and the MCP mirror both read through this same path.
func TestCommandSummaryMasksManualOriginSecrets(t *testing.T) {
	repo := &fakeVariableRepo{byID: map[uuid.UUID]*domain.SoarVariable{
		uuid.New(): {Name: "api_token", Value: "enc:top-secret", IsSecret: true},
	}}
	vars := NewVariableUsecase(repo, fakeCipher{})

	e := &domain.SoarExecution{Origin: domain.ExecutionOriginManual, Command: "curl -H 'Authorization: top-secret' https://x"}
	CommandSummary(context.Background(), vars, e)

	if e.Command == "curl -H 'Authorization: top-secret' https://x" {
		t.Fatalf("manual-origin execution's Command was not masked: %q", e.Command)
	}
	if !contains(e.Command, "****") {
		t.Fatalf("expected the secret to be masked, got %q", e.Command)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
