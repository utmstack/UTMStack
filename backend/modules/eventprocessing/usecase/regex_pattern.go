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

// regexPatternUsecase serves the shared regex vocabulary read-only. patterns.yaml
// is the engine's source of truth — it is what loadCfg merges into Config.Patterns
// and what RegexpCache expands {{.name}} references from — so this type reads that
// file and never writes it.
type regexPatternUsecase struct {
	writer *pipelineWriter // read-only use: ReadPatterns
}

func NewRegexPatternUsecase(writer *pipelineWriter) connectors.RegexPatternUsecase {
	return &regexPatternUsecase{writer: writer}
}

// definitions returns the patterns the engine will actually resolve: the built-in
// vocabulary overlaid with whatever patterns.yaml holds on disk.
//
// Read per call rather than cached at construction, for two reasons. The pipeline
// bootstrap writes patterns.yaml during Module.Start, which runs after this type
// is built — a constructor-time read would see a fresh install's missing file and
// serve the built-ins forever. And an operator replacing the file underneath a
// running process should be reflected here rather than silently ignored. The file
// holds a few dozen entries and these endpoints are admin-frequency, so the read
// costs nothing worth caching.
//
// A read failure degrades to the built-ins instead of failing the request: those
// are always present in the engine too, so a partial answer beats none.
func (u *regexPatternUsecase) definitions() map[string]string {
	defs := make(map[string]string, len(systemPatterns))
	for id, def := range systemPatterns {
		defs[id] = def
	}

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

// response builds the API shape for one pattern. PatternDescription is always
// empty: patterns.yaml carries only name → definition, since that is all the
// engine needs, and there is no longer a write path that could supply more.
func response(patternID, definition string) dto.RegexPatternResponse {
	_, system := systemPatterns[patternID]
	return dto.RegexPatternResponse{
		PatternID:         patternID,
		PatternDefinition: definition,
		SystemOwner:       system,
	}
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
		e := response(id, def)
		if f.System != nil && e.SystemOwner != *f.System {
			continue
		}
		out = append(out, e)
	}

	// Stable order: system first, then alphabetical by patternId.
	sort.Slice(out, func(i, j int) bool {
		if out[i].SystemOwner != out[j].SystemOwner {
			return out[i].SystemOwner
		}
		return out[i].PatternID < out[j].PatternID
	})

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
