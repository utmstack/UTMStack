package usecase

import (
	"context"
	"sort"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

type regexPatternUsecase struct {
	writer *pipelineWriter
	store  *patternStore // in-memory patterns (system + user)
}

func NewRegexPatternUsecase(writer *pipelineWriter) connectors.RegexPatternUsecase {
	uc := &regexPatternUsecase{
		writer: writer,
		store:  newPatternStore(),
	}
	// Seed the built-in system patterns at construction time.
	for k, v := range systemPatterns {
		uc.store.setSystem(k, v)
	}
	return uc
}

func (u *regexPatternUsecase) Create(_ context.Context, req dto.CreateRegexPatternRequest) (*dto.RegexPatternResponse, error) {
	if _, exists := u.store.get(req.PatternID); exists {
		return nil, domain.ErrIDMustBeAbsent // pattern already exists
	}
	u.store.setUser(req.PatternID, req.PatternDescription, req.PatternDefinition)
	_ = u.writer.WritePatterns(u.store.allDefinitions())
	return &dto.RegexPatternResponse{
		PatternID:          req.PatternID,
		PatternDescription: req.PatternDescription,
		PatternDefinition:  req.PatternDefinition,
		SystemOwner:        false,
	}, nil
}

func (u *regexPatternUsecase) Update(_ context.Context, req dto.UpdateRegexPatternRequest) (*dto.RegexPatternResponse, error) {
	entry, exists := u.store.get(req.PatternID)
	if !exists {
		return nil, domain.ErrRegexPatternNotFound
	}
	if entry.system {
		return nil, domain.ErrRegexPatternSystemOwner
	}
	u.store.setUser(req.PatternID, req.PatternDescription, req.PatternDefinition)
	_ = u.writer.WritePatterns(u.store.allDefinitions())
	return &dto.RegexPatternResponse{
		PatternID:          req.PatternID,
		PatternDescription: req.PatternDescription,
		PatternDefinition:  req.PatternDefinition,
		SystemOwner:        false,
	}, nil
}

func (u *regexPatternUsecase) GetByID(_ context.Context, patternID string) (*dto.RegexPatternResponse, error) {
	entry, exists := u.store.get(patternID)
	if !exists {
		return nil, domain.ErrRegexPatternNotFound
	}
	return &dto.RegexPatternResponse{
		PatternID:          patternID,
		PatternDescription: entry.description,
		PatternDefinition:  entry.definition,
		SystemOwner:        entry.system,
	}, nil
}

func (u *regexPatternUsecase) List(_ context.Context, f dto.RegexPatternFilters) (*connectors.ListResult[dto.RegexPatternResponse], error) {
	all := u.store.list()
	search := strings.ToLower(f.Search)
	out := make([]dto.RegexPatternResponse, 0, len(all))
	for _, e := range all {
		if search != "" && !strings.Contains(strings.ToLower(e.PatternID), search) {
			continue
		}
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

func (u *regexPatternUsecase) Delete(_ context.Context, patternID string) error {
	entry, exists := u.store.get(patternID)
	if !exists {
		return domain.ErrRegexPatternNotFound
	}
	if entry.system {
		return domain.ErrRegexPatternSystemOwner
	}
	u.store.remove(patternID)
	_ = u.writer.WritePatterns(u.store.allDefinitions())
	return nil
}

// AddUserPattern is called by PipelineBootstrap to seed user patterns after
// the DB migration (once the DB is gone, the store is the only source of truth).
func (u *regexPatternUsecase) AddUserPattern(patternID, description, definition string) {
	u.store.setUser(patternID, description, definition)
}

// --- simple in-memory store ---

type patternEntry struct {
	description string
	definition  string
	system      bool
}

type patternStore struct {
	entries map[string]patternEntry
}

func newPatternStore() *patternStore { return &patternStore{entries: make(map[string]patternEntry)} }

func (s *patternStore) get(id string) (patternEntry, bool) {
	e, ok := s.entries[id]
	return e, ok
}
func (s *patternStore) setSystem(id, definition string) {
	s.entries[id] = patternEntry{definition: definition, system: true}
}
func (s *patternStore) setUser(id, description, definition string) {
	s.entries[id] = patternEntry{description: description, definition: definition, system: false}
}
func (s *patternStore) remove(id string) { delete(s.entries, id) }

func (s *patternStore) allDefinitions() map[string]string {
	out := make(map[string]string, len(s.entries))
	for id, e := range s.entries {
		out[id] = e.definition
	}
	return out
}

func (s *patternStore) list() []dto.RegexPatternResponse {
	out := make([]dto.RegexPatternResponse, 0, len(s.entries))
	for id, e := range s.entries {
		out = append(out, dto.RegexPatternResponse{
			PatternID:          id,
			PatternDescription: e.description,
			PatternDefinition:  e.definition,
			SystemOwner:        e.system,
		})
	}
	return out
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
