package dispatcher

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	amgr "github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/plugins/soar/internal/cache"
	"github.com/utmstack/UTMStack/plugins/soar/internal/client"
)

const (
	tickInterval = 5 * time.Minute
	originType   = "INCIDENT_RESPONSE_AUTOMATION"
	executedBy   = "SYSTEM"
	MaxRetries   = 3
)

func Start(c *client.BackendClient) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in dispatcher", nil, map[string]any{"panic": r})
		}
	}()

	t := time.NewTicker(tickInterval)
	defer t.Stop()
	tick(c)
	for range t.C {
		tick(c)
	}
}

func tick(c *client.BackendClient) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in dispatcher tick", nil, map[string]any{"panic": r})
		}
	}()

	listCtx, listCancel := context.WithTimeout(context.Background(), 20*time.Second)
	cmds, err := c.ListPendingExecutions(listCtx)
	listCancel()
	if err != nil || len(cmds) == 0 {
		return
	}

	panel, agents, conn, err := dial()
	if err != nil {
		_ = catcher.Error("failed to dial agent-manager", err, nil)
		return
	}
	defer func() { _ = conn.Close() }()

	for _, exec := range cmds {
		// Backend's pending list is unfiltered by rule.active; check the rule
		// cache. Missing means the flow was disabled/deleted (or never indexed,
		// e.g. legacy executions from before the RulePath migration) — skip.
		rule, ok := cache.GetRule(exec.RulePath)
		if !ok {
			continue
		}
		processExecution(c, panel, agents, exec, rule.Shell)
	}
}

func processExecution(c *client.BackendClient, panel amgr.PanelServiceClient, agents amgr.AgentServiceClient, exec client.Execution, shell string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	agentID, found, err := lookupAgentID(ctx, agents, exec.Agent)
	if err != nil {
		_ = catcher.Error("failed to look up agent", err, map[string]any{"execution": exec.ID, "agent": exec.Agent})
		return
	}
	if !found {
		failed := client.StatusFailed
		cause := client.CauseAgentNotFound
		_ = c.UpdateExecution(ctx, exec.ID, client.UpdateExecutionRequest{
			ExecutionStatus:   &failed,
			NonExecutionCause: &cause,
		})
		return
	}

	result, sendErr := sendCommand(ctx, panel, agentID, exec, shell)
	if sendErr != nil {
		if isOfflineError(sendErr) {
			cause := client.CauseAgentOffline
			if exec.ExecutionRetries+1 >= MaxRetries {
				failed := client.StatusFailed
				_ = c.UpdateExecution(ctx, exec.ID, client.UpdateExecutionRequest{
					ExecutionStatus:   &failed,
					NonExecutionCause: &cause,
				})
			} else {
				_ = c.UpdateExecution(ctx, exec.ID, client.UpdateExecutionRequest{
					NonExecutionCause: &cause,
					IncrementRetries:  true,
				})
			}
			return
		}
		_ = catcher.Error("failed to send command", sendErr, map[string]any{"execution": exec.ID})
		failed := client.StatusFailed
		cause := client.CauseUnknown
		_ = c.UpdateExecution(ctx, exec.ID, client.UpdateExecutionRequest{
			ExecutionStatus:   &failed,
			NonExecutionCause: &cause,
		})
		return
	}

	executed := client.StatusExecuted
	_ = c.UpdateExecution(ctx, exec.ID, client.UpdateExecutionRequest{
		ExecutionStatus: &executed,
		CommandResult:   &result,
	})
}

func sendCommand(ctx context.Context, panel amgr.PanelServiceClient, agentID string, exec client.Execution, shell string) (string, error) {
	stream, err := panel.ProcessCommand(ctx)
	if err != nil {
		return "", err
	}

	reason := fmt.Sprintf("Incident response automation: rule %s matched alert %s", exec.RulePath, exec.AlertID)
	cmd := &amgr.UtmCommand{
		AgentId:    agentID,
		Command:    exec.Command,
		ExecutedBy: executedBy,
		OriginType: originType,
		OriginId:   exec.RulePath,
		Reason:     reason,
		Shell:      shell,
	}
	if err := stream.Send(cmd); err != nil {
		return "", err
	}
	if err := stream.CloseSend(); err != nil {
		return "", err
	}

	var b strings.Builder
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		b.WriteString(res.GetResult())
	}
	return b.String(), nil
}

func lookupAgentID(ctx context.Context, client amgr.AgentServiceClient, hostname string) (string, bool, error) {
	if hostname == "" {
		return "", false, nil
	}
	req := &amgr.ListRequest{
		PageNumber:  1,
		PageSize:    100000,
		SearchQuery: "hostname.Is=" + hostname,
	}
	rs, err := client.ListAgents(ctx, req)
	if err != nil {
		return "", false, err
	}
	if len(rs.Rows) == 0 {
		return "", false, nil
	}
	return strconv.FormatUint(uint64(rs.Rows[0].Id), 10), true, nil
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
