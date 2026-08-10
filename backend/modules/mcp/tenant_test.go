package mcp

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// The two invariants this file pins:
//
//  1. On the ctx that reaches an MCP tool handler, the acting tenant is
//     recoverable via authz.TenantIDFromContext AND the acting user is
//     recoverable via ActorFromContext. That is what tenant-scoped repos and
//     the tenancy GORM hooks read. With streamable Stateless mode the request
//     ctx (already enriched by userAuth) becomes the session ctx, so composing
//     the two derivations is what the tool sees.
//
//  2. auditEvent carries the actor's support-access marker forward so audit
//     rows say when a platform administrator acted inside another tenant —
//     matching the REST recorder path (modules/audit/recorder.go).
func TestTenantAndActorReachTheHandler(t *testing.T) {
	actor := &authz.Actor{
		UserID:   uuid.New(),
		Email:    "op@utmstack.com",
		TenantID: "tenant-A",
	}

	ctx := authz.WithTenantID(context.Background(), actor.TenantID)
	ctx = WithActor(ctx, actor)

	if got := authz.TenantIDFromContext(ctx); got != "tenant-A" {
		t.Fatalf("tenant lost from ctx: got %q, want %q", got, "tenant-A")
	}
	got := ActorFromContext(ctx)
	if got == nil || got.TenantID != actor.TenantID || got.Email != actor.Email {
		t.Fatalf("actor lost from ctx: got %+v", got)
	}
}

func TestAuditEventCapturesSupportAccess(t *testing.T) {
	actor := &authz.Actor{
		UserID:   uuid.New(),
		Email:    "platform@utmstack.com",
		TenantID: "tenant-B",
		Support:  authz.SupportFull,
	}

	ev := auditEvent(actor, "incidents.list", statusSuccess, "")

	if ev.SupportAccess != authz.SupportFull {
		t.Fatalf("SupportAccess not carried: got %q, want %q", ev.SupportAccess, authz.SupportFull)
	}
	if ev.UserEmail != actor.Email {
		t.Fatalf("UserEmail not carried: got %q, want %q", ev.UserEmail, actor.Email)
	}
	if ev.UserID == nil || *ev.UserID != actor.UserID {
		t.Fatalf("UserID not carried: got %v, want %v", ev.UserID, actor.UserID)
	}
}
