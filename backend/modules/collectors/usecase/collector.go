package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/collectors/connectors"
	"github.com/utmstack/utmstack/backend/modules/collectors/domain"
	"github.com/utmstack/utmstack/backend/modules/collectors/dto"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type collectorUsecase struct {
	repo      connectors.CollectorRepository
	agentRepo agentManagerRepo
}

type agentManagerRepo interface {
	ListCollectors(ctx context.Context, searchQuery string, pageNumber, pageSize int32, sortBy string) (*agent.ListCollectorResponse, error)
}

func NewCollectorUsecase(repo connectors.CollectorRepository, agentRepo agentManagerRepo) connectors.CollectorUsecase {
	return &collectorUsecase{repo: repo, agentRepo: agentRepo}
}

func (u *collectorUsecase) ListAndSync(ctx context.Context, q dto.CollectorListQuery) (*dto.ListCollectorsResponse, error) {
	parts := []string{}
	if q.Module != "" {
		parts = append(parts, fmt.Sprintf("module.Is=%s", q.Module))
	}
	if q.Hostname != "" {
		parts = append(parts, fmt.Sprintf("hostname.Is=%s", q.Hostname))
	}
	searchQuery := strings.Join(parts, "&")

	pageNumber := int32(q.PageNumber)
	if pageNumber <= 0 {
		pageNumber = 1
	}
	pageSize := int32(q.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	// C1: pass sortBy to gRPC request.
	resp, err := u.agentRepo.ListCollectors(ctx, searchQuery, pageNumber, pageSize, q.SortBy)
	if err != nil {
		return nil, fmt.Errorf("collectorUsecase.ListAndSync: %w", err)
	}

	// Sync side-effect: upsert + mark-offline. Response comes from gRPC slice (C2).
	activeIDs := make([]int64, 0, len(resp.GetRows()))
	dtos := make([]dto.CollectorResponse, 0, len(resp.GetRows()))
	for _, c := range resp.GetRows() {
		col := protoToCollector(c)
		if err := u.repo.Upsert(ctx, col); err != nil {
			return nil, fmt.Errorf("collectorUsecase.ListAndSync upsert: %w", err)
		}
		activeIDs = append(activeIDs, col.ID)
		dtos = append(dtos, collectorToDTO(col))
	}

	if err := u.repo.MarkOfflineExcept(ctx, activeIDs); err != nil {
		return nil, fmt.Errorf("collectorUsecase.ListAndSync mark offline: %w", err)
	}

	return &dto.ListCollectorsResponse{
		Collectors: dtos,
		Total:      resp.GetTotal(),
	}, nil
}

func (u *collectorUsecase) Delete(ctx context.Context, id int64) error {
	if _, err := u.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("collector %d not found: %w", id, ErrNotFound)
		}
		return fmt.Errorf("collector %d lookup: %w", id, err)
	}
	return u.repo.DeleteByID(ctx, id)
}

func parseLastSeen(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return &t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	return nil
}

func protoToCollector(c *agent.Collector) *domain.UtmCollector {
	return &domain.UtmCollector{
		ID:           int64(c.GetId()),
		Status:       c.GetStatus().String(),
		CollectorKey: c.GetCollectorKey(),
		IP:           c.GetIp(),
		Hostname:     c.GetHostname(),
		Version:      c.GetVersion(),
		Module:       c.GetModule().String(),
		LastSeen:     parseLastSeen(c.GetLastSeen()), // C3: correct format
		Active:       true,
	}
}

func collectorToDTO(c *domain.UtmCollector) dto.CollectorResponse {
	return dto.CollectorResponse{
		ID:           c.ID,
		Status:       c.Status,
		CollectorKey: c.CollectorKey,
		IP:           c.IP,
		Hostname:     c.Hostname,
		Version:      c.Version,
		Module:       c.Module,
		LastSeen:     c.LastSeen,
		GroupID:      c.GroupID,
		Active:       c.Active,
	}
}
