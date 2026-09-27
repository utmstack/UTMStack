//go:build linux

// Probe: does this kernel support the netlink process connector (CN_PROC)
// and does it deliver exec events? Run as root:
//
//	sudo go test -run TestCNProcProbe -v ./edr/procwatch
//
// Layouts (include/uapi/linux/cn_proc.h, kernel 6.8+; verified with
// offsetof on this machine, 2026-09-27):
//
//	nlmsghdr   {len, type, flags, seq, pid}              16 bytes
//	cn_msg     {id.idx, id.major, seq, ack, len}         20 bytes
//	mcast_op   {mcast_op}                                 4 bytes  (subscribe, len=4 form)
//	proc_event {what, cpu, timestamp_ns, union event_data} 40 bytes
//
// A PROC_EVENT_EXEC event is 76 bytes minimum; in the received buffer
// what is at 36, exec.process_pid at 52, exec.process_tgid at 56.
// There is NO ppid in the event (the ancient layout had one; the modern
// union does not).
//
// Subscribe uses the len=4 form (mcast_op only): accepted by every kernel.
// The len=8 proc_input form (added in 6.5) is silently dropped by older
// kernels. With len=4 the kernel delivers ALL event types, so the probe
// must filter what==PROC_EVENT_EXEC in userspace.
package procwatch

import (
	"encoding/binary"
	"os/exec"
	"testing"

	"golang.org/x/sys/unix"
)

func TestCNProcProbe(t *testing.T) {
	if testing.Short() {
		t.Skip("probe needs a live kernel + root")
	}
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_CONNECTOR)
	if err != nil {
		t.Fatalf("socket(NETLINK_CONNECTOR): %v (kernel without CONFIG_NET_CONNECTOR?)", err)
	}
	defer unix.Close(fd)

	// struct sockaddr_nl: { u16 family; u16 pad; u32 pid; u32 groups }
	// Groups MUST be CN_IDX_PROC (1): the kernel delivers events via
	// netlink_broadcast(group=CN_IDX_PROC), which only reaches sockets
	// that joined that multicast group at bind time. Groups=0 → zero
	// events, silently (the subscription itself still succeeds).
	sa := &unix.SockaddrNetlink{Family: unix.AF_NETLINK, Groups: 1} // CN_IDX_PROC
	if err := unix.Bind(fd, sa); err != nil {
		t.Fatalf("bind: %v", err)
	}
	t.Log("bind OK — kernel has NETLINK_CONNECTOR")

	// Subscribe: nlmsghdr + cn_msg(len=4) + mcast_op = 40 bytes.
	// The len=4 form (mcast_op only) is the ancient one, accepted by every
	// kernel; the len=8 proc_input form (mcast_op + event_type) was added
	// in 6.5 and is silently dropped by older kernels. With len=4 the
	// kernel delivers ALL event types (fork/exec/exit/...), so we filter
	// what==PROC_EVENT_EXEC in userspace below.
	buf := make([]byte, 40)
	binary.LittleEndian.PutUint32(buf[0:], 40)          // nlmsg_len
	binary.LittleEndian.PutUint16(buf[4:], 1)           // nlmsg_type (arbitrary for connector)
	binary.LittleEndian.PutUint32(buf[8:], 1)           // nlmsg_seq
	binary.LittleEndian.PutUint32(buf[16:], 1)          // cn_msg.id.idx  = CN_IDX_PROC
	binary.LittleEndian.PutUint32(buf[20:], 1)          // cn_msg.id.major = CN_VAL_PROC (routing is by the full {idx,val} pair)
	binary.LittleEndian.PutUint32(buf[24:], 1)          // cn_msg.seq
	binary.LittleEndian.PutUint32(buf[28:], 0)          // cn_msg.ack
	binary.LittleEndian.PutUint32(buf[32:], 4)          // cn_msg.len = sizeof(mcast_op)
	binary.LittleEndian.PutUint32(buf[36:], 1)          // mcast_op = PROC_CN_MCAST_LISTEN
	if err := unix.Sendto(fd, buf, 0, sa); err != nil {
		t.Fatalf("subscribe sendto: %v", err)
	}
	t.Log("CN_PROC PROC_CN_MCAST_LISTEN subscription sent (40-byte len=4 form, all event types)")

	// Spawn a short-lived process AFTER subscribing.
	_ = exec.Command("/bin/true").Run()

	// Read the event (poll-bounded so a silent kernel fails the probe, not hangs).
	_ = unix.SetNonblock(fd, true)
	var pfd [1]unix.PollFd
	pfd[0].Fd = int32(fd)
	pfd[0].Events = unix.POLLIN
	n, err := unix.Poll(pfd[:], 3000)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if n == 0 {
		t.Fatal("no CN_PROC event within 3s — exec notification not delivered (subscription form rejected?)")
	}
	rbuf := make([]byte, 4096)
	got, _, rerr := unix.Recvfrom(fd, rbuf, 0)
	if rerr != nil {
		t.Fatalf("recvfrom: %v", rerr)
	}
	if got < 76 {
		t.Fatalf("short event: %d bytes (need >= 76)", got)
	}
	// Walk the nlmsg chain; the first message may be the MCAST ack
	// (cn_msg.ack set) rather than an event.
	off := 0
	found := false
	for off+36 <= got {
		evLen := int(binary.LittleEndian.Uint32(rbuf[off:])) // nlmsg_len
		if evLen < 36 || off+evLen > got {
			break
		}
		ack := binary.LittleEndian.Uint32(rbuf[off+28:]) // cn_msg.ack
		if ack != 0 {
			t.Logf("MCAST ack received (cn_msg.ack=%d) — subscription accepted", ack)
		}
		if evLen >= 76 {
			what := binary.LittleEndian.Uint32(rbuf[off+36:])
			if what == 0x2 { // PROC_EVENT_EXEC
				pid := int(binary.LittleEndian.Uint32(rbuf[off+52:]))  // process_pid
				tgid := int(binary.LittleEndian.Uint32(rbuf[off+56:])) // process_tgid
				t.Logf("CN_PROC PROC_EVENT_EXEC: process_pid=%d process_tgid=%d (%d bytes)", pid, tgid, evLen)
				if tgid == 0 && pid == 0 {
					t.Fatalf("pid=0 tgid=0 — unexpected, layout mismatch")
				}
				found = true
			}
		}
		off += (evLen + 7) &^ 7 // NLMSG_ALIGN
	}
	if !found {
		t.Fatalf("no PROC_EVENT_EXEC in %d bytes received — layout mismatch or event filtered", got)
	}
	t.Log("PROBE OK — netlink process connector delivers exec events (modern proc_event union)")
}
