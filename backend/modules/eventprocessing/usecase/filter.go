package usecase

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"gopkg.in/yaml.v3"
)

type procSpec struct {
	strFields  []string
	listFields []string
}

var procSpecs = map[string]procSpec{
	"grok":     {strFields: []string{"source"}, listFields: []string{"patterns"}}, // pattern items checked below
	"kv":       {strFields: []string{"source"}},
	"json":     {strFields: []string{"source"}},
	"csv":      {strFields: []string{"source"}},
	"trim":     {strFields: []string{"function"}, listFields: []string{"fields"}},
	"rename":   {strFields: []string{"to"}, listFields: []string{"from"}},
	"cast":     {strFields: []string{"to"}, listFields: []string{"fields"}},
	"reformat": {strFields: []string{"function"}, listFields: []string{"fields"}},
	"delete":   {listFields: []string{"fields"}},
	"drop":     {}, // a bare condition (where) is enough; nothing structurally required
	"add":      {strFields: []string{"function"}},
	"dynamic":  {strFields: []string{"plugin"}},
}

func validateProcessor(name string, raw any) error {
	spec, ok := procSpecs[name]
	if !ok {
		return fmt.Errorf("unknown processor %q", name)
	}
	cfg, _ := raw.(map[string]any) // nil when the processor has no body
	for _, f := range spec.strFields {
		if s, _ := cfg[f].(string); strings.TrimSpace(s) == "" {
			return fmt.Errorf("%q is missing required field %q", name, f)
		}
	}
	for _, f := range spec.listFields {
		if lst, _ := cfg[f].([]any); len(lst) == 0 {
			return fmt.Errorf("%q needs a non-empty %q", name, f)
		}
	}
	if name == "grok" {
		lst, _ := cfg["patterns"].([]any)
		for k, it := range lst {
			m, _ := it.(map[string]any)
			fn, _ := m["fieldName"].(string)
			pt, _ := m["pattern"].(string)
			if strings.TrimSpace(fn) == "" || strings.TrimSpace(pt) == "" {
				return fmt.Errorf("grok.patterns[%d] needs both fieldName and pattern", k)
			}
		}
	}
	return nil
}

type filterConfig struct {
	Pipeline []struct {
		DataTypes []string         `yaml:"dataTypes"`
		Steps     []map[string]any `yaml:"steps"`
	} `yaml:"pipeline"`
}

func extractDataTypes(content []byte) []string {
	seen := map[string]bool{}
	var out []string
	collectDataTypes(content, 0, seen, &out)
	return out
}

func collectDataTypes(content []byte, depth int, seen map[string]bool, out *[]string) {
	if depth > 4 {
		return
	}
	var cfg filterConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return
	}
	for _, p := range cfg.Pipeline {
		for _, dt := range p.DataTypes {
			if dt != "" && !seen[dt] {
				seen[dt] = true
				*out = append(*out, dt)
			}
		}
		for _, step := range p.Steps {
			ls, ok := step["logstash"].(map[string]any)
			if !ok {
				continue
			}
			if f, ok := ls["filter"].(string); ok && strings.TrimSpace(f) != "" {
				collectDataTypes([]byte(f), depth+1, seen, out)
			}
		}
	}
}

func hasDataType(dts []string, want string) bool {
	for _, dt := range dts {
		if dt == want {
			return true
		}
	}
	return false
}

func validateFilterContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("%w: content is empty", domain.ErrFilterInvalidContent)
	}
	var cfg filterConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return fmt.Errorf("%w: not valid YAML / wrong shape: %v", domain.ErrFilterInvalidContent, err)
	}
	if len(cfg.Pipeline) == 0 {
		return fmt.Errorf("%w: must define at least one pipeline entry", domain.ErrFilterInvalidContent)
	}
	for i, p := range cfg.Pipeline {
		if len(p.DataTypes) == 0 {
			return fmt.Errorf("%w: pipeline[%d] needs at least one dataType", domain.ErrFilterInvalidContent, i)
		}
		if len(p.Steps) == 0 {
			return fmt.Errorf("%w: pipeline[%d] needs at least one step", domain.ErrFilterInvalidContent, i)
		}
		for j, step := range p.Steps {
			if len(step) != 1 {
				return fmt.Errorf("%w: pipeline[%d].steps[%d] must have exactly one processor (found %d)", domain.ErrFilterInvalidContent, i, j, len(step))
			}
			for name, body := range step {
				if err := validateProcessor(name, body); err != nil {
					return fmt.Errorf("%w: pipeline[%d].steps[%d]: %v", domain.ErrFilterInvalidContent, i, j, err)
				}
			}
		}
	}
	return nil
}

type filterUsecase struct {
	store *FilterStore
}

func NewFilterUsecase(store *FilterStore) connectors.FilterUsecase {
	return &filterUsecase{store: store}
}

func (u *filterUsecase) Create(_ context.Context, req dto.CreateFilterRequest) (*dto.FilterResponse, error) {
	if err := validateFilterContent(req.Content); err != nil {
		return nil, err
	}
	entry, err := u.store.Create(req.RelPath, []byte(req.Content))
	if err != nil {
		return nil, mapStoreFilterErr(err)
	}
	return toFilterResponse(entry), nil
}

func (u *filterUsecase) Update(_ context.Context, req dto.UpdateFilterRequest) (*dto.FilterResponse, error) {
	if err := validateFilterContent(req.Content); err != nil {
		return nil, err
	}
	entry, err := u.store.Update(req.RelPath, []byte(req.Content))
	if err != nil {
		return nil, mapStoreFilterErr(err)
	}
	return toFilterResponse(entry), nil
}

func (u *filterUsecase) GetByRelPath(_ context.Context, relPath string) (*dto.FilterResponse, error) {
	entry := u.store.GetByRelPath(relPath)
	if entry == nil {
		return nil, domain.ErrFilterNotFound
	}
	return toFilterResponse(entry), nil
}

func (u *filterUsecase) List(_ context.Context, f dto.FilterFilters) ([]dto.FilterResponse, int64, error) {
	all := u.store.List()

	// Apply in-memory filters.
	out := make([]dto.FilterResponse, 0, len(all))
	for i := range all {
		e := &all[i]
		if f.IsActiveEq != nil && e.Active != *f.IsActiveEq {
			continue
		}
		if f.SystemEq != nil && e.System != *f.SystemEq {
			continue
		}
		if f.RelPathContains != nil && !strings.Contains(e.RelPath, *f.RelPathContains) {
			continue
		}
		resp := toFilterResponse(e)
		if f.DataTypeEq != nil && *f.DataTypeEq != "" && !hasDataType(resp.DataTypes, *f.DataTypeEq) {
			continue
		}
		out = append(out, *resp)
	}

	total := int64(len(out))

	// Pagination.
	page, size := f.Page, f.Size
	if size <= 0 {
		size = 50
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * size
	if start >= len(out) {
		return []dto.FilterResponse{}, total, nil
	}
	end := start + size
	if end > len(out) {
		end = len(out)
	}
	return out[start:end], total, nil
}

func (u *filterUsecase) Delete(_ context.Context, relPath string) error {
	return mapStoreFilterErr(u.store.Delete(relPath))
}

func (u *filterUsecase) SetActive(_ context.Context, relPath string, active bool) error {
	return mapStoreFilterErr(u.store.SetEnabled(relPath, active))
}

func (u *filterUsecase) DataTypes(_ context.Context) []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range u.store.List() {
		for _, dt := range extractDataTypes(e.Content) {
			if !seen[dt] {
				seen[dt] = true
				out = append(out, dt)
			}
		}
	}
	sort.Strings(out)
	return out
}

func toFilterResponse(e *FilterEntry) *dto.FilterResponse {
	return &dto.FilterResponse{
		RelPath:   e.RelPath,
		Content:   string(e.Content),
		System:    e.System,
		Active:    e.Active,
		DataTypes: extractDataTypes(e.Content),
	}
}

func mapStoreFilterErr(err error) error {
	if err == nil {
		return nil
	}
	switch err.(type) {
	case errFilterNotFound:
		return domain.ErrFilterNotFound
	case errFilterSystemOwner:
		return domain.ErrFilterSystemOwner
	}
	return err
}
