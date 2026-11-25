package repository

import (
	"testing"

	"github.com/utmstack/go-sdk/v2/store"
)

func TestFloodBucketsFlattenTheNestedGroups(t *testing.T) {
	groups := []store.Group{
		{Field: "tenantId", Key: "t1", Children: []store.Group{
			{Field: "name", Key: "RuleA", Children: []store.Group{
				{Field: "dataSource", Key: "fw", Count: 70},
				{Field: "dataSource", Key: "ids", Count: 5},
			}},
		}},
		{Field: "tenantId", Key: "t2", Children: []store.Group{
			{Field: "name", Key: "RuleA", Children: []store.Group{
				{Field: "dataSource", Key: "fw", Count: 9},
			}},
		}},
	}

	got := floodBucketsFromGroups(groups)
	if len(got) != 3 {
		t.Fatalf("got %d buckets, want 3: %+v", len(got), got)
	}

	if b := got[0]; b.TenantID != "t1" || b.RuleName != "RuleA" || b.DataSource != "fw" || b.Count != 70 {
		t.Errorf("first bucket: %+v", b)
	}

	if b := got[2]; b.TenantID != "t2" || b.Count != 9 {
		t.Errorf("tenants must stay apart: %+v", b)
	}
}

func TestFloodFiltersExcludeFalsePositivesGroupedAndClosedAlerts(t *testing.T) {
	var fields []string
	for _, f := range floodFilters() {
		fields = append(fields, f.Field)
	}

	if len(fields) != 3 || fields[0] != "status" || fields[1] != "tags" || fields[2] != "parentId" {
		t.Errorf("unexpected filters: %v", fields)
	}
}

func TestFloodGroupsByTenantFirst(t *testing.T) {
	if floodGroupFields[0] != "tenantId" {
		t.Error("tenantId is the isolation and must stay in the grouping")
	}
}
