package usecase

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/opensearch/connectors"
	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type Gateway struct {
	repo connectors.GatewayRepository
}

func NewGateway(repo connectors.GatewayRepository) *Gateway {
	return &Gateway{repo: repo}
}

func (g *Gateway) PropertyValues(ctx context.Context, req dto.PropertyValuesRequest) ([]string, error) {
	return g.repo.PropertyValues(ctx, req.IndexPattern, req.Keyword)
}

func (g *Gateway) PropertyValuesWithCount(ctx context.Context, req dto.PropertyValuesWithCountRequest) (map[string]int64, error) {
	top := req.Top
	if top <= 0 {
		top = 10000
	}

	var filters map[string]any
	if len(req.Filters) > 0 {
		filters = common_models.FiltersToQuery(req.Filters)
	}

	return g.repo.PropertyValuesWithCount(ctx, req.Index, req.Field, filters, top, req.OrderByCount, req.SortAsc)
}

func (g *Gateway) IndexProperties(ctx context.Context, pattern string) ([]dto.IndexPropertyType, error) {
	return g.repo.IndexProperties(ctx, pattern)
}

func (g *Gateway) IndexAll(ctx context.Context, pattern string, includeSystem bool, page, size int) ([]dto.IndexInfo, int64, error) {
	all, err := g.repo.Indices(ctx, pattern)
	if err != nil {
		return nil, 0, err
	}

	// Filter system indices.
	filtered := make([]dto.IndexInfo, 0, len(all))
	for _, idx := range all {
		if !includeSystem && strings.HasPrefix(idx.Name, ".") {
			continue
		}
		filtered = append(filtered, idx)
	}

	total := int64(len(filtered))

	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	start := (page - 1) * size
	if start >= len(filtered) {
		return []dto.IndexInfo{}, total, nil
	}
	end := start + size
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

func (g *Gateway) DeleteIndex(ctx context.Context, indices []string) error {
	return g.repo.DeleteIndices(ctx, indices)
}

func (g *Gateway) Search(
	ctx context.Context,
	filters []common_models.FilterType,
	top int,
	indexPattern string,
	includeChildren bool,
	page, size int,
	sortBy, sortOrder string,
) ([]map[string]any, int64, error) {
	if top <= 0 {
		top = 500
	}

	query := common_models.FiltersToQuery(filters)

	from := (page - 1) * size
	actualSize := size
	if top-from < actualSize {
		actualSize = top - from
	}
	if actualSize <= 0 {
		return []map[string]any{}, 0, nil
	}

	hits, total, err := g.repo.Search(ctx, indexPattern, query, from, actualSize, sortBy, sortOrder)
	if err != nil {
		return nil, 0, err
	}

	if !includeChildren {
		return hits, total, nil
	}

	for i, hit := range hits {
		id, ok := hit["id"]
		if !ok || id == nil {
			continue
		}
		idStr := fmt.Sprintf("%v", id)

		echoFilters := []common_models.FilterType{
			{Field: "parentId", Operator: common_models.OpEquals, Value: idStr},
		}
		echoQuery := common_models.FiltersToQuery(echoFilters)

		echoCount, countErr := g.repo.Count(ctx, indexPattern, echoQuery)
		if countErr != nil {
			continue
		}

		hits[i]["hasChildren"] = echoCount > 0
		hits[i]["echoes"] = echoCount

		if echoCount > 0 {
			latestHits, _, latestErr := g.repo.Search(ctx, indexPattern, echoQuery, 0, 1, "@timestamp", "desc")
			if latestErr == nil && len(latestHits) > 0 {
				hits[i]["last_echo"] = latestHits[0]
			}
		}
	}

	return hits, total, nil
}

func (g *Gateway) GenericSearch(
	ctx context.Context,
	index string,
	filters []common_models.FilterType,
	top int,
	page, size int,
) ([]map[string]any, int64, error) {
	if top <= 0 {
		top = 500
	}

	query := common_models.FiltersToQuery(filters)

	from := (page - 1) * size
	actualSize := size
	if top-from < actualSize {
		actualSize = top - from
	}
	if actualSize <= 0 {
		return []map[string]any{}, 0, nil
	}

	hits, total, err := g.repo.Search(ctx, index, query, from, actualSize, "", "")
	if err != nil {
		return nil, 0, err
	}

	if total > int64(top) {
		total = int64(top)
	}

	return hits, total, nil
}

func (g *Gateway) Count(ctx context.Context, filters []common_models.FilterType, indexPattern string) (bool, error) {
	query := common_models.FiltersToQuery(filters)
	return g.repo.CountExists(ctx, indexPattern, query)
}

func (g *Gateway) SearchSQL(
	ctx context.Context,
	sqlDto dto.SqlSearchDto,
	page, size int,
) ([]map[string]any, int64, error) {
	if err := ValidateSQL(sqlDto.Query); err != nil {
		return nil, 0, err
	}

	sanitized := strings.TrimSpace(sqlDto.Query)
	sanitized = reTrailingSemicolons.ReplaceAllString(sanitized, "")
	sanitized = strings.TrimSpace(sanitized)

	paginatedSQL := ApplyPagination(sanitized, page, size)

	hits, err := g.repo.SearchSQL(ctx, paginatedSQL)
	if err != nil {
		return nil, 0, err
	}

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS total_count", sanitized)
	total, err := g.repo.CountSQL(ctx, countSQL)
	if err != nil {
		total = 0
	}

	const maxTotal = int64(10000)
	if total > maxTotal {
		total = maxTotal
	}

	return hits, total, nil
}

func (g *Gateway) SearchCSV(
	ctx context.Context,
	filters []common_models.FilterType,
	top int,
	indexPattern string,
	columns []dto.DataColumn,
) ([]map[string]any, error) {
	hits, _, err := g.Search(ctx, filters, top, indexPattern, false, 1, top, "", "")
	if err != nil {
		return nil, err
	}

	if includesColumn(columns, "echoes") {
		for i, hit := range hits {
			id, ok := hit["id"]
			if !ok || id == nil {
				hits[i]["echoes"] = int64(0)
				continue
			}
			idStr := fmt.Sprintf("%v", id)
			echoFilters := []common_models.FilterType{
				{Field: "parentId", Operator: common_models.OpEquals, Value: idStr},
			}
			echoQuery := common_models.FiltersToQuery(echoFilters)
			_, echoTotal, echoErr := g.repo.Search(ctx, indexPattern, echoQuery, 0, 0, "", "")
			if echoErr != nil {
				hits[i]["echoes"] = int64(0)
			} else {
				hits[i]["echoes"] = echoTotal
			}
		}
	}

	return hits, nil
}

func (g *Gateway) ExportCSV(hits []map[string]any, columns []dto.DataColumn, w io.Writer) error {
	return ExportCSV(hits, columns, w)
}

func (g *Gateway) ClusterStatus(ctx context.Context) (*dto.ClusterStatusResponse, error) {
	health, err := g.repo.ClusterHealth(ctx)
	if err != nil {
		return nil, err
	}
	if health == nil {
		return nil, nil
	}
	return &dto.ClusterStatusResponse{Health: health}, nil
}

// ---------------------------------------------------------------------------
// private helpers
// ---------------------------------------------------------------------------

var reTrailingSemicolons = regexp.MustCompile(`;+$`)

func includesColumn(columns []dto.DataColumn, target string) bool {
	for _, c := range columns {
		if stripKeyword(c.Field) == target {
			return true
		}
	}
	return false
}
