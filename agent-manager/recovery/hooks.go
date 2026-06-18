package recovery

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	agentpb "github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/models"
)

// OnAgentConnect is called by agent/agent_imp.go after a new AgentStream is
// registered in AgentStreamMap. It dispatches PENDING recovery targets for
// that agent in a goroutine. No-op if dispatch is disabled or shutting down.
func OnAgentConnect(_ context.Context, agentID uint) {
	if !configDispatchEnabled || disp.shuttingDown.Load() {
		return
	}
	dispatcherOnAgentConnect(agentID)
}

// OnAgentRegister is called by agent/agent_imp.go at the end of RegisterAgent.
// It runs target resolution for the newly-registered agent so it picks up
// any ACTIVE recoveries that match it. Executes in a goroutine so it never
// blocks the registration response.
//
// The kill-switch (configDispatchEnabled) intentionally does NOT gate this
// function. Per spec R10: "target resolution MUST still function" even when
// dispatch is disabled — target rows must be up-to-date so that toggling the
// kill-switch back on immediately resumes dispatch against the full inventory.
// Only the shuttingDown guard is applied so goroutines are not spawned during
// process shutdown.
func OnAgentRegister(agent *models.Agent) {
	if agent == nil {
		return
	}
	// trackGoroutine atomically checks shuttingDown and increments wg.
	if !disp.trackGoroutine() {
		return
	}
	go func() {
		defer disp.wg.Done()
		db := database.GetDB().GormDB()
		// ResolveTargetsForAgent returns errors but does NOT log them — log
		// here so DB issues during target resolution are visible to operators.
		if err := ResolveTargetsForAgent(db, agent); err != nil {
			catcher.Error("recovery: failed to resolve targets for agent",
				err,
				map[string]any{
					"event":    "recovery_target_resolution_failed",
					"agent_id": agent.ID,
					"process":  "agent-manager",
				})
		}
	}()
}

// OnAgentUpdate is called by agent/agent_imp.go at the end of UpdateAgent
// (after the upsert). It re-runs target resolution because the agent's
// Os or Version may have changed, making it eligible for additional recoveries.
func OnAgentUpdate(agent *models.Agent) {
	// Same logic as OnAgentRegister: re-resolve targets for updated agent.
	OnAgentRegister(agent)
}

// OnCommandResult is called by agent/agent_imp.go when a CommandResult arrives
// on the stream. If the cmd_id matches a pending recovery dispatch, this
// returns true and the caller should NOT also write to CommandResultChannel.
// If the cmd_id is not a recovery dispatch, returns false — the caller falls
// back to the existing panel-driven CommandResultChannel logic.
func OnCommandResult(result *agentpb.CommandResult) bool {
	if !configDispatchEnabled {
		return false
	}
	if result == nil {
		return false
	}
	return dispatcherOnCommandResult(result)
}
