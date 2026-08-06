package mcp

import (
	"context"
	"github.com/google/uuid"

	auditcon "github.com/utmstack/utmstack/backend/modules/audit/connectors"
	auditdom "github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// statusSuccess / statusFailure are the audit status strings used across
// REST handlers; mirror them here so /audit displays MCP rows consistently.
const (
	statusSuccess = "success"
	statusFailure = "failure"
)

func (m *Module) logger() auditcon.Logger {
	if m.deps == nil || m.deps.Audit == nil {
		return nil
	}
	return m.deps.Audit.Logger()
}

func (m *Module) auditAttempt(ctx context.Context, actor *authz.Actor, tool string) {
	// We audit only the outcome (success/failure/denied) — the attempt-level
	// row would double the row count without adding signal. Keep this hook
	// in place though so future correlation work has a single insertion point.
	_ = ctx
	_ = actor
	_ = tool
}

func (m *Module) auditSuccess(ctx context.Context, actor *authz.Actor, tool string) {
	l := m.logger()
	if l == nil {
		return
	}
	l.Log(ctx, auditEvent(actor, tool, statusSuccess, ""))
}

func (m *Module) auditFailure(ctx context.Context, actor *authz.Actor, tool string, err error) {
	l := m.logger()
	if l == nil || err == nil {
		return
	}
	l.Log(ctx, auditEvent(actor, tool, statusFailure, err.Error()))
}

func (m *Module) auditDenied(ctx context.Context, actor *authz.Actor, tool string, err error) {
	l := m.logger()
	if l == nil || err == nil {
		return
	}
	ev := auditEvent(actor, tool, statusFailure, err.Error())
	ev.Metadata["denied"] = true
	l.Log(ctx, ev)
}

func auditEvent(actor *authz.Actor, tool, status, errMsg string) auditcon.Event {
	var uid, sid *uuid.UUID
	var login string
	if actor != nil {
		login = actor.Email
		if actor.UserID != uuid.Nil {
			id := actor.UserID
			uid = &id
		}
		if actor.SessionID != uuid.Nil {
			s := actor.SessionID
			sid = &s
		}
	}
	return auditcon.Event{
		Action:       "mcp.tools.call",
		Status:       status,
		ResourceType: tool,
		ErrorMessage: errMsg,
		Metadata:     map[string]any{"transport": "mcp"},
		EventType:    auditdom.APP_EVENT_MCP_TOOL_CALL,
		UserEmail:    login,
		UserID:       uid,
		SessionID:    sid,
	}
}
