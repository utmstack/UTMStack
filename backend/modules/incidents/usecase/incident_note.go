package usecase

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
)

type incidentNoteUsecase struct {
	noteRepo    connectors.IncidentNoteRepository
	historyRepo connectors.IncidentHistoryRepository
}

func NewIncidentNoteUsecase(
	noteRepo connectors.IncidentNoteRepository,
	historyRepo connectors.IncidentHistoryRepository,
) connectors.IncidentNoteUsecase {
	return &incidentNoteUsecase{
		noteRepo:    noteRepo,
		historyRepo: historyRepo,
	}
}

func (u *incidentNoteUsecase) Create(ctx context.Context, userLogin string, req dto.CreateNoteRequest) (*domain.UtmIncidentNote, error) {
	currentUser := resolveUser(userLogin)

	now := time.Now().UTC()
	note := &domain.UtmIncidentNote{
		IncidentID:   req.IncidentID,
		NoteText:     req.NoteText,
		NoteSendDate: now,
		NoteSendBy:   &currentUser,
	}
	if err := u.noteRepo.Save(ctx, note); err != nil {
		return nil, err
	}

	detail := "New note added to incident"
	by := currentUser
	h := &domain.UtmIncidentHistory{
		IncidentID:        req.IncidentID,
		Action:            domain.ActionNoteAdd.Label,
		ActionType:        domain.ActionNoteAdd.Type,
		ActionDetail:      &detail,
		ActionCreatedDate: now,
		ActionCreatedBy:   &by,
	}
	if err := u.historyRepo.Save(ctx, h); err != nil {
		catcher.Warn("incidents: failed to write note history", map[string]any{"error": err.Error()})
	}

	return note, nil
}

func (u *incidentNoteUsecase) List(ctx context.Context, query dto.IncidentNoteListQuery) ([]domain.UtmIncidentNote, int64, error) {
	return u.noteRepo.FindAll(ctx, query)
}
