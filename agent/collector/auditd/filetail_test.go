//go:build linux
// +build linux

package auditd

import (
	"testing"

	"github.com/elastic/go-libaudit/v2/auparse"
)

type recordingReassembler struct {
	pushed []*auparse.AuditMessage
}

func (r *recordingReassembler) PushMessage(msg *auparse.AuditMessage) {
	r.pushed = append(r.pushed, msg)
}

func TestAuditLogCursor_ResolveAdvancesOnSuccess(t *testing.T) {
	c := &auditLogCursor{pending: map[uint32]pendingOffset{}}
	gen := c.reset(42, 0)

	c.track(gen, 7, 100)
	c.resolve(7, true)

	got := c.resumePosition()
	want := filePosition{Inode: 42, Offset: 100}
	if got != want {
		t.Fatalf("resumePosition() = %+v, want %+v", got, want)
	}
}

func TestAuditLogCursor_ResolveIgnoresFailure(t *testing.T) {
	c := &auditLogCursor{pending: map[uint32]pendingOffset{}}
	gen := c.reset(42, 10)

	c.track(gen, 7, 500)
	c.resolve(7, false) // enqueue failed: must not advance

	got := c.resumePosition()
	want := filePosition{Inode: 42, Offset: 10}
	if got != want {
		t.Fatalf("resumePosition() = %+v, want unchanged %+v", got, want)
	}

	// The pending entry must still be cleaned up so it can't leak forever.
	if _, ok := c.pending[7]; ok {
		t.Fatalf("pending entry for seq 7 was not cleared after resolve")
	}
}

func TestAuditLogCursor_ResolveIgnoresStaleGeneration(t *testing.T) {
	c := &auditLogCursor{pending: map[uint32]pendingOffset{}}
	gen := c.reset(1, 0)

	// Line read while still on generation `gen` (the old file), but a
	// rotation happens (bumping the generation) before it resolves.
	c.track(gen, 3, 900)
	c.reset(2, 0) // simulates detecting rotation to a new inode

	c.resolve(3, true)

	got := c.resumePosition()
	want := filePosition{Inode: 2, Offset: 0}
	if got != want {
		t.Fatalf("resolve() applied a stale generation's offset: got %+v, want %+v", got, want)
	}
}

func TestAuditLogCursor_ResetClearsPending(t *testing.T) {
	c := &auditLogCursor{pending: map[uint32]pendingOffset{}}
	gen := c.reset(1, 0)
	c.track(gen, 1, 50)
	c.track(gen, 2, 60)

	c.reset(2, 0)

	if len(c.pending) != 0 {
		t.Fatalf("reset() left %d stale pending entries", len(c.pending))
	}
}

func TestProcessChunk_TracksOffsetsAndPushesInOrder(t *testing.T) {
	reassembler := &recordingReassembler{}
	cursor := &auditLogCursor{pending: map[uint32]pendingOffset{}}
	tailer := newAuditLogTailer(cursor, reassembler)
	tailer.generation = cursor.reset(1, 0)

	line1 := `type=SYSCALL msg=audit(1700000000.100:1): arch=c000003e syscall=59 success=yes exit=0 a0=1 items=0 ppid=1 pid=2 auid=1000 uid=1000 gid=1000 euid=1000 suid=1000 fsuid=1000 egid=1000 sgid=1000 fsgid=1000 tty=pts0 ses=1 comm="bash" exe="/bin/bash" key="utmstack_exec"` + "\n"
	line2 := `type=SYSCALL msg=audit(1700000000.200:2): arch=c000003e syscall=59 success=yes exit=0 a0=1 items=0 ppid=1 pid=3 auid=1000 uid=1000 gid=1000 euid=1000 suid=1000 fsuid=1000 egid=1000 sgid=1000 fsgid=1000 tty=pts0 ses=1 comm="bash" exe="/bin/bash" key="utmstack_exec"` + "\n"

	remainder := tailer.processChunk([]byte(line1 + line2))

	if len(remainder) != 0 {
		t.Fatalf("expected no remainder for two complete lines, got %d bytes", len(remainder))
	}
	if len(reassembler.pushed) != 2 {
		t.Fatalf("expected 2 messages pushed, got %d", len(reassembler.pushed))
	}
	if reassembler.pushed[0].Sequence != 1 || reassembler.pushed[1].Sequence != 2 {
		t.Fatalf("pushed out of order: seqs %d, %d", reassembler.pushed[0].Sequence, reassembler.pushed[1].Sequence)
	}
	if tailer.readOffset != int64(len(line1)+len(line2)) {
		t.Fatalf("readOffset = %d, want %d", tailer.readOffset, len(line1)+len(line2))
	}

	// Both lines' end offsets must have been tracked against the current
	// generation, ready to be resolved once their events are persisted.
	if p, ok := cursor.pending[1]; !ok || p.offset != int64(len(line1)) {
		t.Fatalf("sequence 1 not tracked at expected offset: %+v, ok=%v", p, ok)
	}
	if p, ok := cursor.pending[2]; !ok || p.offset != int64(len(line1)+len(line2)) {
		t.Fatalf("sequence 2 not tracked at expected offset: %+v, ok=%v", p, ok)
	}
}

func TestProcessChunk_CarriesPartialLineAcrossCalls(t *testing.T) {
	reassembler := &recordingReassembler{}
	cursor := &auditLogCursor{pending: map[uint32]pendingOffset{}}
	tailer := newAuditLogTailer(cursor, reassembler)
	tailer.generation = cursor.reset(1, 0)

	full := `type=SYSCALL msg=audit(1700000000.100:9): arch=c000003e syscall=59 success=yes exit=0 a0=1 items=0 ppid=1 pid=2 auid=1000 uid=1000 gid=1000 euid=1000 suid=1000 fsuid=1000 egid=1000 sgid=1000 fsgid=1000 tty=pts0 ses=1 comm="bash" exe="/bin/bash" key="utmstack_exec"` + "\n"
	split := len(full) / 2

	remainder := tailer.processChunk([]byte(full[:split]))
	if len(reassembler.pushed) != 0 {
		t.Fatalf("expected no message pushed before the line is complete, got %d", len(reassembler.pushed))
	}
	if string(remainder) != full[:split] {
		t.Fatalf("remainder = %q, want the partial line held back", remainder)
	}
	if tailer.readOffset != 0 {
		t.Fatalf("readOffset advanced on a partial line: %d", tailer.readOffset)
	}

	remainder = tailer.processChunk(append(remainder, []byte(full[split:])...))
	if len(remainder) != 0 {
		t.Fatalf("expected no remainder once the line completed, got %d bytes", len(remainder))
	}
	if len(reassembler.pushed) != 1 {
		t.Fatalf("expected 1 message pushed once the line completed, got %d", len(reassembler.pushed))
	}
	if tailer.readOffset != int64(len(full)) {
		t.Fatalf("readOffset = %d, want %d", tailer.readOffset, len(full))
	}
}

func TestProcessLine_SkipsUnparseableLines(t *testing.T) {
	reassembler := &recordingReassembler{}
	cursor := &auditLogCursor{pending: map[uint32]pendingOffset{}}
	tailer := newAuditLogTailer(cursor, reassembler)
	tailer.generation = cursor.reset(1, 0)

	tailer.processLine("", 0)
	tailer.processLine("not an audit line", 10)

	if len(reassembler.pushed) != 0 {
		t.Fatalf("expected unparseable lines to be skipped, pushed %d messages", len(reassembler.pushed))
	}
	if len(cursor.pending) != 0 {
		t.Fatalf("expected no pending entries for unparseable lines, got %d", len(cursor.pending))
	}
}
