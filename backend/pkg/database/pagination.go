package database

const (
	DefaultPageSize = 20
	MaxPageSize     = 200
)

type Params struct {
	Page int `form:"page,default=0"`
	Size int `form:"size,default=20"`
}

func (p Params) Normalized() (page, size int) {
	page, size = p.Page, p.Size
	if page < 0 {
		page = 0
	}
	if size < 1 || size > MaxPageSize {
		size = DefaultPageSize
	}
	return page, size
}

func (p Params) Offset() int {
	page, size := p.Normalized()
	return page * size
}

func (p Params) Limit() int {
	_, size := p.Normalized()
	return size
}

type List[T any] struct {
	Items []T
	Total int64
}
