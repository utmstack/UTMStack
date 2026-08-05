package main

import (
	"testing"
	"time"
)

// An auditd sequence is a counter local to a host, so it is unique only inside
// a tenant. Two tenants running a host of the same name — DC01, srv-01 — both
// emit sequence 1, 2, 3; keyed without the tenant, whichever arrived second was
// dropped as a duplicate and its audit record lost.
func TestDedupIsPerTenant(t *testing.T) {
	auditdDedup = map[string]time.Time{}

	const a, b = "tenant-a", "tenant-b"

	if dedupCheckAndMark(a, "DC01", "1") {
		t.Fatal("the first record was treated as a duplicate")
	}
	if dedupCheckAndMark(b, "DC01", "1") {
		t.Error("another tenant's record with the same host and sequence was dropped")
	}
	if !dedupCheckAndMark(a, "DC01", "1") {
		t.Error("a genuine repeat was not deduplicated")
	}
}
