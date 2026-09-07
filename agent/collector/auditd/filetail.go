//go:build linux
// +build linux

package auditd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/elastic/go-libaudit/v2/auparse"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

const (
	auditLogPath        = "/var/log/audit/audit.log"
	tailPollInterval    = 200 * time.Millisecond
	cursorFlushInterval = 2 * time.Second

	readBufSize = 64 * 1024
)

var auditLogCursorFile = filepath.Join(fs.GetExecutablePath(), "audit_log_cursor.json")

type filePosition struct {
	Inode  uint64 `json:"inode"`
	Offset int64  `json:"offset"`
}

type pendingOffset struct {
	generation int
	offset     int64
}

type auditLogCursor struct {
	mu sync.Mutex

	generation int
	inode      uint64

	safe        filePosition
	writtenSafe filePosition
	pending     map[uint32]pendingOffset
}

func newAuditLogCursor() *auditLogCursor {
	c := &auditLogCursor{pending: make(map[uint32]pendingOffset)}
	if data, err := os.ReadFile(auditLogCursorFile); err == nil {
		var pos filePosition
		if err := json.Unmarshal(data, &pos); err == nil {
			c.safe = pos
			c.writtenSafe = pos
		}
	}
	return c
}

func (c *auditLogCursor) resumePosition() filePosition {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.safe
}

func (c *auditLogCursor) reset(inode uint64, offset int64) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.generation++
	c.inode = inode
	c.safe = filePosition{Inode: inode, Offset: offset}
	c.pending = make(map[uint32]pendingOffset)
	return c.generation
}

func (c *auditLogCursor) track(generation int, seq uint32, fileOffset int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pending[seq] = pendingOffset{generation: generation, offset: fileOffset}
}

func (c *auditLogCursor) resolve(seq uint32, persisted bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	p, ok := c.pending[seq]
	delete(c.pending, seq)
	if !ok || !persisted || p.generation != c.generation {
		return
	}
	if p.offset > c.safe.Offset {
		c.safe = filePosition{Inode: c.inode, Offset: p.offset}
	}
}

func (c *auditLogCursor) flush() {
	c.mu.Lock()
	safe := c.safe
	unchanged := safe == c.writtenSafe
	c.mu.Unlock()

	if unchanged {
		return
	}

	data, err := json.Marshal(safe)
	if err != nil {
		utils.Logger.ErrorF("auditd: error encoding audit.log cursor: %v", err)
		return
	}
	if err := os.WriteFile(auditLogCursorFile, data, 0o600); err != nil {
		utils.Logger.ErrorF("auditd: error persisting audit.log cursor: %v", err)
		return
	}

	c.mu.Lock()
	c.writtenSafe = safe
	c.mu.Unlock()
}

func (c *auditLogCursor) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(cursorFlushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			c.flush()
			return
		case <-ticker.C:
			c.flush()
		}
	}
}

func fileInode(fi os.FileInfo) (uint64, error) {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("cannot read inode for %s", fi.Name())
	}
	return st.Ino, nil
}

type pushMessage interface {
	PushMessage(msg *auparse.AuditMessage)
}

type auditLogTailer struct {
	cursor      *auditLogCursor
	reassembler pushMessage

	file       *os.File
	inode      uint64
	generation int
	readOffset int64
}

func newAuditLogTailer(cursor *auditLogCursor, reassembler pushMessage) *auditLogTailer {
	return &auditLogTailer{cursor: cursor, reassembler: reassembler}
}

func (t *auditLogTailer) open() error {
	fi, err := os.Stat(auditLogPath)
	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", auditLogPath, err)
	}
	inode, err := fileInode(fi)
	if err != nil {
		return err
	}

	resume := t.cursor.resumePosition()
	var startOffset int64
	switch {
	case resume.Inode == 0:
		startOffset = fi.Size()
	case resume.Inode == inode:
		startOffset = resume.Offset
	default:
		utils.Logger.Info("auditd: saved audit.log position belongs to a rotated-out file, reading current file from the beginning")
		startOffset = 0
	}

	f, err := os.Open(auditLogPath)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", auditLogPath, err)
	}
	if _, err := f.Seek(startOffset, io.SeekStart); err != nil {
		f.Close()
		return fmt.Errorf("cannot seek %s: %w", auditLogPath, err)
	}

	t.file = f
	t.inode = inode
	t.readOffset = startOffset
	t.generation = t.cursor.reset(inode, startOffset)
	return nil
}

func (t *auditLogTailer) close() {
	if t.file != nil {
		t.file.Close()
		t.file = nil
	}
}

func (t *auditLogTailer) run(ctx context.Context) error {
	buf := make([]byte, readBufSize)
	var carry []byte

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		n, err := t.file.Read(buf)
		if n > 0 {
			carry = t.processChunk(append(carry, buf[:n]...))
		}
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading %s: %w", auditLogPath, err)
		}

		if n > 0 {
			continue
		}

		if rotated, newInode := t.checkRotation(); rotated {
			t.close()
			f, err := os.Open(auditLogPath)
			if err != nil {
				return fmt.Errorf("cannot open rotated %s: %w", auditLogPath, err)
			}
			t.file = f
			t.inode = newInode
			t.readOffset = 0
			t.generation = t.cursor.reset(newInode, 0)
			carry = nil
			continue
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(tailPollInterval):
		}
	}
}

func (t *auditLogTailer) checkRotation() (rotated bool, newInode uint64) {
	fi, err := os.Stat(auditLogPath)
	if err != nil {
		return false, 0
	}
	inode, err := fileInode(fi)
	if err != nil || inode == t.inode {
		return false, 0
	}
	return true, inode
}

func (t *auditLogTailer) processChunk(data []byte) []byte {
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] != '\n' {
			continue
		}
		line := string(data[start:i])
		lineEndOffset := t.readOffset + int64(i) + 1
		start = i + 1
		t.processLine(line, lineEndOffset)
	}

	t.readOffset += int64(start)
	if start == 0 {
		return data
	}
	remainder := make([]byte, len(data)-start)
	copy(remainder, data[start:])
	return remainder
}

func (t *auditLogTailer) processLine(line string, lineEndOffset int64) {
	if line == "" {
		return
	}
	msg, err := auparse.ParseLogLine(line)
	if err != nil {
		return
	}
	t.cursor.track(t.generation, msg.Sequence, lineEndOffset)
	t.reassembler.PushMessage(msg)
}
