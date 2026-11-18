//go:build linux

package netblock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

type captureNamespaceNFT struct {
	nftOperations
	lastRead, lastPlan []byte
}

func (c *captureNamespaceNFT) Run(ctx context.Context, input []byte, args ...string) ([]byte, error) {
	data, err := c.nftOperations.Run(ctx, input, args...)
	if strings.Contains(strings.Join(args, " "), "list table inet") {
		c.lastRead = append([]byte(nil), data...)
	}
	if bytes.Contains(input, []byte("{\n comment")) {
		c.lastPlan = append([]byte(nil), input...)
	}
	return data, err
}

// Inject exactly one owned post-commit read failure. Actual kernel writes,
// rollback reads and restoration still use the production Linux command runner.
type failReadAfterNamespaceCommit struct {
	nftOperations
	armed, failRead bool
}

func (f *failReadAfterNamespaceCommit) Run(ctx context.Context, input []byte, args ...string) ([]byte, error) {
	if f.failRead && len(input) == 0 && strings.Join(args, " ") == "-j -n list tables" {
		f.failRead = false
		return nil, fmt.Errorf("OWNED_NATIVE_VERIFY_FAILURE")
	}
	data, err := f.nftOperations.Run(ctx, input, args...)
	if err == nil && f.armed && len(input) > 0 {
		f.armed, f.failRead = false, true
	}
	return data, err
}

// This suite must never run against a workstation or the VM's initial network.
// The disposable-VM marker, explicit opt-in and separate mount+network namespaces
// are independent gates. Ordinary `go test` skips every real firewall operation.
func nftNamespaceGate(t *testing.T) {
	t.Helper()
	if os.Getenv("OPENEDR_NFT_NAMESPACE_TEST") != "1" {
		t.Skip("requires explicitly opted-in disposable VM network namespace")
	}
	if os.Geteuid() != 0 {
		t.Fatal("namespace firewall test requires root")
	}
	if data, err := os.ReadFile("/run/openedr-disposable-lab"); err != nil || strings.TrimSpace(string(data)) != "OPENEDR_DISPOSABLE_LAB_ONLY" {
		t.Fatal("disposable VM marker missing")
	}
	for _, ns := range []string{"net", "mnt"} {
		self, err := os.Readlink("/proc/self/ns/" + ns)
		if err != nil {
			t.Fatal(err)
		}
		initial, err := os.Readlink("/proc/1/ns/" + ns)
		if err != nil || self == initial {
			t.Fatalf("must run in a separate %s namespace", ns)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	data, err := exec.CommandContext(ctx, "ip", "-j", "link", "show").CombinedOutput()
	if err != nil {
		t.Fatalf("ip inventory %v %s", err, data)
	}
	var links []struct {
		Name string `json:"ifname"`
	}
	if json.Unmarshal(data, &links) != nil || len(links) != 1 || links[0].Name != "lo" {
		t.Fatal("test namespace must begin with loopback only")
	}
}

func nsRun(t *testing.T, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	data, err := exec.CommandContext(ctx, args[0], args[1:]...).CombinedOutput()
	if err != nil {
		t.Fatalf("namespace fixture command %v: %v\n%s", args, err, data)
	}
	return data
}

func nsCanDial(addr string) bool {
	return nsCanDialPort(addr, "18080")
}

func nsCanDialPort(addr, port string) bool {
	c, err := net.DialTimeout("tcp", net.JoinHostPort(addr, port), 300*time.Millisecond)
	if err != nil {
		return false
	}
	defer c.Close()
	return true
}

func nsCanUDP(addr, port string) bool {
	c, err := net.DialTimeout("udp", net.JoinHostPort(addr, port), 300*time.Millisecond)
	if err != nil {
		return false
	}
	defer c.Close()
	return nsTCPRoundTrip(c)
}

func nsTCPRoundTrip(c net.Conn) bool {
	_ = c.SetDeadline(time.Now().Add(300 * time.Millisecond))
	if _, err := c.Write([]byte("safe")); err != nil {
		return false
	}
	var b [4]byte
	_, err := io.ReadFull(c, b[:])
	return err == nil && string(b[:]) == "safe"
}

func nsPeerUDP(peer, source, target string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	script := `import socket,sys
s=socket.socket(socket.AF_INET6 if ':' in sys.argv[1] else socket.AF_INET,socket.SOCK_DGRAM)
s.settimeout(0.3);s.bind((sys.argv[1],0))
try:
 s.sendto(b'safe',(sys.argv[2],18083))
 data,_=s.recvfrom(16)
 sys.exit(0 if data==b'safe' else 1)
except OSError:
 sys.exit(1)
finally:
 s.close()
`
	return exec.CommandContext(ctx, "ip", "netns", "exec", peer, "python3", "-c", script, source, target).Run() == nil
}

func nsICMP(peer, source, target string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	script := `import socket,struct,sys,time,os
v6=':' in sys.argv[2]
s=socket.socket(socket.AF_INET6 if v6 else socket.AF_INET,socket.SOCK_RAW,socket.IPPROTO_ICMPV6 if v6 else socket.IPPROTO_ICMP)
s.settimeout(0.3)
if sys.argv[1]:s.bind((sys.argv[1],0))
ident=os.getpid()&65535
packet=struct.pack('!BBHHH',128 if v6 else 8,0,0,ident,1)+b'safe'
if not v6:
 words=struct.unpack('!%dH'%(len(packet)//2),packet)
 total=sum(words)
 while total>>16:total=(total&65535)+(total>>16)
 packet=packet[:2]+struct.pack('!H',(~total)&65535)+packet[4:]
try:
 s.sendto(packet,(sys.argv[2],0))
 end=time.monotonic()+0.3
 while time.monotonic()<end:
  data,_=s.recvfrom(2048)
  offset=0 if v6 else (data[0]&15)*4
  if len(data)>=offset+8:
   typ,code,check,found,seq=struct.unpack('!BBHHH',data[offset:offset+8])
   if typ==(129 if v6 else 0) and found==ident:sys.exit(0)
except OSError:pass
finally:s.close()
sys.exit(1)
`
	args := []string{"python3", "-c", script, source, target}
	if peer != "" {
		args = append([]string{"ip", "netns", "exec", peer}, args...)
	}
	return exec.CommandContext(ctx, args[0], args[1:]...).Run() == nil
}

func nsPeerDial(peer, source, target string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	script := `import socket,sys
s=socket.socket(socket.AF_INET6 if ':' in sys.argv[1] else socket.AF_INET,socket.SOCK_STREAM)
s.settimeout(0.3)
s.bind((sys.argv[1],0))
try:
 s.connect((sys.argv[2],18081))
except OSError:
 sys.exit(1)
finally:
 s.close()
`
	return exec.CommandContext(ctx, "ip", "netns", "exec", peer, "python3", "-c", script, source, target).Run() == nil
}

func nsAssert(t *testing.T, what string, got, want bool) {
	t.Helper()
	if got != want {
		t.Fatalf("%s connectivity=%v want %v", what, got, want)
	}
	t.Logf("packet check %s: reachable=%v", what, got)
}

func TestNFTNamespaceEnforcementAndIsolation(t *testing.T) {
	nftNamespaceGate(t)
	peer := fmt.Sprintf("edr-peer-%d", os.Getpid())
	nsRun(t, "ip", "link", "set", "lo", "up")
	nsRun(t, "ip", "netns", "add", peer)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = exec.CommandContext(ctx, "ip", "netns", "delete", peer).Run()
	})
	nsRun(t, "ip", "link", "add", "edr-subject", "type", "veth", "peer", "name", "edr-peer")
	nsRun(t, "ip", "link", "set", "edr-peer", "netns", peer)
	for _, addr := range []string{"192.0.2.2/24", "198.51.100.2/24", "2001:db8:1::2/64", "2001:db8:2::2/64"} {
		nsRun(t, "ip", "addr", "add", addr, "dev", "edr-subject")
	}
	for _, addr := range []string{"192.0.2.1/24", "198.51.100.1/24", "2001:db8:1::1/64", "2001:db8:2::1/64"} {
		nsRun(t, "ip", "netns", "exec", peer, "ip", "addr", "add", addr, "dev", "edr-peer")
	}
	nsRun(t, "ip", "link", "set", "edr-subject", "up")
	nsRun(t, "ip", "netns", "exec", peer, "ip", "link", "set", "lo", "up")
	nsRun(t, "ip", "netns", "exec", peer, "ip", "link", "set", "edr-peer", "up")
	// A local inert TCP server in the peer namespace. No public interface, sample
	// execution, DNS submission or external endpoint is involved.
	serverScript := `import socket,select
servers=[]
listeners=set()
datagrams=set()
for family,host in [(socket.AF_INET,'0.0.0.0'),(socket.AF_INET6,'::')]:
 for port,kind in [(18080,socket.SOCK_STREAM),(53,socket.SOCK_STREAM),(18082,socket.SOCK_DGRAM),(53,socket.SOCK_DGRAM)]:
  s=socket.socket(family,kind)
  s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
  if family==socket.AF_INET6:s.setsockopt(socket.IPPROTO_IPV6,socket.IPV6_V6ONLY,1)
  s.bind((host,port))
  if kind==socket.SOCK_STREAM:s.listen(64);listeners.add(s)
  else:datagrams.add(s)
  servers.append(s)
while True:
 for s in select.select(servers,[],[],1)[0]:
  if s in listeners:
   c,a=s.accept();servers.append(c)
  elif s in datagrams:
   data,a=s.recvfrom(2048);s.sendto(data,a)
  else:
   try:
    data=s.recv(2048)
    if data:s.sendall(data)
    else:servers.remove(s);s.close()
   except OSError:
    servers.remove(s);s.close()
`
	server := exec.Command("ip", "netns", "exec", peer, "python3", "-u", "-c", serverScript)
	server.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = syscall.Kill(-server.Process.Pid, syscall.SIGKILL); _ = server.Wait() })
	for _, host := range []string{"0.0.0.0", "::"} {
		protocol := "tcp4"
		if host == "::" {
			protocol = "tcp6"
		}
		l, err := net.Listen(protocol, net.JoinHostPort(host, "18081"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = l.Close() })
		go func() {
			for {
				c, err := l.Accept()
				if err != nil {
					return
				}
				_ = c.Close()
			}
		}()
		udpProtocol := "udp4"
		if host == "::" {
			udpProtocol = "udp6"
		}
		udp, err := net.ListenPacket(udpProtocol, net.JoinHostPort(host, "18083"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = udp.Close() })
		go func() {
			var buf [2048]byte
			for {
				n, addr, err := udp.ReadFrom(buf[:])
				if err != nil {
					return
				}
				_, _ = udp.WriteTo(buf[:n], addr)
			}
		}()
	}
	deadline := time.Now().Add(4 * time.Second)
	for !nsCanDial("2001:db8:1::1") {
		if time.Now().After(deadline) {
			t.Fatal("IPv6 peer fixture unavailable")
		}
		time.Sleep(50 * time.Millisecond)
	}
	for _, a := range []string{"192.0.2.1", "198.51.100.1", "2001:db8:1::1", "2001:db8:2::1"} {
		nsAssert(t, "baseline "+a, nsCanDial(a), true)
	}
	for _, a := range []string{"192.0.2.1", "2001:db8:1::1"} {
		nsAssert(t, "baseline UDP "+a, nsCanUDP(a, "18082"), true)
		nsAssert(t, "baseline ICMP "+a, nsICMP("", "", a), true)
	}
	runtimeDir := filepath.Join(t.TempDir(), "runtime")
	b, err := newLinuxBlocker(fmt.Sprintf("NamespaceTest-%d", os.Getpid()), runtimeDir)
	if err != nil {
		t.Fatal(err)
	}
	captured := &captureNamespaceNFT{nftOperations: b.c.ops}
	b.c.ops = captured
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("last rendered plan:\n%s\nlast kernel table readback:\n%s", captured.lastPlan, captured.lastRead)
		}
	})
	t.Cleanup(func() {
		if err := b.Close(); err != nil {
			t.Errorf("backend cleanup: %v", err)
		}
	})
	management := []netip.Addr{netip.MustParseAddr("198.51.100.1"), netip.MustParseAddr("2001:db8:2::1")}
	if err = b.SetSafetyPolicy(SafetyPolicy{Management: management}); err != nil {
		t.Fatal(err)
	}
	targets := []netip.Addr{netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("2001:db8:1::1")}
	existing, err := net.DialTimeout("tcp", "192.0.2.1:18080", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = existing.Close() })
	nsAssert(t, "existing TCP before outbound deny", nsTCPRoundTrip(existing), true)
	if err = b.ApplyDelta(targets, nil, nil, nil, "out"); err != nil {
		t.Fatal(err)
	}
	for _, a := range targets {
		nsAssert(t, "out-only deny "+a.String(), nsCanDial(a.String()), false)
		nsAssert(t, "out-only UDP deny "+a.String(), nsCanUDP(a.String(), "18082"), false)
		nsAssert(t, "out-only ICMP deny "+a.String(), nsICMP("", "", a.String()), false)
	}
	nsAssert(t, "existing TCP after outbound deny", nsTCPRoundTrip(existing), false)
	_ = existing.Close()
	nsAssert(t, "out-only permits inbound", nsPeerDial(peer, "192.0.2.1", "192.0.2.2"), true)
	nsAssert(t, "out-only permits inbound UDP", nsPeerUDP(peer, "192.0.2.1", "192.0.2.2"), true)
	nsAssert(t, "out-only permits inbound ICMP", nsICMP(peer, "192.0.2.1", "192.0.2.2"), true)
	if err = b.ApplyDelta(targets, nil, nil, nil, "in"); err != nil {
		t.Fatal(err)
	}
	for _, a := range targets {
		nsAssert(t, "in-only permits outbound "+a.String(), nsCanDial(a.String()), true)
		nsAssert(t, "in-only permits outbound UDP "+a.String(), nsCanUDP(a.String(), "18082"), true)
		nsAssert(t, "in-only permits outbound ICMP "+a.String(), nsICMP("", "", a.String()), true)
	}
	nsAssert(t, "in-only blocks inbound IPv4", nsPeerDial(peer, "192.0.2.1", "192.0.2.2"), false)
	nsAssert(t, "in-only blocks inbound IPv6", nsPeerDial(peer, "2001:db8:1::1", "2001:db8:1::2"), false)
	nsAssert(t, "in-only blocks inbound UDP", nsPeerUDP(peer, "192.0.2.1", "192.0.2.2"), false)
	nsAssert(t, "in-only blocks inbound ICMP", nsICMP(peer, "192.0.2.1", "192.0.2.2"), false)
	if err = b.ApplyDelta(nil, targets, []netip.Prefix{netip.MustParsePrefix("192.0.2.0/24"), netip.MustParsePrefix("2001:db8:1::/64"), netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("2001:db8:2::/64")}, nil, "both"); err != nil {
		t.Fatal(err)
	}
	for _, a := range targets {
		nsAssert(t, "prefix deny "+a.String(), nsCanDial(a.String()), false)
	}
	for _, a := range management {
		nsAssert(t, "management exception within denied prefix "+a.String(), nsCanDial(a.String()), true)
	}
	snapshot, err := b.c.snapshot(contextWithNFTDeadline(t))
	if err != nil {
		t.Fatal(err)
	}
	var drops float64
	for _, r := range snapshot.Rules {
		if expr, ok := r["expr"].([]any); ok {
			for _, v := range expr {
				if m, ok := v.(map[string]any); ok {
					if count, ok := m["counter"].(map[string]any); ok {
						if n, ok := count["packets"].(float64); ok {
							drops += n
						}
					}
				}
			}
		}
	}
	if drops < 2 {
		t.Fatalf("kernel drop counters did not account for probes: %v", drops)
	}
	t.Logf("verified kernel denied-packet counter=%v", drops)
	if err = b.Reset(); err != nil {
		t.Fatal(err)
	}
	for _, a := range targets {
		nsAssert(t, "reset restored "+a.String(), nsCanDial(a.String()), true)
	}
	// Separate private runtime token/lock, with a fresh constructor exercising
	// cross-process-style status and release on the same actual owned table.
	isolationRuntime := filepath.Join(t.TempDir(), "runtime")
	h, err := newLinuxHostIsolator(isolationRuntime)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := h.Release(context.Background()); err != nil {
			t.Errorf("isolation cleanup: %v", err)
		}
	})
	isolationExisting, err := net.DialTimeout("tcp", "192.0.2.1:18080", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = isolationExisting.Close() })
	nsAssert(t, "existing TCP before isolation", nsTCPRoundTrip(isolationExisting), true)
	isolationStarted := time.Now()
	status, err := h.Isolate(context.Background(), IsolationPolicy{Management: management, Resolvers: targets, Duration: 30 * time.Second})
	if err != nil || !status.Active || !status.Verified {
		t.Fatalf("isolation status %+v: %v", status, err)
	}
	initialDeadline := status.ExpiresAt
	remaining := time.Until(initialDeadline)
	if remaining < 25*time.Second || remaining > 30*time.Second {
		t.Fatalf("native isolation lease units: remaining=%s status=%+v", remaining, status)
	}
	t.Logf("native lease seconds: requested=30 remaining=%s deadline=%s elapsed=%s", remaining, initialDeadline.Format(time.RFC3339Nano), time.Since(isolationStarted))
	for _, a := range targets {
		nsAssert(t, "isolated target "+a.String(), nsCanDial(a.String()), false)
		nsAssert(t, "isolated UDP target "+a.String(), nsCanUDP(a.String(), "18082"), false)
		nsAssert(t, "isolated ICMP target "+a.String(), nsICMP("", "", a.String()), false)
		nsAssert(t, "isolated resolver TCP53 "+a.String(), nsCanDialPort(a.String(), "53"), true)
		nsAssert(t, "isolated resolver UDP53 "+a.String(), nsCanUDP(a.String(), "53"), true)
	}
	nsAssert(t, "existing TCP cannot bypass isolation", nsTCPRoundTrip(isolationExisting), false)
	_ = isolationExisting.Close()
	for _, a := range management {
		nsAssert(t, "isolated management "+a.String(), nsCanDial(a.String()), true)
	}
	nsAssert(t, "isolated inbound target", nsPeerDial(peer, "192.0.2.1", "192.0.2.2"), false)
	nsAssert(t, "isolated inbound management", nsPeerDial(peer, "198.51.100.1", "198.51.100.2"), true)
	fresh, err := newLinuxHostIsolator(isolationRuntime)
	if err != nil {
		t.Fatal(err)
	}
	status, err = fresh.Status(context.Background())
	if err != nil || !status.Active || !status.Verified {
		t.Fatalf("fresh status %+v %v", status, err)
	}
	deadlineShift := status.ExpiresAt.Sub(initialDeadline)
	if deadlineShift < -3*time.Second || deadlineShift > time.Second {
		t.Fatalf("fresh status changed original lease deadline by %s: %+v", deadlineShift, status)
	}
	t.Logf("native fresh status seconds: remaining=%s original_deadline_shift=%s", time.Until(status.ExpiresAt), deadlineShift)
	injected := &failReadAfterNamespaceCommit{nftOperations: fresh.controller.ops, armed: true}
	fresh.controller.ops = injected
	if _, err = fresh.Isolate(context.Background(), IsolationPolicy{Management: management, Resolvers: targets, Duration: 20 * time.Second}); err == nil || !strings.Contains(err.Error(), "OWNED_NATIVE_VERIFY_FAILURE") {
		t.Fatalf("fresh replacement did not preserve owned post-commit failure: %v", err)
	}
	if injected.armed || injected.failRead {
		t.Fatal("owned post-commit failure was not consumed exactly once")
	}
	rolledBack, err := fresh.Status(context.Background())
	if err != nil || !rolledBack.Active || !rolledBack.Verified {
		t.Fatalf("fresh rollback status %+v %v", rolledBack, err)
	}
	rollbackShift := rolledBack.ExpiresAt.Sub(initialDeadline)
	if rollbackShift < -3*time.Second || rollbackShift > time.Second {
		t.Fatalf("fresh rollback renewed or collapsed original deadline by %s", rollbackShift)
	}
	t.Logf("native fresh post-commit rollback seconds: remaining=%s original_deadline_shift=%s", time.Until(rolledBack.ExpiresAt), rollbackShift)
	for _, a := range targets {
		nsAssert(t, "fresh rollback retained isolation "+a.String(), nsCanDial(a.String()), false)
	}
	replaced, err := fresh.Isolate(context.Background(), IsolationPolicy{Management: management, Resolvers: targets, Duration: 20 * time.Second})
	if err != nil || !replaced.Active || !replaced.Verified {
		t.Fatalf("fresh replacement status %+v %v", replaced, err)
	}
	remaining = time.Until(replaced.ExpiresAt)
	if remaining < 15*time.Second || remaining > 20*time.Second {
		t.Fatalf("fresh replacement lease units: remaining=%s status=%+v", remaining, replaced)
	}
	t.Logf("native fresh replacement seconds: requested=20 remaining=%s deadline=%s", remaining, replaced.ExpiresAt.Format(time.RFC3339Nano))
	for _, a := range targets {
		nsAssert(t, "fresh replacement isolated target "+a.String(), nsCanDial(a.String()), false)
	}
	if err = fresh.Release(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, a := range targets {
		nsAssert(t, "release restored "+a.String(), nsCanDial(a.String()), true)
	}
	shortStarted := time.Now()
	shortStatus, err := h.Isolate(context.Background(), IsolationPolicy{Management: management, Duration: 6 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	remaining = time.Until(shortStatus.ExpiresAt)
	if !shortStatus.Active || !shortStatus.Verified || remaining < 3*time.Second || remaining > 6*time.Second {
		t.Fatalf("short native lease units: remaining=%s status=%+v", remaining, shortStatus)
	}
	t.Logf("native kernel expiry seconds: requested=6 initial_remaining=%s", remaining)
	nsAssert(t, "lease initially blocks", nsCanDial("192.0.2.1"), false)
	time.Sleep(7 * time.Second)
	t.Logf("native kernel expiry elapsed=%s", time.Since(shortStarted))
	for _, a := range targets {
		nsAssert(t, "kernel expiry restored "+a.String(), nsCanDial(a.String()), true)
	}
	status, err = fresh.Status(context.Background())
	if err != nil || status.Active || !status.Verified {
		t.Fatalf("expired status %+v %v", status, err)
	}
	t.Logf("native expired fresh status: active=%t verified=%t deadline=%s", status.Active, status.Verified, status.ExpiresAt.Format(time.RFC3339Nano))
}

func contextWithNFTDeadline(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), nftOperationTimeout)
	t.Cleanup(cancel)
	return ctx
}
