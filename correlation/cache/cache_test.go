package cache_test

import (
	"testing"
	

	"github.com/utmstack/UTMStack/correlation/cache"
	"github.com/utmstack/UTMStack/correlation/rules"
)

func TestSearch(t *testing.T) {
	cacheStorage := []string{
		`{"@timestamp":"2022-01-01T00:00:00.000Z","field1":"value1","field2":"value2"}`,
		`{"@timestamp":"2022-01-01T00:00:01.000Z","field1":"value1","field2":"value2"}`,
		`{"@timestamp":"2022-01-01T00:00:02.000Z","field1":"value1","field2":"value2"}`,
		`{"@timestamp":"2022-01-01T00:00:03.000Z","field1":"value1","field2":"value2"}`,
		`{"@timestamp":"2022-01-01T00:00:04.000Z","field1":"value1","field2":"value2"}`,
	}
	cache.CacheStorage = cacheStorage
	allOf := []rules.AllOf{
		{Field: "field1", Operator: "==", Value: "value1"},
	}
	oneOf := []rules.OneOf{
		{Field: "field2", Operator: "==", Value: "value2"},
	}
	seconds := 9999999999999
	expected := []string{
		`{"@timestamp":"2022-01-01T00:00:04.000Z","field1":"value1","field2":"value2"}`,
		`{"@timestamp":"2022-01-01T00:00:03.000Z","field1":"value1","field2":"value2"}`,
		`{"@timestamp":"2022-01-01T00:00:02.000Z","field1":"value1","field2":"value2"}`,
		`{"@timestamp":"2022-01-01T00:00:01.000Z","field1":"value1","field2":"value2"}`,
		`{"@timestamp":"2022-01-01T00:00:00.000Z","field1":"value1","field2":"value2"}`,
	}
	result := cache.Search(allOf, oneOf, int64(seconds))
	if len(result) != len(expected) {
		t.Errorf("Expected %d elements, but got %d", len(expected), len(result))
	}
	for i, r := range result {
		if r != expected[i] {
			t.Errorf("Expected %s, but got %s", expected[i], r)
		}
	}
}