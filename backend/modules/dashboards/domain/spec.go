package domain

import (
	"errors"
	"time"
)

type Spec struct {
	Dataset string `json:"dataset"` // logs | alerts

	// DataType narrows to one kind of record inside it — o365, wineventlog,
	// syslog. It is the other half of what an index pattern was: v11-log-o365-*
	// named a table and a data type at once. Empty means every kind, and
	// anything broader than one exact type is a filter.
	DataType string `json:"dataType,omitempty"`

	Chart     Chart      `json:"chart"`               // decides which aggregation answers it
	Metric    Metric     `json:"metric"`              // what is counted or summed
	Dimension string     `json:"dimension,omitempty"` // what it is broken down by
	Columns   []string   `json:"columns,omitempty"`   // table charts only
	Filters   []Filter   `json:"filters,omitempty"`
	Interval  string     `json:"interval,omitempty"` // time charts: 1m, 5m, 1h, 1d…
	Limit     int        `json:"limit,omitempty"`
	From      *time.Time `json:"from,omitempty"`
	To        *time.Time `json:"to,omitempty"`
}

type Chart string

const (
	ChartMetric   Chart = "metric"
	ChartCategory Chart = "category"
	ChartTime     Chart = "time"
	ChartTable    Chart = "table"
)

type Metric struct {
	Agg   Agg    `json:"agg"`
	Field string `json:"field,omitempty"`
}

type Agg string

const (
	AggCount         Agg = "count"
	AggCountDistinct Agg = "count_distinct"
	AggSum           Agg = "sum"
	AggAvg           Agg = "avg"
	AggMin           Agg = "min"
	AggMax           Agg = "max"
)

type Filter struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value,omitempty"`
}

var (
	ErrDatasetRequired   = errors.New("dataset is required")
	ErrUnknownDataset    = errors.New("unknown dataset")
	ErrUnknownChart      = errors.New("unknown chart")
	ErrDimensionRequired = errors.New("this chart needs a dimension")
	ErrFieldRequired     = errors.New("this aggregation needs a field")
	ErrUnknownAgg        = errors.New("unknown aggregation")
	ErrUnknownOp         = errors.New("unknown filter operator")
)

var datasets = map[string]bool{"logs": true, "alerts": true}

func (s Spec) Validate() error {
	if s.Dataset == "" {
		return ErrDatasetRequired
	}
	if !datasets[s.Dataset] {
		return ErrUnknownDataset
	}

	switch s.Chart {
	case ChartMetric, ChartTable:
	case ChartCategory:
		if s.Dimension == "" {
			return ErrDimensionRequired
		}
	case ChartTime:
	default:
		return ErrUnknownChart
	}

	switch s.Metric.Agg {
	case "", AggCount:
	case AggCountDistinct, AggSum, AggAvg, AggMin, AggMax:
		if s.Metric.Field == "" {
			return ErrFieldRequired
		}
	default:
		return ErrUnknownAgg
	}

	return nil
}
