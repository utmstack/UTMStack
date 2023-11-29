package statistics_test

import (
	"testing"

	"github.com/utmstack/UTMStack/correlation/statistics"
)

func TestOne(t *testing.T) {
	statistics.One("macos", "macbook")
	statistics.One("macos", "macbook2")
	statistics.One("macos", "macbook")
	statistics.One("macos", "macbook2")
	statistics.One("macos", "macbook2")
	st := statistics.GetStats()

	if len(st) != 2 {
		t.Fatal("Not have 2 elements")
	}

	if st["macosmacbook"].Count != 2 {
		t.Fatal("Macbook should have 2 logs")
	}

	if st["macosmacbook2"].Count != 3 {
		t.Fatal("Macbook2 should have 3 logs")
	}
}
