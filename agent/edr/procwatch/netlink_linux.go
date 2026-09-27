//go:build linux

// Linux process watcher over the netlink process connector (CN_PROC).
//
// The kernel's cn_proc module multicasts a proc_event for every process
// exec. We subscribe for PROC_EVENT_EXEC only and translate each event
// into a ProcStart, enriching it from /proc (ppid, image, cmdline).
//
// Wire format (include/uapi/linux/cn_proc.h, kernel 6.8+; validated live
// by cnproc_probe_linux_test.go on 7.0.0-34-generic, 2026-09-27):
//
//	nlmsghdr   {len, type, flags, seq, pid}              16 bytes
//	cn_msg     {id.idx, id.major, seq, ack, len}         20 bytes
//	mcast_op   {mcast_op}                                 4 bytes  (subscribe, len=4 form)
//	proc_event {what, cpu, timestamp_ns, union event_data} 40 bytes
//
// A PROC_EVENT_EXEC event is 76 bytes minimum; in the received buffer
// what is at 36, exec.process_pid at 52, exec.process_tgid at 56. The
// modern union carries NO ppid — that comes from /proc.
//
// Subscribe uses the ancient len=4 form (mcast_op only, no event_type):
// it is accepted by every kernel (the 6.5+ len=8 proc_input form is
// silently dropped by older kernels). The kernel then delivers ALL event
// types (fork/exec/exit/...), so we filter what==PROC_EVENT_EXEC in
// userspace (execTgid).
package procwatch

import (
	"context"
	"encoding/binary"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/unix"
)

const (
	cnIdxProc         = 1  // CN_IDX_PROC
	cnValProc         = 1  // CN_VAL_PROC
	procCnMcastListen = 1  // PROC_CN_MCAST_LISTEN
	procEventExec     = 2  // PROC_EVENT_EXEC
	procEventMinLen   = 76 // nlmsghdr + cn_msg + proc_event (exec)
	procReadBufSize   = 64 * 1024
)

// Watcher is the Linux CN_PROC process watcher. It mirrors the Windows
// WMI lifecycle: Run retries subscribe forever until ctx is cancelled.
type Watcher struct{ h Handler }

func New(h Handler) *Watcher { return &Watcher{h: h} }

func (w *Watcher) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err := w.subscribe(ctx); err != nil {
			logger.Error("UTMStack EDR process watcher: %v; retrying", err)
			time.Sleep(3 * time.Second)
		}
	}
}

// subscribe opens the netlink connector socket, joins the CN_IDX_PROC
// multicast group, subscribes for PROC_EVENT_EXEC, and reads events
// until ctx is cancelled or the socket errors.
func (w *Watcher) subscribe(ctx context.Context) error {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_CONNECTOR)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	// Groups MUST be CN_IDX_PROC (1): the kernel delivers events via
	// netlink_broadcast(group=CN_IDX_PROC), which only reaches sockets
	// that joined that multicast group at bind time. Groups=0 → zero
	// events, silently (the subscription itself still succeeds).
	sa := &unix.SockaddrNetlink{Family: unix.AF_NETLINK, Groups: cnIdxProc}
	if err := unix.Bind(fd, sa); err != nil {
		return err
	}

	// Subscribe: nlmsghdr + cn_msg(len=4) + mcast_op = 40 bytes.
	// The len=4 form (mcast_op only) is the ancient one, accepted by every
	// kernel; the len=8 proc_input form (mcast_op + event_type) was added
	// in 6.5 and is silently dropped by older kernels. With len=4 the
	// kernel delivers ALL event types, so we filter exec in userspace.
	// No ack is expected after subscribe: the kernel's ack carries
	// what=PROC_EVENT_NONE, which the per-socket filter drops.
	buf := make([]byte, 40)
	binary.LittleEndian.PutUint32(buf[0:], 40)          // nlmsg_len
	binary.LittleEndian.PutUint16(buf[4:], 1)           // nlmsg_type (arbitrary for connector)
	binary.LittleEndian.PutUint32(buf[8:], 1)           // nlmsg_seq
	binary.LittleEndian.PutUint32(buf[16:], cnIdxProc)  // cn_msg.id.idx
	binary.LittleEndian.PutUint32(buf[20:], cnValProc)  // cn_msg.id.major (routing is by the full {idx,val} pair)
	binary.LittleEndian.PutUint32(buf[24:], 1)          // cn_msg.seq
	binary.LittleEndian.PutUint32(buf[28:], 0)          // cn_msg.ack
	binary.LittleEndian.PutUint32(buf[32:], 4)          // cn_msg.len = sizeof(mcast_op)
	binary.LittleEndian.PutUint32(buf[36:], procCnMcastListen) // mcast_op = PROC_CN_MCAST_LISTEN
	if err := unix.Sendto(fd, buf, 0, sa); err != nil {
		return err
	}

	// Read loop: poll + non-blocking read so ctx cancellation is observed
	// promptly (a plain blocking read would hang the goroutine).
	if err := unix.SetNonblock(fd, true); err != nil {
		return err
	}
	var pfd [1]unix.PollFd
	pfd[0].Fd = int32(fd)
	pfd[0].Events = unix.POLLIN
	rbuf := make([]byte, procReadBufSize)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		n, err := unix.Poll(pfd[:], 500)
		if err != nil {
			return err
		}
		if n == 0 {
			continue
		}
		got, _, rerr := unix.Recvfrom(fd, rbuf, 0)
		if rerr != nil {
			if rerr == unix.EAGAIN || rerr == unix.EWOULDBLOCK {
				continue
			}
			return rerr
		}
		walkExecEvents(rbuf[:got], w.deliver)
	}
}

// walkExecEvents walks the nlmsg chain in buf and invokes onExec for every
// PROC_EVENT_EXEC it finds.
func walkExecEvents(buf []byte, onExec func(tgid int)) {
	for off := 0; off+procEventMinLen <= len(buf); {
		evLen := int(binary.LittleEndian.Uint32(buf[off:])) // nlmsg_len
		if evLen < procEventMinLen {
			break
		}
		tgid, isExec := execTgid(buf, off)
		if isExec {
			onExec(tgid)
		}
		off += (evLen + 7) &^ 7 // NLMSG_ALIGN
	}
}

// execTgid extracts the user-visible PID (process_tgid) from one nlmsg at
// off. It returns (0, false) when the buffer is too short for an exec
// event or the event is not a PROC_EVENT_EXEC.
func execTgid(buf []byte, off int) (tgid int, isExec bool) {
	if off+procEventMinLen > len(buf) {
		return 0, false
	}
	if binary.LittleEndian.Uint32(buf[off+36:]) != procEventExec { // what
		return 0, false
	}
	return int(binary.LittleEndian.Uint32(buf[off+56:])), true // exec.process_tgid
}

// deliver enriches an exec event's tgid from /proc and hands it to the
// handler. The process may already have exited (short-lived execs are
// common), so every /proc read is best-effort: we deliver whenever pid != 0,
// even with an empty image (consumers skip the scan on empty Image but
// still record pid/ppid).
func (w *Watcher) deliver(pid int) {
	if pid == 0 {
		return
	}
	w.h.OnProcStart(ProcStart{
		PID:     pid,
		PPID:    procPPID(pid),
		Image:   procImage(pid),
		Cmdline: procCmdline(pid),
	})
}

// procPPID reads field 4 (ppid) of /proc/<pid>/stat. The comm field (2) is
// parenthesized and may contain spaces/parens, so we split after the LAST
// ')'. Returns 0 on any error (the process may be gone).
func procPPID(pid int) int {
	line, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return 0
	}
	s := string(line)
	i := strings.LastIndex(s, ")")
	if i < 0 || i+1 >= len(s) {
		return 0
	}
	rest := strings.Fields(s[i+1:])
	if len(rest) < 2 {
		return 0
	}
	ppid, _ := strconv.Atoi(rest[1]) // rest[0] is state (field 3)
	return ppid
}

// procImage reads the /proc/<pid>/exe symlink. Returns "" on any error.
func procImage(pid int) string {
	img, err := os.Readlink("/proc/" + strconv.Itoa(pid) + "/exe")
	if err != nil {
		return ""
	}
	return img
}

// procCmdline reads /proc/<pid>/cmdline (NUL-separated args) and joins them
// with spaces. Returns "" on any error.
func procCmdline(pid int) string {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/cmdline")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(strings.TrimRight(string(b), "\x00"), "\x00", " ")
}
