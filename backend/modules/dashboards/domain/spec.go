package domain

import (
	"errors"
	"time"
)

type Spec struct {
	Dataset   string     `json:"dataset"` // logs | alerts
	DataType  string     `json:"dataType,omitempty"`
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

// Metric is what the chart measures. The event store counts records and does
// nothing else — it has no sum, average or cardinality — so count is the only
// answer there is, and a spec asking for another one is refused rather than
// answered with a count that looks right.
type Metric struct {
	Agg   Agg    `json:"agg"`
	Field string `json:"field,omitempty"`
}

type Agg string

const AggCount Agg = "count"

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
	ErrUnknownAgg        = errors.New("the event store only counts records: agg must be count")
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

	if s.Metric.Agg != "" && s.Metric.Agg != AggCount {
		return ErrUnknownAgg
	}

	return nil
}
