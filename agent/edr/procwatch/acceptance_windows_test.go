//go:build windows && edr_acceptance

package procwatch

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/responder"
)

type acceptanceHandler struct{ events chan ProcStart }

func (h acceptanceHandler) OnProcStart(p ProcStart) {
	select {
	case h.events <- p:
	default:
	}
}

func TestWindowsAcceptanceWMISafeChild(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" {
		t.Skip("disposable Windows VM only")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	h := acceptanceHandler{make(chan ProcStart, 256)}
	go New(h).Run(ctx)
	time.Sleep(2 * time.Second) // Let the 1s WMI polling subscription initialize.
	child := exec.Command(os.Getenv("SystemRoot")+`\System32\cmd.exe`, "/d", "/c", "ping -n 12 127.0.0.1 >NUL")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = child.Process.Kill(); _ = child.Wait() }()
	identity, err := (responder.OSKiller{}).ProcessIdentity(child.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	for {
		select {
		case p := <-h.events:
			if p.PID != child.Process.Pid {
				continue
			}
			if p.OSIdentity != identity || !p.IdentityRequired || !strings.HasSuffix(strings.ToLower(p.Image), `\cmd.exe`) {
				t.Fatalf("unbound event: %+v expected identity=%d", p, identity)
			}
			t.Logf("WMI observed own child pid=%d creation=%d", p.PID, p.OSIdentity)
			return
		case <-ctx.Done():
			t.Fatal("own harmless child was not observed by the real WMI subscription")
		}
	}
}
