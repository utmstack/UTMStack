package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
)

const (
	shellOriginType = "INCIDENT_RESPONSE_AUTOMATION"
	shellExecutedBy = "SYSTEM"
)

// ErrAgentOffline signals that the target agent could not receive the command;
// the dispatcher decides whether to retry or fail based on this sentinel.
var ErrAgentOffline = errors.New("soar shell: agent offline")

// ErrAgentNotFound signals that no agent matched the requested hostname.
var ErrAgentNotFound = errors.New("soar shell: agent not found")

// agentManagerClient is the slice of *agentmanager.AgentManagerClient that
// Shell actually needs — narrowed to an interface so a test can fake the
// agent-manager round trip instead of needing a real gRPC connection.
type agentManagerClient interface {
	ProcessCommand(ctx context.Context, cmd *agent.UtmCommand) (*agent.CommandResult, error)
	ListAgents(ctx context.Context, searchQuery string) ([]*agent.Agent, int32, error)
}

// Shell runs the node's command on a UTMStack endpoint agent via the
// agent-manager gRPC bidi stream. Output is the raw stdout — enrichment nodes
// get it as JSON when it parses.
type Shell struct {
	agent agentManagerClient
}

func NewShell(client *agentmanager.AgentManagerClient) *Shell { return &Shell{agent: client} }

func (Shell) Type() string { return "shell" }

func (s *Shell) Execute(ctx context.Context, exec *domain.SoarExecution) (json.RawMessage, error) {
	if s.agent == nil {
		return nil, errors.New("soar shell: agent-manager client not configured")
	}
	if exec.Command == "" {
		return nil, errors.New("soar shell: empty command")
	}
	if exec.Agent == "" {
		return nil, ErrAgentNotFound
	}

	agentID, found, err := s.resolveAgent(ctx, exec.Agent)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrAgentNotFound
	}

	cmd := &agent.UtmCommand{
		AgentId:    agentID,
		Command:    exec.Command,
		ExecutedBy: shellExecutedBy,
		OriginType: shellOriginType,
		OriginId:   exec.RulePath,
		Reason:     fmt.Sprintf("Incident response automation: rule %s node %s alert %s", exec.RulePath, exec.NodeID, exec.AlertID),
		Shell:      exec.Shell,
	}

	res, err := s.agent.ProcessCommand(ctx, cmd)
	if err != nil {
		if isOfflineError(err) {
			return nil, ErrAgentOffline
		}
		return nil, err
	}

	result := res.GetResult()
	exec.Result = result

	// SOAR-F2: a nonzero exit code is a real failure — exec.Result above
	// already carries the command's output (including the agent's own
	// stderr/error text) for diagnostics, but returning here lets the
	// dispatcher (dispatch.go's invoke/settle) record ExecutionStatusFailed
	// and take the error branch edge instead of silently treating this as
	// EXECUTED/success.
	if res.GetExitCode() != 0 {
		return nil, fmt.Errorf("soar shell: command exited with status %d", res.GetExitCode())
	}

	// Enrichment shells get their output surfaced as JSON when the stdout
	// parses. Non-JSON output is fine — dispatcher won't record it.
	if exec.Kind == domain.NodeKindEnrichment {
		trimmed := strings.TrimSpace(result)
		if trimmed == "" {
			return json.RawMessage(`{}`), nil
		}
		var probe any
		if err := json.Unmarshal([]byte(trimmed), &probe); err != nil {
			return nil, fmt.Errorf("soar shell enrichment: stdout is not JSON: %w", err)
		}
		return json.RawMessage(trimmed), nil
	}
	return nil, nil
}

func (s *Shell) resolveAgent(ctx context.Context, hostname string) (string, bool, error) {
	rows, _, err := s.agent.ListAgents(ctx, "hostname.Is="+hostname)
	if err != nil {
		return "", false, err
	}
	if len(rows) == 0 {
		return "", false, nil
	}
	return strconv.FormatUint(uint64(rows[0].GetId()), 10), true, nil
}

func isOfflineError(err error) bool {
	if err == nil {
		return false
	}
	if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
		return true
	}
	return strings.Contains(err.Error(), "not found or is disconnected")
}
