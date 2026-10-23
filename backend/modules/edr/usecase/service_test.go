package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/edr/dto"
)

// fakes
type fakeGateway struct {
	agents   []dto.Endpoint
	statuses []dto.EdrStatus
	cmd      *dto.CommandRow
}

func (f *fakeGateway) ListEndpoints(context.Context) ([]dto.Endpoint, error) { return f.agents, nil }
func (f *fakeGateway) EdrStatuses(context.Context) ([]dto.EdrStatus, error)  { return f.statuses, nil }
func (f *fakeGateway) Dispatch(ctx context.Context, agentID uint32, action, payloadJSON, executedBy, reason string) (string, error) {
	return "cmd-1", nil
}
func (f *fakeGateway) Command(ctx context.Context, agentID uint32, cmdID string) (*dto.CommandRow, error) {
	if f.cmd != nil && f.cmd.CmdID == cmdID {
		return f.cmd, nil
	}
	return nil, nil
}

type fakeCounter struct{ n int64 }

func (f fakeCounter) CountInWindow(context.Context, time.Time, time.Time) (int64, error) {
	return f.n, nil
}

func TestOverviewFold(t *testing.T) {
	now := time.Now().UTC()
	svc := New(nil, &fakeGateway{
		agents: []dto.Endpoint{
			{AgentID: 1, Hostname: "a", Status: "online"},
			{AgentID: 2, Hostname: "b", Status: "online"}, // no status: no module
		},
		statuses: []dto.EdrStatus{
			{
				AgentID: 1,
				StatusJSON: `{"running":true,"engine_healthy":true,"sigdb_version":"1","signature_last_update":"` +
					now.Add(-24*time.Hour).Format(time.RFC3339) + `","blocklist_enforce":true,"blocklist_indicators":10}`,
				ReportedAt:    now.Add(-1 * time.Minute),
				PolicyVersion: "v1",
			},
		},
	}, fakeCounter{n: 7})

	ov, err := svc.Overview(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ov.Endpoints != 2 || ov.WithModule != 1 || ov.Healthy != 1 || ov.DetectionsToday != 7 {
		t.Fatalf("bad overview: %+v", ov)
	}
}

// TestCommandResult distinguishes refused (row exists, status error) from
// offline/never-received (no row).
func TestCommandResult(t *testing.T) {
	t.Run("refused", func(t *testing.T) {
		svc := New(nil, &fakeGateway{cmd: &dto.CommandRow{CmdID: "c1", Status: "error", Result: "(unsupported_action)"}}, nil)
		row, err := svc.CommandResult(context.Background(), 1, "c1")
		if err != nil || row.Status != "error" {
			t.Fatalf("want error row, got %+v err=%v", row, err)
		}
	})
	t.Run("never received", func(t *testing.T) {
		svc := New(nil, &fakeGateway{}, nil)
		if _, err := svc.CommandResult(context.Background(), 1, "c1"); !errors.Is(err, ErrCommandNotFound) {
			t.Fatalf("want ErrCommandNotFound, got %v", err)
		}
	})
}

func TestEndpointDrift(t *testing.T) {
	now := time.Now().UTC()
	svc := New(nil, &fakeGateway{
		agents:   []dto.Endpoint{{AgentID: 1, Hostname: "a"}},
		statuses: []dto.EdrStatus{{AgentID: 1, StatusJSON: `{}`, ReportedAt: now.Add(-2 * time.Hour)}},
	}, nil)
	ep, err := svc.Endpoint(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !ep.Drift {
		t.Fatalf("expected drift for 2h-old report: %+v", ep)
	}
}
