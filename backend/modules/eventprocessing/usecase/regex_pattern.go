package usecase

import (
	"context"
	"sort"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

type regexPatternUsecase struct {
	writer connectors.EngineConfigRepository // read-only use: ReadPatterns
}

func NewRegexPatternUsecase(writer connectors.EngineConfigRepository) connectors.RegexPatternUsecase {
	return &regexPatternUsecase{writer: writer}
}

// definitions is the vocabulary the engine will actually resolve: the built-ins
// overlaid with whatever patterns.yaml holds. A read failure degrades to the
// built-ins rather than failing the request — those are always present in the
// engine too, so a partial answer beats none.
func (u *regexPatternUsecase) definitions() map[string]string {
	defs := u.writer.BuiltInPatterns()

	onDisk, err := u.writer.ReadPatterns()
	if err != nil {
		_ = catcher.Error("failed to read patterns.yaml, serving built-in patterns only", err, nil)
		return defs
	}
	for id, def := range onDisk {
		defs[id] = def
	}
	return defs
}

func response(patternID, definition string) dto.RegexPatternResponse {
	return dto.RegexPatternResponse{PatternID: patternID, PatternDefinition: definition}
}

func (u *regexPatternUsecase) GetByID(_ context.Context, patternID string) (*dto.RegexPatternResponse, error) {
	definition, exists := u.definitions()[patternID]
	if !exists {
		return nil, domain.ErrRegexPatternNotFound
	}
	out := response(patternID, definition)
	return &out, nil
}

func (u *regexPatternUsecase) List(_ context.Context, f dto.RegexPatternFilters) (*connectors.ListResult[dto.RegexPatternResponse], error) {
	defs := u.definitions()
	search := strings.ToLower(f.Search)

	out := make([]dto.RegexPatternResponse, 0, len(defs))
	for id, def := range defs {
		if search != "" && !strings.Contains(strings.ToLower(id), search) {
			continue
		}
		out = append(out, response(id, def))
	}

	sort.Slice(out, func(i, j int) bool { return out[i].PatternID < out[j].PatternID })

	total := int64(len(out))
	page, size := normPage(f.Page, f.Size)
	start := page * size
	if start >= len(out) {
		return &connectors.ListResult[dto.RegexPatternResponse]{Items: []dto.RegexPatternResponse{}, Total: total}, nil
	}
	end := start + size
	if end > len(out) {
		end = len(out)
	}
	return &connectors.ListResult[dto.RegexPatternResponse]{Items: out[start:end], Total: total}, nil
}

func normPage(page, size int) (int, int) {
	if page < 0 {
		page = 0
	}
	if size < 1 || size > 200 {
		size = 20
	}
	return page, size
}
