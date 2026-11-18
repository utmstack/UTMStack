package netblock

import (
	"errors"
	"io"
)

const (
	MaxNetworkFeedRawBytes      int64 = 64 << 20
	MaxNetworkFeedRawTotal      int64 = 128 << 20
	MaxNetworkFeedExpandedBytes int64 = 128 << 20
	MaxNetworkFeedExpandedTotal int64 = 256 << 20
	MaxNetworkFeedIndicators          = nftMaxEntries
	MaxNetworkFeedDailyOps            = 2 * nftMaxEntries
	MaxNetworkFeedArtifacts           = 64
	maxNetworkFeedLineBytes           = 1 << 20
)

var ErrNetworkFeedBudget = errors.New("network indicator feed exceeds bounded ingestion budget")

type networkFeedLimits struct {
	RawArtifact, RawTotal, ExpandedArtifact, ExpandedTotal int64
	Indicators, DailyOps, Artifacts                        int
}

func defaultNetworkFeedLimits() networkFeedLimits {
	return networkFeedLimits{MaxNetworkFeedRawBytes, MaxNetworkFeedRawTotal, MaxNetworkFeedExpandedBytes, MaxNetworkFeedExpandedTotal, MaxNetworkFeedIndicators, MaxNetworkFeedDailyOps, MaxNetworkFeedArtifacts}
}
func (l networkFeedLimits) valid() bool {
	m := defaultNetworkFeedLimits()
	return l.RawArtifact > 0 && l.RawArtifact <= m.RawArtifact && l.RawTotal > 0 && l.RawTotal <= m.RawTotal && l.ExpandedArtifact > 0 && l.ExpandedArtifact <= m.ExpandedArtifact && l.ExpandedTotal > 0 && l.ExpandedTotal <= m.ExpandedTotal && l.Indicators > 0 && l.Indicators <= m.Indicators && l.DailyOps > 0 && l.DailyOps <= m.DailyOps && l.Artifacts > 0 && l.Artifacts <= m.Artifacts
}

func smallerBudget(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func readNetworkFeed(r io.Reader, max int64) ([]byte, error) {
	if max <= 0 || max > MaxNetworkFeedRawBytes {
		return nil, ErrNetworkFeedBudget
	}
	b, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, ErrNetworkFeedBudget
	}
	return b, nil
}
