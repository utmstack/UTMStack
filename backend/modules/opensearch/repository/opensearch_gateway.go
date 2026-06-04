package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/opensearch/connectors"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
)

type osGatewayRepo struct {
	client *ospkg.Client
}

func NewOSGatewayRepository(client *ospkg.Client) connectors.GatewayRepository {
	return &osGatewayRepo{client: client}
}

func (r *osGatewayRepo) PropertyValues(ctx context.Context, index, field string) ([]string, error) {
	counts, err := r.client.GetFieldValues(ctx, index, field, nil, 10000, true, false)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(counts))
	for k := range counts {
		out = append(out, k)
	}
	return out, nil
}

func (r *osGatewayRepo) PropertyValuesWithCount(
	ctx context.Context,
	index, field string,
	filters map[string]any,
	top int,
	orderByCount, sortAsc bool,
) (map[string]int64, error) {
	return r.client.GetFieldValues(ctx, index, field, filters, top, orderByCount, sortAsc)
}

func (r *osGatewayRepo) IndexProperties(ctx context.Context, pattern string) ([]ospkg.IndexPropertyType, error) {
	return r.client.GetIndexProperties(ctx, pattern)
}

func (r *osGatewayRepo) Indices(ctx context.Context, pattern string) ([]ospkg.IndexInfo, error) {
	return r.client.GetIndices(ctx, pattern)
}

func (r *osGatewayRepo) DeleteIndices(ctx context.Context, indices []string) error {
	return r.client.DeleteIndices(ctx, indices)
}

func (r *osGatewayRepo) Search(
	ctx context.Context,
	index string,
	query map[string]any,
	from, size int,
	sortField, sortOrder string,
) ([]map[string]any, int64, error) {
	return r.client.SearchWithTotal(ctx, index, query, from, size, sortField, sortOrder)
}

func (r *osGatewayRepo) Count(ctx context.Context, index string, query map[string]any) (int64, error) {
	_, total, err := r.client.SearchWithTotal(ctx, index, query, 0, 0, "", "")
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *osGatewayRepo) CountExists(ctx context.Context, index string, query map[string]any) (bool, error) {
	total, err := r.Count(ctx, index, query)
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *osGatewayRepo) SearchSQL(ctx context.Context, sql string) ([]map[string]any, error) {
	return r.client.SearchSQL(ctx, sql)
}

func (r *osGatewayRepo) CountSQL(ctx context.Context, wrappedSQL string) (int64, error) {
	return r.client.CountSQL(ctx, wrappedSQL)
}

func (r *osGatewayRepo) ClusterHealth(ctx context.Context) (*ospkg.ClusterHealth, error) {
	return r.client.GetClusterHealth(ctx)
}
