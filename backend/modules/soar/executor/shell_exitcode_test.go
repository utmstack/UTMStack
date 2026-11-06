package executor

import (
	"context"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
)

type fakeAgentManagerClient struct {
	result     *agent.CommandResult
	processErr error
	agents     []*agent.Agent
}

func (f *fakeAgentManagerClient) ProcessCommand(context.Context, *agent.UtmCommand) (*agent.CommandResult, error) {
	return f.result, f.processErr
}

func (f *fakeAgentManagerClient) ListAgents(context.Context, string) ([]*agent.Agent, int32, error) {
	return f.agents, int32(len(f.agents)), nil
}

func newTestShell(client agentManagerClient) *Shell { return &Shell{agent: client} }

func testExec() *domain.SoarExecution {
	return &domain.SoarExecution{Command: "whoami", Agent: "DC01"}
}

// SOAR-F2: a shell step whose command failed must return an error — the
// dispatcher's invoke()/settle() otherwise records ExecutionStatusExecuted
// and follows the success branch for a command that actually failed.
func TestShellExecuteFailsOnNonzeroExitCode(t *testing.T) {
	client := &fakeAgentManagerClient{
		agents: []*agent.Agent{{Id: 1}},
		result: &agent.CommandResult{Result: "permission denied", ExitCode: 1},
	}
	s := newTestShell(client)

	exec := testExec()
	_, err := s.Execute(context.Background(), exec)
	if err == nil {
		t.Fatal("Execute() = nil error, want an error for a nonzero exit code")
	}
	if exec.Result != "permission denied" {
		t.Errorf("exec.Result = %q, want the command's output preserved for diagnostics", exec.Result)
	}
}

// A zero exit code (including the zero-value default an unpatched agent
// would send) must still be treated as success — no regression for agents
// built before this field existed.
func TestShellExecuteSucceedsOnZeroExitCode(t *testing.T) {
	client := &fakeAgentManagerClient{
		agents: []*agent.Agent{{Id: 1}},
		result: &agent.CommandResult{Result: "root\n", ExitCode: 0},
	}
	s := newTestShell(client)

	exec := testExec()
	if _, err := s.Execute(context.Background(), exec); err != nil {
		t.Fatalf("Execute() = %v, want nil for a successful command", err)
	}
	if exec.Result != "root\n" {
		t.Errorf("exec.Result = %q, want the command's output", exec.Result)
	}
}

// Enrichment nodes still get their JSON output surfaced on success.
func TestShellExecuteEnrichmentStillParsesJSONOnSuccess(t *testing.T) {
	client := &fakeAgentManagerClient{
		agents: []*agent.Agent{{Id: 1}},
		result: &agent.CommandResult{Result: `{"ok":true}`, ExitCode: 0},
	}
	s := newTestShell(client)

	exec := testExec()
	exec.Kind = domain.NodeKindEnrichment
	output, err := s.Execute(context.Background(), exec)
	if err != nil {
		t.Fatalf("Execute() = %v, want nil", err)
	}
	if string(output) != `{"ok":true}` {
		t.Errorf("output = %q, want the parsed JSON passed through", string(output))
	}
}
