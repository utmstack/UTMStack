//go:build linux

// Linux process watcher over the netlink process connector (CN_PROC).
//
// The kernel's cn_proc module multicasts a proc_event for every process
// exec. We subscribe for PROC_EVENT_EXEC only and translate each event
// into a ProcStart, enriching it from /proc (ppid, image, cmdline).
//
// Wire format (include/uapi/linux/cn_proc.h, kernel 7.0.0-34-generic;
// validated live by cnprobe_hdr on 2026-10-05, sizeof(cn_msg)=20, sizeof
// (proc_event)=40):
//
//	nlmsghdr   {len, type, flags, seq, pid}              16 bytes  @  0
//	cn_msg     {id.idx, id.val, seq, ack, len, flags}    20 bytes  @ 16
//		id.idx        u16 @ 16
//		id.val        u16 @ 18
//		seq           u32 @ 20
//		ack           u32 @ 24
//		len           u16 @ 28  (proc_event size = 40)
//		flags         u16 @ 30
//	proc_event {what, cpu, ts_ns, union}                 40 bytes  @ 36
//		what                u32 @ 36
//		cpu                 u32 @ 40
//		timestamp_ns        u64 @ 44
//		exec.process_pid    u32 @ 52
//		exec.process_tgid   u32 @ 56
//
// A PROC_EVENT_EXEC nlmsg is 76 bytes (16+20+40); the modern union carries
// NO ppid — that comes from /proc.
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
	procEventMinLen   = 76 // nlmsghdr (16) + cn_msg (20) + proc_event (40)
	whatOffset        = 36 // proc_event.what in the received buffer
	execTgidOffset    = 56 // proc_event.event_data.exec.process_tgid
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

	// Subscribe: nlmsghdr (16) + cn_msg (20) + mcast_op (4) = 40 bytes.
	// cn_msg fields are u16/u16/u32/u32/u16/u16 (16 named bytes + 4 pad).
	// The len=4 form (mcast_op only) is the ancient one, accepted by every
	// kernel. The kernel delivers ALL event types, so we filter exec in
	// userspace. No ack is expected after subscribe.
	//
	// Wire offsets (validated by cnprobe_hdr on 7.0.0-34-generic):
	//	16  id.idx  u16
	//	18  id.val  u16
	//	20  seq     u32
	//	24  ack     u32
	//	28  len     u16  (= 4, sizeof mcast_op)
	//	30  flags   u16
	//	32  padding u32
	//	36  mcast_op u32
	buf := buildSubscribeMsg()
	if err := unix.Sendto(fd, buf, 0, sa); err != nil {
		return err
	}
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

// buildSubscribeMsg returns the 40-byte CN_PROC subscription message
// (nlmsghdr + cn_msg + mcast_op), byte-identical to a kernel struct
// packing. The kernel routes by the {id.idx, id.val} pair and requires
// cn_msg.len >= 4; both are u16, so writing them as u32 (the Y2.5 bug)
// produces {1,0} with len=0 and the subscription silently delivers
// nothing.
func buildSubscribeMsg() []byte {
	buf := make([]byte, 40)
	binary.LittleEndian.PutUint32(buf[0:], 40)          // nlmsg_len
	binary.LittleEndian.PutUint16(buf[4:], 1)           // nlmsg_type
	binary.LittleEndian.PutUint32(buf[8:], 1)           // nlmsg_seq
	binary.LittleEndian.PutUint16(buf[16:], cnIdxProc)  // cn_msg.id.idx (u16!)
	binary.LittleEndian.PutUint16(buf[18:], cnValProc)  // cn_msg.id.val (u16!)
	binary.LittleEndian.PutUint32(buf[20:], 1)          // cn_msg.seq
	binary.LittleEndian.PutUint32(buf[24:], 0)          // cn_msg.ack
	binary.LittleEndian.PutUint16(buf[28:], 4)          // cn_msg.len = sizeof(mcast_op)
	binary.LittleEndian.PutUint16(buf[30:], 0)          // cn_msg.flags
	// buf[32:36] = padding (zeroed by make)
	binary.LittleEndian.PutUint32(buf[36:], procCnMcastListen) // mcast_op
	return buf
}

// walkExecEvents walks the nlmsg chain in buf and invokes onExec for every
// PROC_EVENT_EXEC it finds. Forks (68 bytes), exits, and other event types
// share the same socket and interleave with execs; each message is exactly
// one 76-byte exec or a shorter non-exec, so a short nlmsg_len is expected
// and we advance past it (NLMSG_ALIGN) rather than breaking — breaking
// would skip every exec that follows in the same recv.
func walkExecEvents(buf []byte, onExec func(tgid int)) {
	for off := 0; off+procEventMinLen <= len(buf); {
		evLen := int(binary.LittleEndian.Uint32(buf[off:])) // nlmsg_len
		if evLen < procEventMinLen || off+evLen > len(buf) {
			break // malformed / truncated — don't walk garbage
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
	if binary.LittleEndian.Uint32(buf[off+whatOffset:]) != procEventExec { // what
		return 0, false
	}
	return int(binary.LittleEndian.Uint32(buf[off+execTgidOffset:])), true // exec.process_tgid
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
