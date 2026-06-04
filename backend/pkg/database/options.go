package database

import "gorm.io/gorm"

// WhereClause is a parameterized condition.
type WhereClause struct {
	Query string
	Args  []any
}

// PreloadConfig configures a relation preload, optionally with a condition.
type PreloadConfig struct {
	Relation  string
	Condition string
	Args      []any
}

// QueryOptions is the accumulated query spec built by QueryOption funcs.
type QueryOptions struct {
	Where    []WhereClause
	Preloads []PreloadConfig
	Joins    []string
	Order    string
	Limit    int
	Offset   int
	Select   []string
	Distinct bool
}

// QueryOption mutates QueryOptions; compose them as varargs.
type QueryOption func(*QueryOptions)

// ApplyOptions folds the options into a QueryOptions value.
func ApplyOptions(opts ...QueryOption) *QueryOptions {
	o := &QueryOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Where adds a raw parameterized condition (for fixed, server-controlled scoping).
func Where(query string, args ...any) QueryOption {
	return func(o *QueryOptions) {
		o.Where = append(o.Where, WhereClause{Query: query, Args: args})
	}
}

// Preload eager-loads the named relations.
func Preload(relations ...string) QueryOption {
	return func(o *QueryOptions) {
		for _, r := range relations {
			o.Preloads = append(o.Preloads, PreloadConfig{Relation: r})
		}
	}
}

// PreloadWithCondition eager-loads a relation filtered by a condition.
func PreloadWithCondition(relation, query string, args ...any) QueryOption {
	return func(o *QueryOptions) {
		o.Preloads = append(o.Preloads, PreloadConfig{Relation: relation, Condition: query, Args: args})
	}
}

// Join adds raw JOIN clauses.
func Join(joins ...string) QueryOption {
	return func(o *QueryOptions) { o.Joins = append(o.Joins, joins...) }
}

// OrderBy sets the ORDER BY clause (no-op when empty).
func OrderBy(order string) QueryOption {
	return func(o *QueryOptions) {
		if order != "" {
			o.Order = order
		}
	}
}

// Limit sets the row cap.
func Limit(n int) QueryOption { return func(o *QueryOptions) { o.Limit = n } }

// Offset sets how many rows to skip.
func Offset(n int) QueryOption { return func(o *QueryOptions) { o.Offset = n } }

// Paginate sets limit+offset from a 1-indexed page and size.
func Paginate(page, size int) QueryOption {
	return func(o *QueryOptions) {
		if page < 1 {
			page = 1
		}
		if size < 1 {
			size = 10
		}
		o.Limit = size
		o.Offset = (page - 1) * size
	}
}

// Select restricts the retrieved columns.
func Select(columns ...string) QueryOption {
	return func(o *QueryOptions) { o.Select = append(o.Select, columns...) }
}

// Distinct adds DISTINCT.
func Distinct() QueryOption {
	return func(o *QueryOptions) { o.Distinct = true }
}

// Latest orders by created_at DESC and limits to 1.
func Latest() QueryOption {
	return func(o *QueryOptions) {
		o.Order = "created_at DESC"
		o.Limit = 1
	}
}

// ByID is shorthand for Where("id = ?", id).
func ByID(id any) QueryOption { return Where("id = ?", id) }

// ByIDs is shorthand for Where("id IN ?", ids).
func ByIDs(ids any) QueryOption { return Where("id IN ?", ids) }

// Apply folds the options onto a GORM query.
func Apply(tx *gorm.DB, opts ...QueryOption) *gorm.DB {
	o := ApplyOptions(opts...)
	for _, w := range o.Where {
		tx = tx.Where(w.Query, w.Args...)
	}
	for _, j := range o.Joins {
		tx = tx.Joins(j)
	}
	for _, p := range o.Preloads {
		if p.Condition != "" {
			tx = tx.Preload(p.Relation, append([]any{p.Condition}, p.Args...)...)
		} else {
			tx = tx.Preload(p.Relation)
		}
	}
	if len(o.Select) > 0 {
		tx = tx.Select(o.Select)
	}
	if o.Distinct {
		tx = tx.Distinct()
	}
	if o.Order != "" {
		tx = tx.Order(o.Order)
	}
	if o.Limit > 0 {
		tx = tx.Limit(o.Limit)
	}
	if o.Offset > 0 {
		tx = tx.Offset(o.Offset)
	}
	return tx
}
