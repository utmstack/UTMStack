//go:build linux

package netblock

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/fs"
	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/sys/unix"
)

const nativeNetworkBase = "/opt/openedr-network"

var nativeManagementPeers = []string{"192.0.2.1", "198.51.100.1", "203.0.113.1", "2001:db8:1::1", "2001:db8:2::1", "2001:db8:3::1"}

func nativeManagementDNS(t *testing.T, stage *atomic.Int32, queries *atomic.Uint64) {
	t.Helper()
	listener, err := net.ListenPacket("udp4", "192.0.2.53:53")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	t.Cleanup(func() { listener.Close(); <-done })
	go func() {
		defer close(done)
		var buf [1024]byte
		for {
			n, addr, err := listener.ReadFrom(buf[:])
			if err != nil {
				return
			}
			var parser dnsmessage.Parser
			header, err := parser.Start(buf[:n])
			if err != nil {
				continue
			}
			questions, err := parser.AllQuestions()
			if err != nil || len(questions) != 1 || questions[0].Name.String() != "management.openedr.test." {
				continue
			}
			queries.Add(1)
			q := questions[0]
			response := dnsmessage.Message{Header: dnsmessage.Header{ID: header.ID, Response: true, Authoritative: true, RecursionDesired: header.RecursionDesired, RecursionAvailable: true}, Questions: questions}
			current := stage.Load()
			if current == 2 {
				response.Header.RCode = dnsmessage.RCodeNameError
			} else {
				a4, a6 := "192.0.2.1", "2001:db8:1::1"
				if current != 0 {
					a4, a6 = "198.51.100.1", "2001:db8:2::1"
				}
				resource := dnsmessage.Resource{Header: dnsmessage.ResourceHeader{Name: q.Name, Type: q.Type, Class: dnsmessage.ClassINET, TTL: 0}}
				switch q.Type {
				case dnsmessage.TypeA:
					resource.Body = &dnsmessage.AResource{A: netip.MustParseAddr(a4).As4()}
					response.Answers = append(response.Answers, resource)
				case dnsmessage.TypeAAAA:
					resource.Body = &dnsmessage.AAAAResource{AAAA: netip.MustParseAddr(a6).As16()}
					response.Answers = append(response.Answers, resource)
				}
			}
			data, err := response.Pack()
			if err == nil {
				_, _ = listener.WriteTo(data, addr)
			}
		}
	}()
}

func nativeNetworkStatus(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(nativeNetworkBase, "status.json"))
	if err != nil {
		return nil
	}
	var status map[string]any
	if json.Unmarshal(data, &status) != nil {
		return nil
	}
	return status
}

func nativeNetworkPhase(t *testing.T, label string, want []bool, degraded bool, after time.Time, pid int) {
	t.Helper()
	deadline := time.Now().Add(48 * time.Second)
	for {
		matches := true
		for i, host := range nativeManagementPeers {
			if nsCanDial(host) != want[i] {
				matches = false
			}
		}
		status := nativeNetworkStatus(t)
		updated, _ := time.Parse(time.RFC3339Nano, fmt.Sprint(status["updated_at"]))
		fresh := updated.After(after) && status["process_id"] == float64(pid) && status["engine_healthy"] == true
		if matches && fresh && status["blocklist_enforcement_degraded"] == degraded {
			raw, _ := json.Marshal(map[string]any{"phase": label, "status": status, "expected_reachable": want})
			t.Logf("NATIVE_NETWORK_RESULT %s", raw)
			for i, host := range nativeManagementPeers {
				nsAssert(t, label+" "+host, nsCanDial(host), want[i])
			}
			gate := exec.Command(filepath.Join(nativeNetworkBase, "dependency-native.test"), "-test.v", "-test.run=^TestNativeLinuxNetworkReadinessFixture$", "-test.timeout=10s")
			wantReady := "1"
			if degraded {
				wantReady = "0"
			}
			gate.Env = append(os.Environ(), "OPENEDR_NATIVE_NETWORK_READINESS=1", "OPENEDR_NETWORK_WANT_READY="+wantReady)
			if out, err := gate.CombinedOutput(); err != nil {
				t.Fatalf("actual release network readiness: %v %s", err, out)
			} else {
				t.Logf("%s", out)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s: management DNS change did not reach live OS policy without config rewrite; packets_match=%t status=%v", label, matches, status)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestNFTNamespaceServiceRefreshesManagementDNS(t *testing.T) {
	nftNamespaceGate(t)
	if fs.GetExecutablePath() != nativeNetworkBase {
		t.Fatal("service DNS fixture requires exact owned install directory")
	}
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"edr.json", "edr.db", "status.json"} {
		if _, err := os.Lstat(filepath.Join(nativeNetworkBase, name)); !os.IsNotExist(err) {
			t.Fatal("refusing preexisting service state", name, err)
		}
	}
	binary := filepath.Join(nativeNetworkBase, "utmstack_edr_linux_"+runtime.GOARCH)
	if st, err := os.Stat(binary); err != nil || !st.Mode().IsRegular() {
		t.Fatal("owned service executable missing", err)
	}
	peer := fmt.Sprintf("edr-dns-%d", os.Getpid())
	nsRun(t, "ip", "link", "set", "lo", "up")
	nsRun(t, "ip", "netns", "add", peer)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = exec.CommandContext(ctx, "ip", "netns", "delete", peer).Run()
	})
	nsRun(t, "ip", "link", "add", "edr-dns-s", "type", "veth", "peer", "name", "edr-dns-p")
	nsRun(t, "ip", "link", "set", "edr-dns-p", "netns", peer)
	for _, address := range []string{"192.0.2.2/24", "192.0.2.53/24", "198.51.100.2/24", "203.0.113.2/24", "2001:db8:1::2/64", "2001:db8:2::2/64", "2001:db8:3::2/64"} {
		nsRun(t, "ip", "addr", "add", address, "dev", "edr-dns-s", "nodad")
	}
	for _, address := range []string{"192.0.2.1/24", "198.51.100.1/24", "203.0.113.1/24", "2001:db8:1::1/64", "2001:db8:2::1/64", "2001:db8:3::1/64"} {
		nsRun(t, "ip", "netns", "exec", peer, "ip", "addr", "add", address, "dev", "edr-dns-p", "nodad")
	}
	nsRun(t, "ip", "link", "set", "edr-dns-s", "up")
	nsRun(t, "ip", "netns", "exec", peer, "ip", "link", "set", "lo", "up")
	nsRun(t, "ip", "netns", "exec", peer, "ip", "link", "set", "edr-dns-p", "up")
	serverScript := `import socket,select
servers=[]
for family,host in [(socket.AF_INET,'0.0.0.0'),(socket.AF_INET6,'::')]:
 s=socket.socket(family,socket.SOCK_STREAM);s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
 if family==socket.AF_INET6:s.setsockopt(socket.IPPROTO_IPV6,socket.IPV6_V6ONLY,1)
 s.bind((host,18080));s.listen(64);servers.append(s)
while True:
 for s in select.select(servers,[],[],1)[0]:
  c,a=s.accept();c.close()
`
	server := exec.Command("ip", "netns", "exec", peer, "python3", "-u", "-c", serverScript)
	server.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = syscall.Kill(-server.Process.Pid, syscall.SIGKILL); _ = server.Wait() })
	deadline := time.Now().Add(4 * time.Second)
	for !nsCanDial("2001:db8:1::1") {
		if time.Now().After(deadline) {
			t.Fatal("owned IPv6 peer unavailable")
		}
		time.Sleep(50 * time.Millisecond)
	}
	for _, host := range nativeManagementPeers {
		nsAssert(t, "unfiltered DNS fixture "+host, nsCanDial(host), true)
	}
	var stage atomic.Int32
	var queries atomic.Uint64
	nativeManagementDNS(t, &stage, &queries)
	resolv := filepath.Join(t.TempDir(), "resolv.conf")
	if err := os.WriteFile(resolv, []byte("nameserver 192.0.2.53\noptions timeout:1 attempts:1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := unix.Mount(resolv, "/etc/resolv.conf", "", unix.MS_BIND, ""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = unix.Unmount("/etc/resolv.conf", 0) })
	db := filepath.Join(nativeNetworkBase, "owned-db")
	if err := os.Mkdir(db, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(db, "owned.ndb"), []byte("Safe.Network.Fixture:0:*:4f50454e4544525f4f574e45445f4e4554574f524b\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Enabled = true
	cfg.Server = "management.openedr.test"
	cfg.SignatureDBDir = db
	cfg.Ransomware.Enabled = false
	cfg.ScheduledScan.Enabled = false
	off := false
	cfg.Sensors = config.Sensors{FileWatcher: &off, ProcessGuard: &off, AMSI: &off, Behavioral: &off}
	cfg.LinuxProtection.NetworkEnforcement = true
	cfg.ScanConcurrency = 1
	cfg.Blocklist.Direction = "out"
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	configBefore, err := os.ReadFile(config.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	configHash := sha256.Sum256(configBefore)
	info, err := os.Stat(config.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	configMtime := info.ModTime()
	t.Cleanup(func() {
		after, err := os.ReadFile(config.ConfigFile)
		st, statErr := os.Stat(config.ConfigFile)
		if err != nil || statErr != nil || sha256.Sum256(after) != configHash || !st.ModTime().Equal(configMtime) {
			t.Error("DNS fixture rewrote configuration bytes or modification time")
		} else {
			t.Logf("NATIVE_NETWORK_CONFIG_UNCHANGED sha256=%x mtime=%s", configHash, configMtime.Format(time.RFC3339Nano))
		}
	})
	c, err := cache.Open(config.DBFile)
	if err != nil {
		t.Fatal(err)
	}
	var indicators []cache.IndicatorRecord
	for _, host := range nativeManagementPeers {
		indicators = append(indicators, cache.IndicatorRecord{Value: host, Type: "ip", Level: 1})
	}
	err = c.ReplaceIndicators(indicators)
	c.Close()
	if err != nil {
		t.Fatal(err)
	}
	log, err := os.Create(filepath.Join(nativeNetworkBase, "service-dns-output.txt"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary)
	cmd.Env = append(os.Environ(), "GODEBUG=netdns=go")
	cmd.Stdout = log
	cmd.Stderr = log
	if err := cmd.Start(); err != nil {
		log.Close()
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		_ = cmd.Process.Signal(syscall.SIGTERM)
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("service exit: %v", err)
			}
		case <-time.After(6 * time.Second):
			_ = cmd.Process.Kill()
			<-done
			t.Error("owned network service did not stop")
		}
		log.Close()
	}
	t.Cleanup(func() {
		stop()
		if t.Failed() {
			data, _ := os.ReadFile(filepath.Join(nativeNetworkBase, "service-dns-output.txt"))
			t.Logf("owned service output:\n%s", data)
		}
	})
	nativeNetworkPhase(t, "initial", []bool{true, false, false, true, false, false}, false, time.Time{}, cmd.Process.Pid)
	changed := time.Now().UTC()
	stage.Store(1)
	nativeNetworkPhase(t, "rotated", []bool{false, true, false, false, true, false}, false, changed, cmd.Process.Pid)
	changed = time.Now().UTC()
	stage.Store(2)
	nativeNetworkPhase(t, "resolution-unavailable", []bool{true, true, true, true, true, true}, true, changed, cmd.Process.Pid)
	changed = time.Now().UTC()
	stage.Store(3)
	nativeNetworkPhase(t, "resolution-recovered", []bool{false, true, false, false, true, false}, false, changed, cmd.Process.Pid)
	after, err := os.ReadFile(config.ConfigFile)
	if err != nil || sha256.Sum256(after) != configHash {
		t.Fatal("DNS control rewrote configuration", err)
	}
	t.Logf("NATIVE_NETWORK_DNS_QUERIES %d", queries.Load())
	stop()
	raw := nsRun(t, "nft", "-j", "list", "tables")
	var tables struct {
		Objects []map[string]json.RawMessage `json:"nftables"`
	}
	if json.Unmarshal(raw, &tables) != nil {
		t.Fatal("invalid cleanup inventory")
	}
	for _, object := range tables.Objects {
		if _, ok := object["table"]; ok {
			t.Fatalf("owned service retained a kernel table: %s", raw)
		}
	}
	if data, _ := os.ReadFile(filepath.Join(nativeNetworkBase, "status.json")); !bytes.Contains(data, []byte(`"running":false`)) && !strings.Contains(string(data), `"running": false`) {
		t.Fatal("stopped service status missing")
	}
}
