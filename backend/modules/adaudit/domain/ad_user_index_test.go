package domain

import (
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

// GORM splits an index tag on commas, so a predicate with a comma in it (an IN
// list) reaches PostgreSQL cut at the first one and the whole AutoMigrate fails
// with "syntax error at end of input". The backend then cannot start.
func TestEveryPartialIndexKeepsItsWholePredicate(t *testing.T) {
	sch, err := schema.Parse(&ADUser{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	for _, idx := range sch.ParseIndexes() {
		name := idx.Name
		if idx.Where == "" {
			continue
		}
		if strings.Count(idx.Where, "(") != strings.Count(idx.Where, ")") {
			t.Errorf("index %s: predicate %q is cut off; GORM splits tags on commas, so keep commas out of it", name, idx.Where)
		}
	}
}

func TestTheCloudIndexCoversAllFourCloudSources(t *testing.T) {
	sch, err := schema.Parse(&ADUser{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	var where string
	found := false
	for _, idx := range sch.ParseIndexes() {
		if idx.Name == "idx_aduser_cloud" {
			where, found = idx.Where, true
		}
	}
	if !found {
		t.Fatal("no idx_aduser_cloud")
	}
	for _, src := range []string{"o365", "azure", "gcp", "aws"} {
		if !strings.Contains(where, "'"+src+"'") {
			t.Errorf("predicate %q does not cover %s", where, src)
		}
	}
}
