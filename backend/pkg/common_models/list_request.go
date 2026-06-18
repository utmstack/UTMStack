package common_models

// IListRequest is satisfied by any paginated request — including module filter
// structs that embed ListRequest — so generic helpers can read paging without
// knowing the concrete type.
type IListRequest interface {
	GetPageNumber() int
	GetPageSize() int
	GetSearchQuery() string
	GetSortBy() string
}

// ListRequest is the common paging/search/sort input. form tags let it bind from
// query params (GET list endpoints); json tags cover body binding. SearchQuery
// carries the filter DSL (field.op.value&...), SortBy the sort DSL (field.dir&...).
type ListRequest struct {
	PageNumber  int    `json:"page_number" form:"page_number"`
	PageSize    int    `json:"page_size" form:"page_size"`
	SearchQuery string `json:"search_query" form:"search_query"`
	SortBy      string `json:"sort_by" form:"sort_by"`
}

// Value receivers so both ListRequest and *ListRequest satisfy IListRequest.
func (r ListRequest) GetPageNumber() int     { return r.PageNumber }
func (r ListRequest) GetPageSize() int       { return r.PageSize }
func (r ListRequest) GetSearchQuery() string { return r.SearchQuery }
func (r ListRequest) GetSortBy() string      { return r.SortBy }

// Normalize clamps paging values: page >= 1, size defaulted to defaultSize when
// unset and capped at maxSize (pass maxSize <= 0 for no cap).
func Normalize(page, size, defaultSize, maxSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = defaultSize
	}
	if maxSize > 0 && size > maxSize {
		size = maxSize
	}
	return page, size
}
