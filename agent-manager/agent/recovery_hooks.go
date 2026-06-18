package agent

import (
	"context"
	"sync"

	"github.com/utmstack/UTMStack/agent-manager/models"
)

var (
	OnAgentConnectHook  func(ctx context.Context, agentID uint)
	OnAgentRegisterHook func(agent *models.Agent)
	OnAgentUpdateHook   func(agent *models.Agent)
	OnCommandResultHook func(result *CommandResult) bool
	LockStreamHook      func(agentID uint) sync.Locker
)

func RegisterRecoveryHooks(
	onConnect func(ctx context.Context, agentID uint),
	onRegister func(agent *models.Agent),
	onUpdate func(agent *models.Agent),
	onResult func(result *CommandResult) bool,
	lockStream func(agentID uint) sync.Locker,
) {
	if OnAgentConnectHook == nil {
		OnAgentConnectHook = onConnect
	}
	if OnAgentRegisterHook == nil {
		OnAgentRegisterHook = onRegister
	}
	if OnAgentUpdateHook == nil {
		OnAgentUpdateHook = onUpdate
	}
	if OnCommandResultHook == nil {
		OnCommandResultHook = onResult
	}
	if LockStreamHook == nil {
		LockStreamHook = lockStream
	}
}
