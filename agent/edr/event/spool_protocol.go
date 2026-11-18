package event

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const maxLegacySpoolBytes = 64 << 20
const maxDrainRecords = 256
const maxDrainBytes = 1 << 20
const spoolHeaderPrefix = `{"_openedr_spool":`

type spoolHeader struct {
	Version    int    `json:"_openedr_spool"`
	Generation string `json:"generation"`
}
type spoolCursor struct {
	Version    int    `json:"version"`
	Generation string `json:"generation"`
	Offset     int64  `json:"offset"`
}

func validGeneration(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 16 && hex.EncodeToString(b) == s
}
func newSpoolHeader() ([]byte, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(spoolHeader{2, hex.EncodeToString(b[:])})
	return append(raw, '\n'), err
}
func decodeSpoolJSON(raw []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return ErrSpoolIntegrity
	}
	return nil
}

// ensureSegment atomically upgrades legacy bytes without trusting an old bare
// byte offset. Legacy migration may replay already relayed records; it cannot
// infer which rotated file an unbound decimal cursor described. No old data is
// truncated to fit the new admission cap. Oversized/corrupt inputs are retained.
func ensureSegment(path string, create, upgrade bool) (generation string, headerBytes int64, upgraded bool, err error) {
	f, err := openSpoolRegular(path, os.O_RDONLY, 0600)
	if os.IsNotExist(err) && create {
		var raw []byte
		var e error
		if upgrade {
			raw, e = newSpoolHeader()
		}
		if e != nil {
			return "", 0, false, e
		}
		tmp := path + ".migrating"
		out, e := openSpoolRegular(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if e != nil {
			return "", 0, false, e
		}
		defer os.Remove(tmp)
		_, e = out.Write(raw)
		e = errors.Join(e, out.Sync(), out.Close())
		if e != nil {
			return "", 0, false, e
		}
		if e = replaceSpoolFile(tmp, path); e != nil {
			return "", 0, false, e
		}
		if e = syncSpoolDirectory(filepath.Dir(path)); e != nil {
			return "", 0, false, e
		}
		var h spoolHeader
		_ = json.Unmarshal(bytes.TrimSpace(raw), &h)
		return h.Generation, int64(len(raw)), false, nil
	}
	if err != nil {
		return "", 0, false, err
	}
	var first [128]byte
	n, readErr := f.ReadAt(first[:], 0)
	if readErr != nil && readErr != io.EOF {
		f.Close()
		return "", 0, false, readErr
	}
	if bytes.HasPrefix(first[:n], []byte(spoolHeaderPrefix)) {
		end := bytes.IndexByte(first[:n], '\n')
		var h spoolHeader
		if end < 0 || decodeSpoolJSON(first[:end], &h) != nil || h.Version != 2 || !validGeneration(h.Generation) {
			f.Close()
			return "", 0, false, fmt.Errorf("%w: invalid segment header", ErrSpoolIntegrity)
		}
		if err = f.Close(); err != nil {
			return "", 0, false, err
		}
		return h.Generation, int64(end + 1), false, nil
	}
	if !upgrade {
		f.Close()
		return "", 0, false, nil
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return "", 0, false, err
	}
	if info.Size() > maxLegacySpoolBytes {
		f.Close()
		return "", 0, false, fmt.Errorf("%w: oversized legacy segment", ErrSpoolIntegrity)
	}
	raw, err := newSpoolHeader()
	if err != nil {
		f.Close()
		return "", 0, false, err
	}
	// A fixed reserved staging name bounds restart debris. It is never adopted
	// as authority; each retry rewrites it from the still-authoritative source.
	tmp := path + ".migrating"
	out, err := openSpoolRegular(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		f.Close()
		return "", 0, false, err
	}
	defer os.Remove(tmp)
	_, err = out.Write(raw)
	if err == nil {
		var copied int64
		copied, err = io.Copy(out, io.LimitReader(f, maxLegacySpoolBytes+1))
		if copied > maxLegacySpoolBytes {
			err = ErrSpoolIntegrity
		}
	}
	err = errors.Join(err, out.Sync(), out.Close(), f.Close())
	if err != nil {
		return "", 0, false, err
	}
	if err = replaceSpoolFile(tmp, path); err != nil {
		return "", 0, false, err
	}
	if err = syncSpoolDirectory(filepath.Dir(path)); err != nil {
		return "", 0, false, err
	}
	var h spoolHeader
	_ = json.Unmarshal(bytes.TrimSpace(raw), &h)
	return h.Generation, int64(len(raw)), true, nil
}

func lockSpool(ctx context.Context, path string) (func(), error) {
	f, err := openSpoolRegular(path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	for {
		locked, e := trySpoolLock(f)
		if e != nil {
			f.Close()
			return nil, e
		}
		if locked {
			return func() { unlockSpoolFile(f); f.Close() }, nil
		}
		timer := time.NewTimer(5 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			f.Close()
			return nil, errors.Join(ErrSpoolBusy, ctx.Err())
		case <-timer.C:
		}
	}
}
func readSpoolCursor(path string) (spoolCursor, error) {
	c := spoolCursor{Version: 2}
	f, err := openSpoolRegular(path, os.O_RDONLY, 0600)
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil {
		return c, err
	}
	if len(raw) > 4096 {
		return c, ErrSpoolIntegrity
	}
	trimmed := strings.TrimSpace(string(raw))
	if n, e := strconv.ParseInt(trimmed, 10, 64); e == nil && n >= 0 {
		return c, nil
	} // legacy unbound cursor: replay conservatively once
	if decodeSpoolJSON(raw, &c) != nil || c.Version != 2 || !validGeneration(c.Generation) || c.Offset < 0 {
		return c, fmt.Errorf("%w: invalid relay cursor", ErrSpoolIntegrity)
	}
	return c, nil
}
func writeSpoolCursor(path string, c spoolCursor) error {
	raw, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return writeSpoolState(path, append(raw, '\n'))
}
func writeSpoolState(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := openSpoolRegular(path+".next", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer os.Remove(path + ".next")
	_, err = f.Write(raw)
	err = errors.Join(err, f.Sync(), f.Close())
	if err != nil {
		return err
	}
	if err = replaceSpoolFile(path+".next", path); err != nil {
		return err
	}
	return syncSpoolDirectory(filepath.Dir(path))
}
func boundedSpoolLine(r *bufio.Reader) ([]byte, error) {
	var line []byte
	for {
		part, err := r.ReadSlice('\n')
		if len(line)+len(part) > MaxSpoolRecordBytes {
			return nil, ErrSpoolRecord
		}
		line = append(line, part...)
		if err != bufio.ErrBufferFull {
			return line, err
		}
	}
}

// DrainSpool transfers a bounded batch to emit, committing only accepted lines.
// emit must return promptly (the production LogQueue send is nonblocking).
// Local cursor errors or a process crash can replay an accepted batch; this is
// at-least-once admission to the agent queue, not remote delivery acknowledgement.
// Cancellation and rejected records retain all remaining bytes for a later run.
func DrainSpool(ctx context.Context, path, offset string, emit func(string) bool) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	lockCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	unlock, err := lockSpool(lockCtx, path)
	if err != nil {
		return 0, err
	}
	defer unlock()
	protocol, err := readSpoolState(path+".protocol", 64)
	if os.IsNotExist(err) {
		return drainLegacySpool(ctx, path, offset, emit)
	}
	if err != nil {
		return 0, err
	}
	if string(protocol) != "openedr-event-spool-v2\n" {
		return 0, ErrSpoolIntegrity
	}
	cursor, err := readSpoolCursor(offset)
	if err != nil {
		return 0, err
	}
	records, total := 0, 0
	for _, segment := range []string{path + ".1", path} {
		gen, header, _, err := ensureSegment(segment, false, true)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return cursor.Offset, err
		}
		f, err := openSpoolRegular(segment, os.O_RDONLY, 0600)
		if err != nil {
			return cursor.Offset, err
		}
		fi, err := f.Stat()
		if err != nil {
			f.Close()
			return cursor.Offset, err
		}
		pos := int64(0)
		if cursor.Generation == gen {
			pos = cursor.Offset
		}
		if pos > fi.Size()-header {
			f.Close()
			return cursor.Offset, fmt.Errorf("%w: cursor exceeds segment", ErrSpoolIntegrity)
		}
		if pos > 0 {
			var last [1]byte
			if _, e := f.ReadAt(last[:], header+pos-1); e != nil || last[0] != '\n' {
				f.Close()
				return pos, fmt.Errorf("%w: cursor not at a record boundary", ErrSpoolIntegrity)
			}
		}
		if _, err = f.Seek(header+pos, 0); err != nil {
			f.Close()
			return pos, err
		}
		reader := bufio.NewReader(f)
		var stopErr error
		stopped := false
		for records < maxDrainRecords && total < maxDrainBytes {
			if err = ctx.Err(); err != nil {
				stopErr = err
				stopped = true
				break
			}
			line, e := boundedSpoolLine(reader)
			if e != nil {
				if e == io.EOF && len(line) > 0 {
					e = fmt.Errorf("%w: partial record tail", ErrSpoolIntegrity)
				}
				if e != io.EOF {
					stopErr = e
				}
				break
			}
			if len(line) == 0 || line[len(line)-1] != '\n' {
				stopErr = ErrSpoolIntegrity
				break
			}
			value := strings.TrimSuffix(string(line), "\n")
			if value != "" && !emit(value) {
				stopped = true
				break
			}
			pos += int64(len(line))
			records++
			total += len(line)
		}
		closeErr := f.Close()
		next := spoolCursor{Version: 2, Generation: gen, Offset: pos}
		if next != cursor {
			if err = writeSpoolCursor(offset, next); err != nil {
				return pos, err
			}
			cursor = next
		}
		if closeErr != nil {
			return pos, closeErr
		}
		if stopErr != nil {
			return pos, stopErr
		}
		if stopped || pos < fi.Size()-header {
			return pos, nil
		}
		if segment == path+".1" {
			// The sealed segment is complete and its accepted offset is durable.
			// No producer can rename/append while this shared lock is held.
			if err = os.Remove(segment); err != nil {
				return pos, err
			}
			if err = syncSpoolDirectory(filepath.Dir(path)); err != nil {
				return pos, err
			}
		}
		if records >= maxDrainRecords || total >= maxDrainBytes {
			return pos, nil
		}
	}
	return cursor.Offset, nil
}

// An older producer holds a persistent active descriptor and does not take our
// lock. Until a new producer declares the protocol, keep the old non-mutating
// reader behavior: never rename its files underneath that descriptor. Legacy
// rotation still has its old coverage limitations; this path is compatibility,
// not the v2 accepted-record retention guarantee.
func drainLegacySpool(ctx context.Context, path, offset string, emit func(string) bool) (int64, error) {
	var start int64
	raw, e := readSpoolState(offset, 4096)
	if e == nil {
		var err error
		start, err = strconv.ParseInt(strings.TrimSpace(string(raw)), 10, 64)
		if err != nil || start < 0 {
			return 0, ErrSpoolIntegrity
		}
	} else if !os.IsNotExist(e) {
		return 0, e
	}
	f, err := openSpoolRegular(path, os.O_RDONLY, 0600)
	if os.IsNotExist(err) {
		return start, nil
	}
	if err != nil {
		return start, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return start, err
	}
	if fi.Size() < start {
		start = 0
	}
	if _, err = f.Seek(start, 0); err != nil {
		return start, err
	}
	reader := bufio.NewReader(f)
	pos := start
	var readErr error
	for n, total := 0, 0; n < maxDrainRecords && total < maxDrainBytes; n++ {
		if readErr = ctx.Err(); readErr != nil {
			break
		}
		line, e := boundedSpoolLine(reader)
		if e != nil {
			if e != io.EOF {
				readErr = e
			}
			break
		}
		value := strings.TrimSuffix(string(line), "\n")
		if value != "" && !emit(value) {
			break
		}
		pos += int64(len(line))
		total += len(line)
	}
	if pos != start || e != nil && os.IsNotExist(e) {
		if err = writeSpoolState(offset, []byte(strconv.FormatInt(pos, 10))); err != nil {
			return pos, err
		}
	}
	return pos, readErr
}

func readSpoolState(path string, limit int64) ([]byte, error) {
	f, err := openSpoolRegular(path, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit {
		return nil, ErrSpoolIntegrity
	}
	return raw, nil
}
