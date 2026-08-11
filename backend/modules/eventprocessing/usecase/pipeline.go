package usecase

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"gopkg.in/yaml.v3"
	"path/filepath"
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
	var cfg domain.PipelineSpec
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

func firstPipelineOrder(content []byte) int32 {
	var cfg domain.PipelineSpec
	if err := yaml.Unmarshal(content, &cfg); err != nil || len(cfg.Pipeline) == 0 {
		return 0
	}
	return cfg.Pipeline[0].Order
}

func hasDataType(dts []string, want string) bool {
	for _, dt := range dts {
		if dt == want {
			return true
		}
	}
	return false
}

const customFilterDefaultOrder = 100

func validateFilterContent(content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("%w: content is empty", domain.ErrPipelineInvalidContent)
	}
	var cfg domain.PipelineSpec
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return fmt.Errorf("%w: not valid YAML / wrong shape: %v", domain.ErrPipelineInvalidContent, err)
	}
	if len(cfg.Pipeline) == 0 {
		return fmt.Errorf("%w: must define at least one pipeline entry", domain.ErrPipelineInvalidContent)
	}
	for i, p := range cfg.Pipeline {
		if len(p.DataTypes) == 0 {
			return fmt.Errorf("%w: pipeline[%d] needs at least one dataType", domain.ErrPipelineInvalidContent, i)
		}
		if len(p.Steps) == 0 {
			return fmt.Errorf("%w: pipeline[%d] needs at least one step", domain.ErrPipelineInvalidContent, i)
		}
		for j, step := range p.Steps {
			if len(step) != 1 {
				return fmt.Errorf("%w: pipeline[%d].steps[%d] must have exactly one processor (found %d)", domain.ErrPipelineInvalidContent, i, j, len(step))
			}
			for name, body := range step {
				if err := validateProcessor(name, body); err != nil {
					return fmt.Errorf("%w: pipeline[%d].steps[%d]: %v", domain.ErrPipelineInvalidContent, i, j, err)
				}
			}
		}
	}
	return nil
}

func normalizeFilterOrder(content string) (string, error) {
	var cfg domain.PipelineSpec
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return content, err
	}
	changed := false
	for i := range cfg.Pipeline {
		if cfg.Pipeline[i].Order == 0 {
			cfg.Pipeline[i].Order = customFilterDefaultOrder
			changed = true
		}
	}
	if !changed {
		return content, nil
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return content, err
	}
	return string(out), nil
}

type pipelineUsecase struct {
	store  connectors.PipelineRepository
	config connectors.EngineConfigRepository
}

func NewPipelineUsecase(store connectors.PipelineRepository, config connectors.EngineConfigRepository) connectors.PipelineUsecase {
	return &pipelineUsecase{store: store, config: config}
}

func (u *pipelineUsecase) Create(ctx context.Context, req dto.CreatePipelineRequest) (*dto.PipelineResponse, error) {
	if err := validateFilterContent(req.Content); err != nil {
		return nil, err
	}
	content, err := normalizeFilterOrder(req.Content)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrPipelineInvalidContent, err)
	}
	// Empty for every on-prem/single-tenant install — only the shared SaaS
	// deployment's auth middleware ever populates this.
	entry, err := u.store.Create(req.RelPath, []byte(content), authz.TenantIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	return toFilterResponse(entry), nil
}

// visiblePipeline reports whether the caller may see this pipeline: the ones
// the release ships and the global ones belong to everybody, a tenant's belong
// to it alone.
func visiblePipeline(p *domain.Pipeline, tenant string) bool {
	return tenant == "" || p.TenantID == "" || p.TenantID == tenant
}

// readable resolves a pipeline the caller may see. Another tenant's answers
// "not found" rather than "forbidden": relPath is the identity, and confirming
// one exists would tell a customer what another has written.
func (u *pipelineUsecase) readable(ctx context.Context, relPath string) (*domain.Pipeline, error) {
	p := u.store.GetByRelPath(relPath)
	if p == nil || !visiblePipeline(p, authz.TenantIDFromContext(ctx)) {
		return nil, domain.ErrPipelineNotFound
	}
	return p, nil
}

// writable additionally requires the pipeline to be the caller's own.
func (u *pipelineUsecase) writable(ctx context.Context, relPath string) error {
	p, err := u.readable(ctx, relPath)
	if err != nil {
		return err
	}
	if p.System {
		return domain.ErrPipelineSystemOwner
	}
	if tenant := authz.TenantIDFromContext(ctx); tenant != "" && p.TenantID != tenant {
		return domain.ErrPipelineNotFound
	}
	return nil
}

func (u *pipelineUsecase) Update(ctx context.Context, req dto.UpdatePipelineRequest) (*dto.PipelineResponse, error) {
	if err := validateFilterContent(req.Content); err != nil {
		return nil, err
	}
	if err := u.writable(ctx, req.RelPath); err != nil {
		return nil, err
	}
	content, err := normalizeFilterOrder(req.Content)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrPipelineInvalidContent, err)
	}
	entry, err := u.store.Update(req.RelPath, []byte(content))
	if err != nil {
		return nil, err
	}
	return toFilterResponse(entry), nil
}

func (u *pipelineUsecase) GetByRelPath(ctx context.Context, relPath string) (*dto.PipelineResponse, error) {
	entry, err := u.readable(ctx, relPath)
	if err != nil {
		return nil, err
	}
	return toFilterResponse(entry), nil
}

func (u *pipelineUsecase) List(ctx context.Context, f dto.PipelineFilters) (*connectors.ListResult[dto.PipelineResponse], error) {
	all := u.store.List(authz.TenantIDFromContext(ctx))
	tenant := authz.TenantIDFromContext(ctx)

	// Apply in-memory filters.
	out := make([]dto.PipelineResponse, 0, len(all))
	for i := range all {
		e := &all[i]
		if !visiblePipeline(e, tenant) {
			continue
		}
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

	sort.Slice(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].RelPath < out[j].RelPath
	})

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
		return &connectors.ListResult[dto.PipelineResponse]{Items: []dto.PipelineResponse{}, Total: total}, nil
	}
	end := start + size
	if end > len(out) {
		end = len(out)
	}
	return &connectors.ListResult[dto.PipelineResponse]{Items: out[start:end], Total: total}, nil
}

func (u *pipelineUsecase) Delete(ctx context.Context, relPath string) error {
	if err := u.writable(ctx, relPath); err != nil {
		return err
	}
	return u.store.Delete(relPath)
}

// SetActive is checked against readable, not writable: a tenant may switch a
// shipped pipeline off for itself without being able to rewrite it.
func (u *pipelineUsecase) SetActive(ctx context.Context, relPath string, active bool) error {
	if _, err := u.readable(ctx, relPath); err != nil {
		return err
	}
	return u.store.SetEnabled(authz.TenantIDFromContext(ctx), relPath, active)
}

// SetOrder is checked against readable, not writable: which pipeline runs first
// for a data type is operator configuration, like switching one off, and the
// store writes the order back into the shipped file on purpose.
// SetOrder records the whole sequence this tenant wants, in tenants.yaml.
//
// It used to rewrite the order inside the pipeline file, which for a shipped
// pipeline meant reordering it for every tenant at once — the file is one copy
// shared by all of them. The engine reads the per-tenant list and falls back to
// the file order for anyone who never set one.
//
// Names the caller cannot see are refused rather than dropped: silently
// omitting them would save an order the tenant did not ask for.
func (u *pipelineUsecase) SetOrder(ctx context.Context, order []string) error {
	byName := make(map[string]bool, len(order))
	for _, p := range u.store.List(authz.TenantIDFromContext(ctx)) {
		if visiblePipeline(&p, authz.TenantIDFromContext(ctx)) {
			byName[pipelineIdentity(p.RelPath)] = true
		}
	}
	for _, name := range order {
		if !byName[name] {
			return fmt.Errorf("%w: %s", domain.ErrPipelineNotFound, name)
		}
	}
	return u.config.SetPipelineOrder(authz.TenantIDFromContext(ctx), order)
}

// pipelineIdentity is the name the engine matches on: the file's base name
// without its extension, the same identity used in the disabled list.
func pipelineIdentity(relPath string) string {
	base := filepath.Base(relPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func (u *pipelineUsecase) DataTypes(ctx context.Context) []string {
	seen := map[string]bool{}
	tenant := authz.TenantIDFromContext(ctx)
	var out []string
	for _, e := range u.store.List(authz.TenantIDFromContext(ctx)) {
		if !visiblePipeline(&e, tenant) {
			continue
		}
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

func toFilterResponse(e *domain.Pipeline) *dto.PipelineResponse {
	return &dto.PipelineResponse{
		RelPath:   e.RelPath,
		Content:   string(e.Content),
		System:    e.System,
		Active:    e.Active,
		DataTypes: extractDataTypes(e.Content),
		Order:     firstPipelineOrder(e.Content),
	}
}
