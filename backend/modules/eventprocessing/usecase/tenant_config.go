package usecase

import (
	"context"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

type tenantConfigUsecase struct {
	writer *pipelineWriter
	store  *assetStore // in-memory assets
}

func NewTenantConfigUsecase(writer *pipelineWriter) connectors.TenantConfigUsecase {
	return &tenantConfigUsecase{writer: writer, store: newAssetStore()}
}

func (u *tenantConfigUsecase) Create(_ context.Context, req dto.CreateTenantConfigRequest) (*dto.TenantConfigResponse, error) {
	if u.store.has(req.AssetName) {
		return nil, domain.ErrIDMustBeAbsent // asset already exists
	}
	u.store.set(req.AssetName, req.AssetHostnameList, req.AssetIpList,
		req.AssetConfidentiality, req.AssetIntegrity, req.AssetAvailability)
	u.syncFile()
	return u.store.toResponse(req.AssetName), nil
}

func (u *tenantConfigUsecase) Update(_ context.Context, req dto.UpdateTenantConfigRequest) (*dto.TenantConfigResponse, error) {
	if !u.store.has(req.AssetName) {
		return nil, domain.ErrTenantConfigNotFound
	}
	u.store.set(req.AssetName, req.AssetHostnameList, req.AssetIpList,
		req.AssetConfidentiality, req.AssetIntegrity, req.AssetAvailability)
	u.syncFile()
	return u.store.toResponse(req.AssetName), nil
}

func (u *tenantConfigUsecase) GetByID(_ context.Context, assetName string) (*dto.TenantConfigResponse, error) {
	if !u.store.has(assetName) {
		return nil, domain.ErrTenantConfigNotFound
	}
	return u.store.toResponse(assetName), nil
}

func (u *tenantConfigUsecase) List(_ context.Context, f dto.TenantConfigFilters) (*connectors.ListResult[dto.TenantConfigResponse], error) {
	all := u.store.list()
	search := strings.ToLower(f.Search)
	out := make([]dto.TenantConfigResponse, 0, len(all))
	for _, a := range all {
		if search != "" && !strings.Contains(strings.ToLower(a.AssetName), search) {
			continue
		}
		out = append(out, a)
	}
	total := int64(len(out))
	page, size := normPage(f.Page, f.Size)
	start := page * size
	if start >= len(out) {
		return &connectors.ListResult[dto.TenantConfigResponse]{Items: []dto.TenantConfigResponse{}, Total: total}, nil
	}
	end := start + size
	if end > len(out) {
		end = len(out)
	}
	return &connectors.ListResult[dto.TenantConfigResponse]{Items: out[start:end], Total: total}, nil
}

func (u *tenantConfigUsecase) Delete(_ context.Context, assetName string) error {
	if !u.store.has(assetName) {
		return domain.ErrTenantConfigNotFound
	}
	u.store.remove(assetName)
	u.syncFile()
	return nil
}

// AddAsset is called by PipelineBootstrap to seed assets from the legacy DB.
func (u *tenantConfigUsecase) AddAsset(name string, hostnames, ips []string, conf, integ, avail int) {
	u.store.set(name, hostnames, ips, conf, integ, avail)
}

func (u *tenantConfigUsecase) syncFile() {
	assets := u.store.toFileYAML()
	_ = u.writer.WriteTenants(assets)
}

// --- in-memory asset store ---

type assetEntry struct {
	hostnames []string
	ips       []string
	conf      int
	integ     int
	avail     int
}

type assetStore struct {
	entries map[string]assetEntry
}

func newAssetStore() *assetStore { return &assetStore{entries: make(map[string]assetEntry)} }

func (s *assetStore) has(name string) bool { _, ok := s.entries[name]; return ok }
func (s *assetStore) remove(name string)   { delete(s.entries, name) }
func (s *assetStore) set(name string, hostnames, ips []string, conf, integ, avail int) {
	s.entries[name] = assetEntry{hostnames: hostnames, ips: ips, conf: conf, integ: integ, avail: avail}
}

func (s *assetStore) toResponse(name string) *dto.TenantConfigResponse {
	e := s.entries[name]
	return &dto.TenantConfigResponse{
		AssetName:            name,
		AssetHostnameList:    e.hostnames,
		AssetIpList:          e.ips,
		AssetConfidentiality: e.conf,
		AssetIntegrity:       e.integ,
		AssetAvailability:    e.avail,
	}
}

func (s *assetStore) list() []dto.TenantConfigResponse {
	out := make([]dto.TenantConfigResponse, 0, len(s.entries))
	for name := range s.entries {
		out = append(out, *s.toResponse(name))
	}
	return out
}

func (s *assetStore) toFileYAML() []assetFileYAML {
	out := make([]assetFileYAML, 0, len(s.entries))
	for name, e := range s.entries {
		out = append(out, assetFileYAML{
			Name:            name,
			Hostnames:       e.hostnames,
			Ips:             e.ips,
			Confidentiality: uint32(e.conf),
			Availability:    uint32(e.avail),
			Integrity:       uint32(e.integ),
		})
	}
	return out
}
