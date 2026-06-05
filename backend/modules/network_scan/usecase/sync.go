package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
)

// AssetSync orchestrates the periodic reconciliation of network_scan assets against
// OpenSearch ingestion statistics. Mirrors the Java AssetSynchronizationService at a
// high level. Kept dependency-minimal so it can run unattended.
type AssetSync struct {
	repo        connectors.NetworkScanRepository
	dataInputs  connectors.DataInputStatusGateway
	agent       connectors.AgentGateway
	provider    *SourceActivityProvider
	overlapSecs int
}

func NewAssetSync(
	repo connectors.NetworkScanRepository,
	dataInputs connectors.DataInputStatusGateway,
	agent connectors.AgentGateway,
	provider *SourceActivityProvider,
) *AssetSync {
	return &AssetSync{
		repo:        repo,
		dataInputs:  dataInputs,
		agent:       agent,
		provider:    provider,
		overlapSecs: 5,
	}
}

// RunOnce performs a single reconciliation cycle.
func (s *AssetSync) RunOnce(ctx context.Context) error {
	cp, err := s.dataInputs.GetCheckpoint(ctx)
	if err != nil {
		return err
	}
	since := time.Now().Add(-12 * time.Hour).UTC()
	if cp != nil && !cp.LastProcessedTimestamp.IsZero() {
		since = cp.LastProcessedTimestamp.Add(-time.Duration(s.overlapSecs) * time.Second)
	}

	activities, err := s.provider.Fetch(ctx, since)
	if err != nil {
		return err
	}
	if len(activities) == 0 {
		return s.advanceCheckpoint(ctx, cp, time.Now().UTC())
	}

	var agentNames map[string]struct{}
	if s.agent != nil {
		names, listErr := s.agent.ListAgentNames(ctx)
		if listErr != nil {
			_ = catcher.Error("network_scan sync: ListAgentNames", listErr, nil)
		} else {
			agentNames = make(map[string]struct{}, len(names))
			for _, n := range names {
				agentNames[n] = struct{}{}
			}
		}
	}

	// Index existing assets by name so we can update in place.
	existing, err := s.repo.FindAllByUpdateLevelIn(ctx, []domain.UpdateLevel{
		domain.UpdateLevelDataSource,
	})
	if err != nil {
		return err
	}
	byName := make(map[string]*domain.UtmNetworkScan, len(existing))
	for i := range existing {
		byName[strings.ToLower(existing[i].AssetName)] = &existing[i]
	}

	now := time.Now().UTC()
	var maxTs time.Time = now
	for _, a := range activities {
		if a.Timestamp.After(maxTs) {
			maxTs = a.Timestamp
		}
		if err := s.processActivity(ctx, a, byName, agentNames, now); err != nil {
			_ = catcher.Error("network_scan sync: processActivity "+a.DataSource, err, nil)
		}
	}

	return s.advanceCheckpoint(ctx, cp, maxTs)
}

func (s *AssetSync) processActivity(
	ctx context.Context,
	a SourceActivity,
	byName map[string]*domain.UtmNetworkScan,
	agentNames map[string]struct{},
	now time.Time,
) error {
	key := strings.ToLower(a.DataSource)
	asset, ok := byName[key]
	alive := true
	if !ok {
		asset = &domain.UtmNetworkScan{
			AssetName:      a.DataSource,
			DiscoveredAt:   &now,
			AssetStatus:    string(domain.AssetStatusNew),
			RegisteredMode: string(domain.AssetRegisteredModeDynamic),
		}
	}
	asset.ModifiedAt = &now
	asset.AssetAlive = &alive
	asset.UpdateLevel = string(domain.UpdateLevelDataSource)
	if agentNames != nil {
		_, isAgent := agentNames[a.DataSource]
		asset.IsAgent = &isAgent
	}
	if err := s.repo.Save(ctx, asset); err != nil {
		return err
	}
	byName[key] = asset

	// Persist a row in utm_data_input_status. ID is stable per (source, dataType).
	id := statusID(a.DataSource, a.DataType)
	status := connectors.DataInputStatus{
		ID:        id,
		Source:    a.DataSource,
		DataType:  a.DataType,
		Timestamp: a.Timestamp.Unix(),
	}
	return s.dataInputs.Upsert(ctx, status)
}

func (s *AssetSync) advanceCheckpoint(ctx context.Context, prev *connectors.Checkpoint, to time.Time) error {
	id := int64(1)
	if prev != nil && prev.ID != 0 {
		id = prev.ID
	}
	return s.dataInputs.SaveCheckpoint(ctx, connectors.Checkpoint{
		ID:                     id,
		LastProcessedTimestamp: to,
	})
}

func statusID(source, dataType string) string {
	h := sha256.Sum256([]byte(strings.ToLower(source) + "|" + strings.ToLower(dataType)))
	return hex.EncodeToString(h[:])[:64]
}
