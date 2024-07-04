package utils

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Operator string

const (
	Is         Operator = "Is"
	IsNot      Operator = "IsNot"
	Contain    Operator = "Contain"
	NotContain Operator = "NotContain"
	In         Operator = "In"
	NotIn      Operator = "NotIn"
)

type Filter struct {
	Field string
	Op    Operator
	Value interface{}
}

func NewFilter(searchQuery string) []Filter {
	defer func() {
		if r := recover(); r != nil {
			// Handle the panic here
			fmt.Println("Panic occurred:", r)
		}
	}()

	filters := make([]Filter, 0)
	if searchQuery == "" {
		return filters
	}
	query := strings.Split(searchQuery, "&")
	if len(query) == 0 {
		return filters
	}
	for _, v := range query {
		filter := strings.Split(v, "=")
		filerQuery := strings.Split(filter[0], ".")
		filters = append(filters, Filter{
			Field: filerQuery[0],
			Op:    resolveOperator(filerQuery[1]),
			Value: filter[1]})
	}
	return filters
}

func resolveOperator(op string) Operator {
	var operator Operator
	switch op {
	case "Is":
		operator = Is
	case "IsNot":
		operator = IsNot
	case "Contain":
		operator = Contain
	case "NotContain":
		operator = NotContain
	case "In":
		operator = In
	case "NotIn":
		operator = NotIn
	}
	return operator
}

func FilterScope(filters []Filter) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, filter := range filters {
			switch filter.Op {
			case Is:
				db.Where(filter.Field+" = ?", filter.Value)
			case IsNot:
				db.Where(filter.Field+" <> ?", filter.Value)
			case Contain:
				db.Where(filter.Field+" like %?%", filter.Value)
			case NotContain:
				db.Where(filter.Field+" not like %?%", filter.Value)
			case In:
				db.Where(filter.Field+" in (?)", filter.Value)
			case NotIn:
				db.Where(filter.Field+" not in (?)", filter.Value)
			}
		}
		return db
	}
}
