package database

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// AbstractRepository is a generic base repository over the Database provider.
// Embed it in a concrete repository (`AbstractRepository[MyEntity, MyID]`) to get
// CRUD + a declarative list (GetAll) for free; add only entity-specific methods.
//
// It assumes the primary-key column is named "id".
type AbstractRepository[T any, ID any] struct {
	db *DB
}

func NewAbstractRepository[T any, ID any](db *DB) AbstractRepository[T, ID] {
	return AbstractRepository[T, ID]{db: db}
}

// GetByID returns the row by primary key, or (nil, nil) when not found.
func (r AbstractRepository[T, ID]) GetByID(ctx context.Context, id ID) (*T, error) {
	var item T
	err := r.db.FindOne(ctx, &item, Where("id = ?", id))
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Create inserts item and returns the persisted row (with DB-assigned fields).
func (r AbstractRepository[T, ID]) Create(ctx context.Context, item T) (*T, error) {
	if err := r.db.Create(ctx, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateByID saves the full updated row.
func (r AbstractRepository[T, ID]) UpdateByID(ctx context.Context, id ID, updated T) error {
	return r.db.Update(ctx, &updated)
}

// Upsert inserts item or updates the row conflicting on conflictColumns.
func (r AbstractRepository[T, ID]) Upsert(ctx context.Context, item *T, conflictColumns []string) error {
	return r.db.Upsert(ctx, item, conflictColumns)
}

// Remove hard-deletes the row by primary key.
func (r AbstractRepository[T, ID]) Remove(ctx context.Context, id ID) error {
	var model T
	return r.db.DeleteWhere(ctx, &model, Where("id = ?", id))
}

// GetAll runs a filtered/sorted/paginated query driven by the request's
// search_query/sort_by DSL, restricted to the `allowed` columns. defaultOrder is
// the ORDER BY used when the client supplies no (valid) sort. extra options add
// fixed, server-controlled scoping the client cannot override (e.g. a tenant id).
func (r AbstractRepository[T, ID]) GetAll(
	ctx context.Context,
	req common_models.IListRequest,
	allowed []string,
	defaultOrder string,
	extra ...QueryOption,
) (common_models.ListResponse[T], error) {
	page, size := common_models.Normalize(req.GetPageNumber(), req.GetPageSize(), 50, 500)

	// Filter clauses are shared by the count and the find; ordering/pagination
	// are NOT applied to the count.
	filterOpts := append([]QueryOption{}, extra...)
	filterOpts = append(filterOpts, WithFilters(ParseSearchQuery(req.GetSearchQuery(), allowed)))

	total, err := r.db.Count(ctx, new(T), filterOpts...)
	if err != nil {
		return common_models.ListResponse[T]{}, err
	}

	order := ParseSort(req.GetSortBy(), allowed)
	if order == "" {
		order = defaultOrder
	}
	findOpts := append(filterOpts, OrderBy(order), Paginate(page, size))

	var items []T
	if err := r.db.FindAll(ctx, &items, findOpts...); err != nil {
		return common_models.ListResponse[T]{}, err
	}
	return common_models.NewListResponse(items, page, size, int(total)), nil
}
