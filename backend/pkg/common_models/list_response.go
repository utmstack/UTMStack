package common_models

import "math"

//for securely generic typing

type IListResponse[T any] interface {
	GetItems() []T
	GetPageNumber() int
	GetPageSize() int
	GetTotalItems() int
	GetTotalPages() int

	SetItems([]T)
	SetPageNumber(int)
	SetPageSize(int)
	SetTotalItems(int)
	SetTotalPages(int)
}

type ListResponse[T any] struct {
	Items      []T `json:"items"`
	PageNumber int `json:"page_number"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

func (lr *ListResponse[T]) SetItems(items []T) {
	lr.Items = items
}

func (lr *ListResponse[T]) SetPageNumber(pageNumber int) {
	lr.PageNumber = pageNumber
}

func (lr *ListResponse[T]) SetPageSize(pageSize int) {
	lr.PageSize = pageSize
}

func (lr *ListResponse[T]) SetTotalItems(totalItems int) {
	lr.TotalItems = totalItems
}

func (lr *ListResponse[T]) SetTotalPages(totalPages int) {
	lr.TotalPages = totalPages
}

func (lr *ListResponse[T]) GetItems() []T {
	return lr.Items
}

func (lr *ListResponse[T]) GetPageNumber() int {
	return lr.PageNumber
}

func (lr *ListResponse[T]) GetPageSize() int {
	return lr.PageSize
}

func (lr *ListResponse[T]) GetTotalItems() int {
	return lr.TotalItems
}

func (lr *ListResponse[T]) GetTotalPages() int {
	return lr.TotalPages
}

// NewListResponse builds the envelope and computes TotalPages from size/total.
// A nil item slice is normalized to an empty slice so the JSON is [] not null.
func NewListResponse[T any](items []T, page, size, totalItems int) ListResponse[T] {
	if items == nil {
		items = []T{}
	}
	pages := 0
	if size > 0 {
		pages = int(math.Ceil(float64(totalItems) / float64(size)))
	}
	return ListResponse[T]{
		Items:      items,
		PageNumber: page,
		PageSize:   size,
		TotalItems: totalItems,
		TotalPages: pages,
	}
}

// MapList converts a ListResponse[T] into a ListResponse[U] by applying fn to
// each item, preserving the page metadata — e.g. domain entities → DTOs.
func MapList[T any, U any](src ListResponse[T], fn func(T) U) ListResponse[U] {
	items := make([]U, 0, len(src.Items))
	for _, it := range src.Items {
		items = append(items, fn(it))
	}
	return ListResponse[U]{
		Items:      items,
		PageNumber: src.PageNumber,
		PageSize:   src.PageSize,
		TotalItems: src.TotalItems,
		TotalPages: src.TotalPages,
	}
}
