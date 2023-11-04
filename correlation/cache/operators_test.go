package cache

import (
	"testing"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		val1     string
		val2     string
		want     bool
	}{
		{
			name:     "equal",
			operator: "==",
			val1:     "hello",
			val2:     "hello",
			want:     true,
		},
		{
			name:     "not equal",
			operator: "!=",
			val1:     "hello",
			val2:     "world",
			want:     true,
		},
		{
			name:     "case insensitive equal",
			operator: "::",
			val1:     "Hello",
			val2:     "hello",
			want:     true,
		},
		{
			name:     "case insensitive not equal",
			operator: "!!",
			val1:     "Hello",
			val2:     "world",
			want:     true,
		},
		{
			name:     "contains",
			operator: "contains",
			val1:     "hello world",
			val2:     "world",
			want:     true,
		},
		{
			name:     "not contains",
			operator: "not contain",
			val1:     "hello world",
			val2:     "universe",
			want:     true,
		},
		{
			name:     "in",
			operator: "in",
			val1:     "hello",
			val2:     "hello,world",
			want:     true,
		},
		{
			name:     "not in",
			operator: "not in",
			val1:     "hello",
			val2:     "world,universe",
			want:     true,
		},
		{
			name:     "start with",
			operator: "start with",
			val1:     "hello world",
			val2:     "hello",
			want:     true,
		},
		{
			name:     "not start with",
			operator: "not start with",
			val1:     "hello world",
			val2:     "world",
			want:     true,
		},
		{
			name:     "end with",
			operator: "end with",
			val1:     "hello world",
			val2:     "world",
			want:     true,
		},
		{
			name:     "not end with",
			operator: "not end with",
			val1:     "hello world",
			val2:     "hello",
			want:     true,
		},
		{
			name:     "regexp",
			operator: "regexp",
			val1:     "hello world",
			val2:     "hello.*",
			want:     true,
		},
		{
			name:     "not regexp",
			operator: "not regexp",
			val1:     "hello world",
			val2:     "mon.*",
			want:     true,
		},
		{
			name:     "less than",
			operator: "<",
			val1:     "1",
			val2:     "2",
			want:     true,
		},
		{
			name:     "greater than",
			operator: ">",
			val1:     "2",
			val2:     "1",
			want:     true,
		},
		{
			name:     "less than or equal",
			operator: "<=",
			val1:     "1",
			val2:     "1",
			want:     true,
		},
		{
			name:     "greater than or equal",
			operator: ">=",
			val1:     "2",
			val2:     "2",
			want:     true,
		},
		{
			name:     "exist",
			operator: "exist",
			val1:     "hello",
			val2:     "",
			want:     true,
		},
		{
			name:     "in cidr",
			operator: "in cidr",
			val1:     "192.168.1.1",
			val2:     "192.168.1.0/24",
			want:     true,
		},
		{
			name:     "not in cidr",
			operator: "not in cidr",
			val1:     "192.168.1.1",
			val2:     "192.168.2.0/24",
			want:     true,
		},
		{
			name:     "invalid operator",
			operator: "invalid",
			val1:     "hello",
			val2:     "world",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compare(tt.operator, tt.val1, tt.val2); got != tt.want {
				t.Errorf("compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalElement(t *testing.T) {
	elem := `{"name": "John", "age": 30}`
	tests := []struct {
		name   string
		field  string
		op     string
		value  string
		want   bool
	}{
		{
			name:   "field exists and matches",
			field:  "name",
			op:     "==",
			value:  "John",
			want:   true,
		},
		{
			name:   "field exists and does not match",
			field:  "name",
			op:     "==",
			value:  "Jane",
			want:   false,
		},
		{
			name:   "field does not exist and operator is not exist",
			field:  "address",
			op:     "not exist",
			value:  "",
			want:   true,
		},
		{
			name:   "field does not exist and operator is exist",
			field:  "address",
			op:     "exist",
			value:  "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evalElement(elem, tt.field, tt.op, tt.value); got != tt.want {
				t.Errorf("evalElement() = %v, want %v", got, tt.want)
			}
		})
	}
}