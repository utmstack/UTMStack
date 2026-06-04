package usecase

import (
	"context"
	"sort"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/domain"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
	lsErrors "github.com/utmstack/utmstack/backend/modules/logstash/errors"
	"github.com/utmstack/utmstack/backend/modules/logstash/repository"
	"gorm.io/gorm"
)

const (
	pipelineStatusUp   = "up"
	pipelineStatusDown = "down"

	engineStatusGreen  = "green"
	engineStatusYellow = "yellow"
	engineStatusRed    = "red"
)

type PipelineUsecase struct {
	db                 *gorm.DB
	pipelineRepo       connectors.PipelineRepository
	pipelineFilterRepo connectors.PipelineFilterRepository
	filterRepo         connectors.FilterRepository
}

func NewPipelineUsecase(
	db *gorm.DB,
	pipelineRepo connectors.PipelineRepository,
	pipelineFilterRepo connectors.PipelineFilterRepository,
	filterRepo connectors.FilterRepository,
) connectors.PipelineUsecase {
	return &PipelineUsecase{
		db:                 db,
		pipelineRepo:       pipelineRepo,
		pipelineFilterRepo: pipelineFilterRepo,
		filterRepo:         filterRepo,
	}
}

func (uc *PipelineUsecase) List(ctx context.Context, page, size int) ([]dto.UtmLogstashPipelineDTO, int64, error) {
	all, err := uc.pipelineRepo.AllActivePipelinesByServer(ctx)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))

	start := page * size
	if start >= len(all) {
		return []dto.UtmLogstashPipelineDTO{}, total, nil
	}
	end := start + size
	if end > len(all) {
		end = len(all)
	}
	slice := all[start:end]

	result := make([]dto.UtmLogstashPipelineDTO, 0, len(slice))
	for _, p := range slice {
		result = append(result, dto.PipelineDTOFromDomain(p))
	}
	return result, total, nil
}

func (uc *PipelineUsecase) GetStats(ctx context.Context) (*dto.ApiStatsResponse, error) {
	isEngineUp := true // Java stub — always true

	// Initialize with Java constructor defaults: version="2.0.0", status="down".
	engineResp := &dto.ApiEngineResponse{
		Version: "2.0.0",
		Status:  pipelineStatusDown,
	}

	activePipelines, err := uc.pipelineRepo.AllActivePipelinesByServer(ctx)
	if err != nil {
		return nil, err
	}

	if !isEngineUp {
		down := pipelineStatusDown
		for i := range activePipelines {
			activePipelines[i].PipelineStatus = &down
		}
	}

	pipelineStats := make([]dto.PipelineStats, 0, len(activePipelines))
	for _, p := range activePipelines {
		ps := toPipelineStats(p)
		pipelineStats = append(pipelineStats, ps)
	}

	// Sort pipeline stats by status descending (mirrors Java Comparator.comparing(PipelineStats::getPipelineStatus).reversed()).
	// Alphabetically descending: "up" sorts before "down".
	sort.Slice(pipelineStats, func(i, j int) bool {
		si := ""
		sj := ""
		if pipelineStats[i].PipelineStatus != nil {
			si = *pipelineStats[i].PipelineStatus
		}
		if pipelineStats[j].PipelineStatus != nil {
			sj = *pipelineStats[j].PipelineStatus
		}
		return si > sj
	})

	// Compute engine status from UP pipeline count
	if isEngineUp {
		var upCount int64
		for _, p := range activePipelines {
			if p.PipelineStatus != nil && *p.PipelineStatus == pipelineStatusUp {
				upCount++
			}
		}
		total := int64(len(activePipelines))
		switch {
		case upCount == 0:
			engineResp.Status = engineStatusRed
		case upCount == total:
			engineResp.Status = engineStatusGreen
		default:
			engineResp.Status = engineStatusYellow
		}
	}

	return &dto.ApiStatsResponse{
		General:   engineResp,
		Pipelines: pipelineStats,
	}, nil
}

func (uc *PipelineUsecase) GetByID(ctx context.Context, id int64) (*dto.UtmLogstashPipelineVM, error) {
	p, err := uc.pipelineRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, lsErrors.ErrPipelineNotFound
	}

	filters, err := uc.pipelineFilterRepo.GetFilters(ctx, int32(p.ID))
	if err != nil {
		return nil, err
	}

	return &dto.UtmLogstashPipelineVM{
		Pipeline: dto.PipelineDTOFromDomain(*p),
		Filters:  filters,
	}, nil
}

func (uc *PipelineUsecase) Validate(ctx context.Context, vm dto.UtmLogstashPipelineVM, mode string) (*dto.PipelineErrors, error) {
	modeUpper := strings.ToUpper(strings.TrimSpace(mode))
	if modeUpper != dto.PipelineValidationModeInsert && modeUpper != dto.PipelineValidationModeUpdate {
		return nil, &invalidModeError{mode: mode}
	}

	var errs []dto.Validation
	p := vm.Pipeline
	filters := vm.Filters

	// Common validations (INSERT + UPDATE)
	pipelineName := ""
	if p.PipelineName != nil {
		pipelineName = strings.TrimSpace(*p.PipelineName)
	}
	if pipelineName == "" {
		errs = append(errs, dto.Validation{
			Entity: "Pipeline",
			Field:  "Pipeline name",
			Msg:    "Value is null or empty",
		})
	}

	if len(filters) == 0 {
		name := "Undefined pipeline"
		if pipelineName != "" {
			name = pipelineName
		}
		errs = append(errs, dto.Validation{
			Entity: "Filter relation",
			Field:  "Filter id",
			Msg:    "There is no filter associated to the pipeline: " + name,
		})
	} else {
		for _, f := range filters {
			if f.FilterID == 0 {
				errs = append(errs, dto.Validation{
					Entity: "Filter",
					Field:  "Filter id",
					Msg:    "Value is null",
				})
				continue
			}
			// Verify filter exists
			existing, err := uc.filterRepo.GetByID(ctx, int64(f.FilterID))
			if err != nil {
				return nil, err
			}
			if existing == nil {
				errs = append(errs, dto.Validation{
					Entity: "Filter",
					Field:  "Filter id",
					Msg:    "Value " + int32ToString(f.FilterID) + " not exist",
				})
			}
			if modeUpper == dto.PipelineValidationModeInsert && f.ID != 0 {
				errs = append(errs, dto.Validation{
					Entity: "Filter relation",
					Field:  "Relation id",
					Msg:    "Value must be null when inserting",
				})
			}
		}
	}

	// INSERT-only: pipeline ID must be zero/nil
	if modeUpper == dto.PipelineValidationModeInsert && p.ID != 0 {
		errs = append(errs, dto.Validation{
			Entity: "Pipeline",
			Field:  "Pipeline id",
			Msg:    "Value must be null when inserting",
		})
	}

	// UPDATE-only: pipeline ID must be non-zero
	if modeUpper == dto.PipelineValidationModeUpdate && p.ID == 0 {
		errs = append(errs, dto.Validation{
			Entity: "Pipeline",
			Field:  "Pipeline id",
			Msg:    "Value can't be null when updating",
		})
	}

	if len(errs) == 0 {
		return nil, nil
	}
	return &dto.PipelineErrors{ValidationErrors: errs}, nil
}

func (uc *PipelineUsecase) Delete(ctx context.Context, id int64) error {
	// Check existence and system ownership before opening the transaction.
	p, err := uc.pipelineRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if p == nil {
		return lsErrors.ErrPipelineNotFound
	}
	if p.SystemOwner {
		return lsErrors.ErrPipelineSystemOwner
	}

	return uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create tx-scoped repos so every operation inside the transaction uses the same tx.
		txPipelineFilterRepo := repository.NewPipelineFilterRepository(tx)
		txFilterRepo := repository.NewFilterRepository(tx)
		txPipelineRepo := repository.NewPipelineRepository(tx)

		// Step 1: get filter relations
		relations, err := txPipelineFilterRepo.GetFilters(ctx, int32(p.ID))
		if err != nil {
			return err
		}

		// Step 2: delete only relations belonging to THIS pipeline
		if err := txPipelineFilterRepo.DeleteRelationsByPipelineID(ctx, int32(p.ID)); err != nil {
			return err
		}

		// Step 3: delete non-system filters
		for _, rel := range relations {
			f, err := txFilterRepo.GetByID(ctx, int64(rel.FilterID))
			if err != nil {
				return err
			}
			if f != nil && !f.SystemOwner {
				if err := txFilterRepo.Delete(ctx, f.ID); err != nil {
					return err
				}
			}
		}

		// Step 4: delete pipeline
		return txPipelineRepo.DeletePipeline(ctx, id)
	})
}

func toPipelineStats(p domain.UtmLogstashPipeline) dto.PipelineStats {
	name := ""
	if p.PipelineName != nil {
		name = *p.PipelineName
	}
	return dto.PipelineStats{
		ID:                  p.ID,
		PipelineID:          p.PipelineID,
		PipelineName:        name,
		PipelineStatus:      p.PipelineStatus,
		ModuleName:          p.ModuleName,
		SystemOwner:         p.SystemOwner,
		PipelineDescription: p.PipelineDescription,
		Events:              dto.PipelineEvents{Out: p.EventsOut},
		Errors:              0,
	}
}

func int32ToString(n int32) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 12)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

type invalidModeError struct{ mode string }

func (e *invalidModeError) Error() string {
	return "The value of mode that you provide is wrong, only INSERT or UPDATE are allowed: " + e.mode
}
