package usecase

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type adversaryUsecase struct {
	repo connectors.AdversaryRepository
}

func NewAdversaryUsecase(repo connectors.AdversaryRepository) connectors.AdversaryUsecase {
	return &adversaryUsecase{repo: repo}
}

func (u *adversaryUsecase) FetchAdversaryAlerts(
	ctx context.Context,
	filters []common_models.FilterType,
) ([]dto.AdversaryResponse, error) {
	groups, err := u.repo.AdversaryGroups(ctx, filters)
	if err != nil {
		return nil, err
	}

	out := make([]dto.AdversaryResponse, 0, len(groups))
	for _, g := range groups {
		if resp, ok := buildAdversaryGroup(g); ok {
			out = append(out, resp)
		}
	}
	return out, nil
}

// buildAdversaryGroup turns one attacker's sample of alerts into the parent and
// echo shape the drawer renders. A group whose alerts are all echoes has no
// parent to hang them from and is dropped, not shown headless.
func buildAdversaryGroup(g connectors.AdversaryGroup) (dto.AdversaryResponse, bool) {
	alerts := make([]domain.UtmAlert, 0, len(g.Alerts))
	for _, raw := range g.Alerts {
		var a domain.UtmAlert
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		alerts = append(alerts, a)
	}
	if len(alerts) == 0 {
		return dto.AdversaryResponse{}, false
	}

	// The store returns a group's sample in no particular order, so newest-first
	// is restored here: it decides which adversary description is shown and the
	// order the alerts are listed in.
	sort.SliceStable(alerts, func(i, j int) bool {
		return alerts[i].Timestamp > alerts[j].Timestamp
	})

	children := make(map[string][]domain.UtmAlert)
	var parents []domain.UtmAlert
	for _, a := range alerts {
		if a.ParentID == "" {
			parents = append(parents, a)
			continue
		}
		children[a.ParentID] = append(children[a.ParentID], a)
	}
	if len(parents) == 0 {
		return dto.AdversaryResponse{}, false
	}

	withChildren := make([]dto.AlertWithChildren, 0, len(parents))
	for _, p := range parents {
		kids := children[p.ID]
		if kids == nil {
			kids = []domain.UtmAlert{}
		}
		withChildren = append(withChildren, dto.AlertWithChildren{Alert: p, Children: kids})
	}

	// The adversary is described by whichever alert saw it most recently.
	var side *domain.Side
	for _, a := range alerts {
		if a.Adversary != nil {
			side = a.Adversary
			break
		}
	}

	return dto.AdversaryResponse{Adversary: side, Alerts: withChildren}, true
}
