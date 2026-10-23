package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/edr/connectors"
	"github.com/utmstack/utmstack/backend/modules/edr/domain"
	"github.com/utmstack/utmstack/backend/modules/edr/dto"
)

// statusJSON mirrors the agent-side status doc (agent/edr/service statusDoc).
// Fields are additive on the agent; unknown ones are ignored here.
type statusJSON struct {
	Running                bool                 `json:"running"`
	EngineHealthy          bool                 `json:"engine_healthy"`
	SigDBVersion           string               `json:"sigdb_version"`
	SignatureLastUpdate    string               `json:"signature_last_update"`
	BlocklistEnabled       bool                 `json:"blocklist_enabled"`
	BlocklistEnforce       bool                 `json:"blocklist_enforce"`
	BlocklistIndicators    int                  `json:"blocklist_indicators"`
	BlocklistActiveFilters int                  `json:"blocklist_active_filters"`
	BlocklistFeedStale     bool                 `json:"blocklist_feed_stale"`
	BlocklistFeedAgeSec    int64                `json:"blocklist_feed_age_sec"`
	BlocklistRecentBlocks  int                  `json:"blocklist_recent_blocks"`
	AllowlistPaths         []string             `json:"allowlist_paths_entries"`
	AllowlistProcesses     []string             `json:"allowlist_processes_entries"`
	AllowlistCommands      []string             `json:"allowlist_commands_entries"`
	QuarantineItems        []dto.QuarantineItem `json:"quarantine_items"`
}

const (
	staleAfter    = 15 * time.Minute // no status report in 3x the 5min cadence
	sigStaleAfter = 7 * 24 * time.Hour
)

type Service struct {
	repo    connectors.Repository
	gateway connectors.AgentGateway
	counter connectors.DetectionCounter
}

func New(repo connectors.Repository, gateway connectors.AgentGateway, counter connectors.DetectionCounter) *Service {
	return &Service{repo: repo, gateway: gateway, counter: counter}
}

// Policy CRUD.
func (s *Service) CreatePolicy(ctx context.Context, name, desc, version, doc string) (*domain.Policy, error) {
	if !json.Valid([]byte(doc)) {
		return nil, fmt.Errorf("doc is not valid JSON")
	}
	p := &domain.Policy{Name: name, Description: desc, Version: version, Doc: doc}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) GetPolicy(ctx context.Context, id uint) (*domain.Policy, error) {
	return s.repo.Get(ctx, id)
}
func (s *Service) ListPolicies(ctx context.Context) ([]domain.Policy, error) { return s.repo.List(ctx) }

func (s *Service) UpdatePolicy(ctx context.Context, id uint, name, desc, version, doc string) (*domain.Policy, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc != "" {
		if !json.Valid([]byte(doc)) {
			return nil, fmt.Errorf("doc is not valid JSON")
		}
		p.Doc = doc
	}
	if name != "" {
		p.Name = name
	}
	if desc != "" {
		p.Description = desc
	}
	if version != "" {
		p.Version = version
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) DeletePolicy(ctx context.Context, id uint) error { return s.repo.Delete(ctx, id) }

// endpointsFold fetches agents + statuses once and returns the folded list.
func (s *Service) endpointsFold(ctx context.Context) ([]dto.Endpoint, error) {
	agents, err := s.gateway.ListEndpoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("list endpoints: %w", err)
	}
	statuses, err := s.gateway.EdrStatuses(ctx)
	if err != nil {
		// Statuses are best-effort: a flaky manager read degrades the fold,
		// not the agent list itself.
		catcher.Error("edr statuses fetch failed; endpoints show no module state", err, map[string]any{"process": "edr"})
		statuses = nil
	}
	byAgent := make(map[uint32]dto.EdrStatus, len(statuses))
	for _, st := range statuses {
		byAgent[st.AgentID] = st
	}
	now := time.Now().UTC()
	for i := range agents {
		a := &agents[i]
		st, ok := byAgent[a.AgentID]
		if !ok {
			continue
		}
		a.HasModule = true
		a.LastReport = &st.ReportedAt
		a.PolicyVersion = st.PolicyVersion
		var doc statusJSON
		if err := json.Unmarshal([]byte(st.StatusJSON), &doc); err == nil {
			a.ModuleHealthy = doc.Running && doc.EngineHealthy
			a.SigDBVersion = doc.SigDBVersion
			if t, perr := time.Parse(time.RFC3339, doc.SignatureLastUpdate); perr == nil {
				a.StaleSignature = now.Sub(t) > sigStaleAfter
			}
		}
		a.Drift = now.Sub(st.ReportedAt) > staleAfter
	}
	return agents, nil
}

func (s *Service) Overview(ctx context.Context) (*dto.OverviewResponse, error) {
	agents, err := s.endpointsFold(ctx)
	if err != nil {
		return nil, err
	}
	out := &dto.OverviewResponse{Endpoints: len(agents)}
	for _, a := range agents {
		if !a.HasModule {
			continue
		}
		out.WithModule++
		if a.ModuleHealthy {
			out.Healthy++
		} else {
			out.Unhealthy++
		}
		if a.StaleSignature {
			out.StaleSignatures++
		}
		if a.Drift {
			out.Drifted++
		}
	}
	if s.counter != nil {
		// "Last day" = trailing 24h.
		from := time.Now().UTC().Add(-24 * time.Hour)
		n, cerr := s.counter.CountInWindow(ctx, from, time.Now().UTC())
		if cerr != nil {
			catcher.Error("edr detection count failed", cerr, map[string]any{"process": "edr"})
		}
		out.DetectionsToday = n
	}
	return out, nil
}

func (s *Service) Endpoints(ctx context.Context) ([]dto.Endpoint, error) {
	return s.endpointsFold(ctx)
}

func (s *Service) Endpoint(ctx context.Context, agentID uint32) (*dto.Endpoint, error) {
	agents, err := s.endpointsFold(ctx)
	if err != nil {
		return nil, err
	}
	for i := range agents {
		if agents[i].AgentID == agentID {
			return &agents[i], nil
		}
	}
	return nil, ErrEndpointNotFound
}

// Actions the endpoint can actually run. kill_process/full_scan/isolate/
// release_isolation have no agent-side implementation yet — dispatching them
// is answered "unsupported_action" by the endpoint, which is the honest
// surface for the UI.
var knownActions = map[string]struct{}{
	"scan_path": {}, "full_scan": {}, "kill_process": {}, "isolate": {},
	"release_isolation": {}, "quarantine_list": {}, "quarantine_restore": {},
	"quarantine_purge": {}, "blocklist_refresh": {}, "policy_set": {},
	"allow_add": {}, "allow_remove": {}, "enable": {}, "disable": {}, "status": {},
}

// Action dispatches one EDR command to one endpoint.
func (s *Service) Action(ctx context.Context, agentID uint32, action, payload, executedBy, reason string) (*dto.ActionResult, error) {
	if _, ok := knownActions[action]; !ok {
		return nil, fmt.Errorf("unknown action %q", action)
	}
	res := &dto.ActionResult{AgentID: agentID, Action: action}
	if reason == "" {
		reason = "edr console action"
	}
	cmdID, err := s.gateway.Dispatch(ctx, agentID, action, payload, executedBy, reason)
	if err != nil {
		res.Error = err.Error()
		return res, err
	}
	res.CmdID = cmdID
	res.Dispatch = true
	return res, nil
}

// AssignPolicy pushes a policy to a set of endpoints via policy_set. The
// payload is the merge shape the agent expects: {"policy":doc,"version":v}.
func (s *Service) AssignPolicy(ctx context.Context, policyID uint, agentIDs []uint32, executedBy, reason string) ([]dto.ActionResult, error) {
	p, err := s.repo.Get(ctx, policyID)
	if err != nil {
		return nil, err
	}
	if !json.Valid([]byte(p.Doc)) {
		return nil, fmt.Errorf("policy doc is not valid JSON")
	}
	payload, err := json.Marshal(map[string]any{"policy": json.RawMessage(p.Doc), "version": p.Version})
	if err != nil {
		return nil, err
	}
	if reason == "" {
		reason = "edr policy assignment"
	}
	out := make([]dto.ActionResult, 0, len(agentIDs))
	for _, id := range agentIDs {
		res := dto.ActionResult{AgentID: id, Action: "policy_set"}
		cmdID, derr := s.gateway.Dispatch(ctx, id, "policy_set", string(payload), executedBy, reason)
		if derr != nil {
			res.Error = derr.Error()
		} else {
			res.CmdID = cmdID
			res.Dispatch = true
		}
		out = append(out, res)
	}
	return out, nil
}

// Allowlist folds the per-endpoint allowlist entries from statuses.
func (s *Service) Allowlist(ctx context.Context, agentID *uint32) ([]dto.AllowlistEntry, error) {
	agents, err := s.endpointsFold(ctx)
	if err != nil {
		return nil, err
	}
	_ = agents
	statuses, err := s.gateway.EdrStatuses(ctx)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var out []dto.AllowlistEntry
	add := func(cat, val string) {
		k := cat + "\x00" + val
		if _, dup := seen[k]; dup {
			return
		}
		seen[k] = struct{}{}
		out = append(out, dto.AllowlistEntry{Category: cat, Value: val})
	}
	for _, st := range statuses {
		if agentID != nil && st.AgentID != *agentID {
			continue
		}
		var doc statusJSON
		if err := json.Unmarshal([]byte(st.StatusJSON), &doc); err != nil {
			continue
		}
		for _, v := range doc.AllowlistPaths {
			add("allow_paths", v)
		}
		for _, v := range doc.AllowlistProcesses {
			add("allow_processes", v)
		}
		for _, v := range doc.AllowlistCommands {
			add("allow_commands", v)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].Value < out[j].Value
	})
	return out, nil
}

// BlocklistStatus folds the per-endpoint blocklist feed health.
func (s *Service) BlocklistStatus(ctx context.Context) (*dto.BlocklistStatus, error) {
	statuses, err := s.gateway.EdrStatuses(ctx)
	if err != nil {
		return nil, err
	}
	agg := &dto.BlocklistStatus{}
	agg.FeedStale = true // any stale feed => fleet-wide stale
	for _, st := range statuses {
		var doc statusJSON
		if err := json.Unmarshal([]byte(st.StatusJSON), &doc); err != nil {
			continue
		}
		agg.Indicators = max(agg.Indicators, doc.BlocklistIndicators)
		agg.ActiveFilters = max(agg.ActiveFilters, doc.BlocklistActiveFilters)
		agg.RecentBlocks += doc.BlocklistRecentBlocks
		if !doc.BlocklistFeedStale {
			agg.FeedStale = false
		}
		agg.FeedAgeSec = max(agg.FeedAgeSec, doc.BlocklistFeedAgeSec)
		agg.Enforced = agg.Enforced || doc.BlocklistEnforce
		agg.EnforcementDeg = agg.EnforcementDeg || !doc.BlocklistEnforce
	}
	return agg, nil
}

// Quarantine folds live quarantine items across the fleet.
func (s *Service) Quarantine(ctx context.Context, agentID *uint32) ([]dto.QuarantineItem, error) {
	agents, err := s.endpointsFold(ctx)
	if err != nil {
		return nil, err
	}
	host := map[uint32]string{}
	for _, a := range agents {
		host[a.AgentID] = a.Hostname
	}
	statuses, err := s.gateway.EdrStatuses(ctx)
	if err != nil {
		return nil, err
	}
	var out []dto.QuarantineItem
	for _, st := range statuses {
		if agentID != nil && st.AgentID != *agentID {
			continue
		}
		var doc statusJSON
		if err := json.Unmarshal([]byte(st.StatusJSON), &doc); err != nil {
			continue
		}
		for _, item := range doc.QuarantineItems {
			item.AgentID = st.AgentID
			item.Hostname = host[st.AgentID]
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].QuarantinedAt > out[j].QuarantinedAt
	})
	return out, nil
}

// QuarantineItem acts on one quarantined file: restore or purge.
func (s *Service) QuarantineItem(ctx context.Context, agentID uint32, itemID, op, executedBy string) (*dto.ActionResult, error) {
	if op != "restore" && op != "purge" {
		return nil, fmt.Errorf("op must be restore or purge")
	}
	action := "quarantine_" + op
	payload, err := json.Marshal(map[string]string{"id": itemID})
	if err != nil {
		return nil, err
	}
	return s.Action(ctx, agentID, action, string(payload), executedBy, "edr quarantine "+op)
}

// CommandResult returns the current state of a dispatched action by its cmd
// id. Three distinct answers: the row is absent (dispatch itself failed or
// the agent never received it → ErrCommandNotFound, UI shows offline), the
// row is pending (agent has not answered yet), or the row is executed/error
// (the endpoint answered — a refusal like unsupported_action lands here with
// its reason in Result).
func (s *Service) CommandResult(ctx context.Context, agentID uint32, cmdID string) (*dto.CommandRow, error) {
	row, err := s.gateway.Command(ctx, agentID, cmdID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, ErrCommandNotFound
	}
	return row, nil
}

var (
	ErrEndpointNotFound = fmt.Errorf("endpoint not found")
	ErrCommandNotFound  = fmt.Errorf("command not found: dispatch failed or the endpoint is offline")
)
