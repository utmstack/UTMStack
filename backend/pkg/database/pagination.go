package database

import "context"

type PageInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

type PageResult[T any] struct {
	Data     []T      `json:"data"`
	PageInfo PageInfo `json:"pageInfo"`
}

func NewPageInfo(page, pageSize int, totalItems int64) PageInfo {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	totalPages := int(totalItems) / pageSize
	if int(totalItems)%pageSize > 0 {
		totalPages++
	}
	return PageInfo{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

func FindPaginated[T any](ctx context.Context, db Database, page, pageSize int, opts ...QueryOption) (*PageResult[T], error) {
	var data []T
	var model T

	total, err := db.Count(ctx, &model, opts...)
	if err != nil {
		return nil, err
	}

	opts = append(opts, Paginate(page, pageSize))
	if err := db.FindAll(ctx, &data, opts...); err != nil {
		return nil, err
	}

	return &PageResult[T]{
		Data:     data,
		PageInfo: NewPageInfo(page, pageSize, total),
	}, nil
}
