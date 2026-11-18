//go:build linux

package netblock

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// One-way receive avoids treating an echo's independently filtered return path
// as evidence about the requested direction. Every socket stays in the two
// gated disposable namespaces; neither is the guest's initial network.
func nsDHCPDatagram(t *testing.T, peer string, inbound bool, source, target string, sport, dport int) bool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	receive := `import socket,sys
s=socket.socket(socket.AF_INET6 if ':' in sys.argv[1] else socket.AF_INET,socket.SOCK_DGRAM)
s.settimeout(0.4);s.bind((sys.argv[1],int(sys.argv[2])))
print('ready',flush=True)
try:
 data,remote=s.recvfrom(64)
 sys.exit(0 if data==b'openedr-dhcp-inert' and remote[0]==sys.argv[3] and remote[1]==int(sys.argv[4]) else 2)
except socket.timeout:sys.exit(1)
finally:s.close()
`
	send := `import errno,socket,sys
s=socket.socket(socket.AF_INET6 if ':' in sys.argv[1] else socket.AF_INET,socket.SOCK_DGRAM)
s.bind((sys.argv[1],int(sys.argv[2])))
try:s.sendto(b'openedr-dhcp-inert',(sys.argv[3],int(sys.argv[4])))
except OSError as e:
 if e.errno in (errno.EPERM,errno.EACCES):sys.exit(3)
 raise
finally:s.close()
`
	rargs := []string{"python3", "-u", "-c", receive, target, strconv.Itoa(dport), source, strconv.Itoa(sport)}
	sargs := []string{"python3", "-c", send, source, strconv.Itoa(sport), target, strconv.Itoa(dport)}
	if inbound {
		sargs = append([]string{"ip", "netns", "exec", peer}, sargs...)
	} else {
		rargs = append([]string{"ip", "netns", "exec", peer}, rargs...)
	}
	r := exec.CommandContext(ctx, rargs[0], rargs[1:]...)
	var stderr bytes.Buffer
	r.Stderr = &stderr
	out, err := r.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Start(); err != nil {
		t.Fatal(err)
	}
	ready, readyErr := bufio.NewReader(out).ReadString('\n')
	if readyErr != nil || ready != "ready\n" {
		_ = r.Wait()
		t.Fatalf("DHCP receiver fixture: %v %q %s", readyErr, ready, stderr.String())
	}
	locallyDenied := false
	if data, err := exec.CommandContext(ctx, sargs[0], sargs[1:]...).CombinedOutput(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 3 {
			// An OUTPUT drop may synchronously reject sendto. Still require
			// receiver non-delivery; bind/start/other errors remain failures.
			locallyDenied = true
		} else {
			_ = r.Process.Kill()
			_ = r.Wait()
			t.Fatalf("DHCP sender fixture: %v %s", err, data)
		}
	}
	err = r.Wait()
	if ctx.Err() != nil {
		t.Fatal("DHCP fixture deadline expired")
	}
	if err == nil {
		if locallyDenied {
			t.Fatal("locally denied datagram unexpectedly reached receiver")
		}
		return true
	}
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
		t.Fatalf("DHCP receiver failed: %v %s", err, stderr.String())
	}
	return false
}

func TestNFTNamespaceDHCPClientExceptions(t *testing.T) {
	nftNamespaceGate(t)
	peer := fmt.Sprintf("edr-dhcp-%d", os.Getpid())
	nsRun(t, "ip", "link", "set", "lo", "up")
	nsRun(t, "ip", "netns", "add", peer)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = exec.CommandContext(ctx, "ip", "netns", "delete", peer).Run()
	})
	nsRun(t, "ip", "link", "add", "edr-dhcp-s", "type", "veth", "peer", "name", "edr-dhcp-p")
	nsRun(t, "ip", "link", "set", "edr-dhcp-p", "netns", peer)
	for _, a := range []string{"192.0.2.2/24", "2001:db8:1::2/64"} {
		args := []string{"ip", "addr", "add", a, "dev", "edr-dhcp-s"}
		if strings.Contains(a, ":") {
			args = append(args, "nodad")
		}
		nsRun(t, args...)
	}
	for _, a := range []string{"192.0.2.1/24", "2001:db8:1::1/64"} {
		args := []string{"ip", "netns", "exec", peer, "ip", "addr", "add", a, "dev", "edr-dhcp-p"}
		if strings.Contains(a, ":") {
			args = append(args, "nodad")
		}
		nsRun(t, args...)
	}
	nsRun(t, "ip", "link", "set", "edr-dhcp-s", "up")
	nsRun(t, "ip", "netns", "exec", peer, "ip", "link", "set", "lo", "up")
	nsRun(t, "ip", "netns", "exec", peer, "ip", "link", "set", "edr-dhcp-p", "up")
	type packet struct {
		name, source, target string
		inbound              bool
		sport, dport         int
		allowed              bool
	}
	var cases []packet
	for _, v6 := range []bool{false, true} {
		subject, remote, client, server := "192.0.2.2", "192.0.2.1", 68, 67
		family := "v4"
		if v6 {
			subject, remote, client, server, family = "2001:db8:1::2", "2001:db8:1::1", 546, 547, "v6"
		}
		for _, inbound := range []bool{false, true} {
			source, target, sport, dport, direction := subject, remote, client, server, "output"
			if inbound {
				source, target, sport, dport, direction = remote, subject, server, client, "input"
			}
			cases = append(cases, packet{family + " " + direction + " client policy", source, target, inbound, sport, dport, true})
			cases = append(cases, packet{family + " " + direction + " reversed ports", source, target, inbound, dport, sport, false})
			otherClient, otherServer := 546, 547
			if v6 {
				otherClient, otherServer = 68, 67
			}
			if inbound {
				otherClient, otherServer = otherServer, otherClient
			}
			cases = append(cases, packet{family + " " + direction + " other-family ports", source, target, inbound, otherClient, otherServer, false})
		}
	}
	for _, c := range cases {
		nsAssert(t, "baseline DHCP "+c.name, nsDHCPDatagram(t, peer, c.inbound, c.source, c.target, c.sport, c.dport), true)
	}
	h, err := newLinuxHostIsolator(filepath.Join(t.TempDir(), "runtime"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := h.Release(context.Background()); err != nil {
			t.Errorf("DHCP isolation release: %v", err)
		}
	})
	status, err := h.Isolate(context.Background(), IsolationPolicy{Management: []netip.Addr{netip.MustParseAddr("198.51.100.7")}, Duration: 30 * time.Second})
	if err != nil || !status.Active || !status.Verified {
		t.Fatalf("DHCP isolation installation %+v: %v", status, err)
	}
	for _, c := range cases {
		nsAssert(t, "isolated DHCP "+c.name, nsDHCPDatagram(t, peer, c.inbound, c.source, c.target, c.sport, c.dport), c.allowed)
	}
}
