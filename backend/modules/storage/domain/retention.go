package domain

import (
	"fmt"
	"time"
)

type Dataset string

const (
	DatasetLogs   Dataset = "logs"
	DatasetAlerts Dataset = "alerts"
	DatasetStats  Dataset = "statistics"
)

func Datasets() []Dataset { return []Dataset{DatasetLogs, DatasetAlerts, DatasetStats} }

func (d Dataset) Valid() bool {
	switch d {
	case DatasetLogs, DatasetAlerts, DatasetStats:
		return true
	}
	return false
}

const Day = 24 * time.Hour

type Retention struct {
	Dataset  Dataset
	KeepDays int
	ColdDays int
}

func (r Retention) Tiered() bool { return r.ColdDays > 0 }

func (r Retention) Keep() time.Duration { return time.Duration(r.KeepDays) * Day }
func (r Retention) Cold() time.Duration { return time.Duration(r.ColdDays) * Day }

var defaults = map[Dataset]int{
	DatasetLogs:   730,
	DatasetAlerts: 730,
	DatasetStats:  1095,
}

func DefaultRetention(d Dataset) Retention {
	return Retention{Dataset: d, KeepDays: defaults[d]}
}

const MaxKeepDays = 36500

func (r Retention) Validate() error {
	if !r.Dataset.Valid() {
		return fmt.Errorf("%w: %s", ErrUnknownDataset, r.Dataset)
	}
	if r.KeepDays < 1 {
		return ErrKeepRequired
	}
	if r.KeepDays > MaxKeepDays {
		return fmt.Errorf("%w: %d days", ErrKeepRequired, r.KeepDays)
	}
	if r.ColdDays < 0 {
		return ErrColdNegative
	}
	if r.ColdDays > 0 && r.ColdDays >= r.KeepDays {
		return fmt.Errorf("%w: cold at %d days, deleted at %d", ErrColdBeforeDelete, r.ColdDays, r.KeepDays)
	}
	return nil
}
