package usecase

import (
	"errors"
	"strings"
	"testing"
)

// The error used to be the bare "invalid ingestion-stats parameter": a model that left out group_by could not
// tell what to change and tried again 53 times.
func TestARejectedParameterSaysWhatToChange(t *testing.T) {
	_, err := resolveGroupBy("", true)
	if !errors.Is(err, ErrInvalidStatsParam) || !strings.Contains(err.Error(), "dataSource or dataType") {
		t.Fatalf("%v", err)
	}

	_, err = resolveGroupBy("host", true)
	if !errors.Is(err, ErrInvalidStatsParam) || !strings.Contains(err.Error(), `"host"`) {
		t.Fatalf("%v", err)
	}

	_, _, err = resolveStatus("lost")
	if !errors.Is(err, ErrInvalidStatsParam) || !strings.Contains(err.Error(), "parsing_dropped") {
		t.Fatalf("%v", err)
	}

	if g, err := resolveGroupBy("dataType", true); err != nil || g != "dataType" {
		t.Fatalf("%v %v", g, err)
	}
}
