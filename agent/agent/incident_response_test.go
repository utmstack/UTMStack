package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"google.golang.org/grpc"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

// commandProcessor logs, and the package logger is built by whatever starts the
// agent. Nothing does here.
func TestMain(m *testing.M) {
	utils.InitLogger(filepath.Join(os.TempDir(), "utmstack-agent-test.log"))
	os.Exit(m.Run())
}

// captureStream stands in for the server connection and keeps what was sent
// back. Only Send is exercised; the rest of the interface exists to satisfy it.
type captureStream struct {
	grpc.ClientStream
	sent []*BidirectionalStream
}

func (c *captureStream) Send(m *BidirectionalStream) error {
	c.sent = append(c.sent, m)
	return nil
}

func (c *captureStream) Recv() (*BidirectionalStream, error) { panic("not used") }

// The refusal has to be the agent's, not the server's: a server that has been
// taken over is the case this exists for, so "the server did not send it" is
// not an answer. The command here would leave a file behind, and the assertion
// is that it does not.
func TestNoRemoteControlDoesNotRunTheCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the command is written for a POSIX shell")
	}

	dir := t.TempDir()
	witness := filepath.Join(dir, "it-ran")

	stream := &captureStream{}
	cnf := &config.Config{AgentID: 1, NoRemoteControl: true}

	if err := commandProcessor(dir, stream, cnf, "touch "+witness, "cmd-1", "sh"); err != nil {
		t.Fatalf("commandProcessor: %v", err)
	}

	if _, err := os.Stat(witness); err == nil {
		t.Fatal("the command ran on an agent installed with no-remote-control")
	}

	if len(stream.sent) != 1 {
		t.Fatalf("sent %d results, want 1 — the console has to be told", len(stream.sent))
	}
	got := stream.sent[0].GetResult().GetResult()
	if !strings.Contains(got, "no-remote-control") {
		t.Fatalf("result = %q, want it to say why it was refused", got)
	}
}

// The same path with the flag off still runs, so the check cannot be passing
// the test above by breaking execution for everyone.
func TestWithoutTheFlagTheCommandStillRuns(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the command is written for a POSIX shell")
	}

	dir := t.TempDir()
	witness := filepath.Join(dir, "it-ran")

	stream := &captureStream{}
	cnf := &config.Config{AgentID: 1}

	if err := commandProcessor(dir, stream, cnf, "touch "+witness, "cmd-2", "sh"); err != nil {
		t.Fatalf("commandProcessor: %v", err)
	}

	if _, err := os.Stat(witness); err != nil {
		t.Fatalf("the command did not run: %v", err)
	}
}
