package engine

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
	"strings"
	"unicode/utf16"
)

// The reference's macro project dump and module text, learned from its
// outputs (clamscan --leave-temps dumps of generated projects, hand-built
// compressed streams and 473 real Office files; behavior in
// docs/edr/native-engine/parity/macros.md, lab scripts in
// docs/2026-10-08-macro-oracle/ of the workspace). No reference source was
// read.

// errVBAInflate is the reference's decompressor producing nothing: the module
// (or the directory) yields no text.
var errVBAInflate = errors.New("vba: decompression yields no data")

// vbaInflate decompresses compressed VBA source the way the reference does,
// which differs from MS-OVBA: the signature byte and the first chunk header
// are skipped unread; items fill a 4,096-byte window whose copy tokens take
// their bit split from the position in the window, and chunk sizes and
// uncompressed chunks are ignored (every chunk is read as tokens). Only a
// literal item at a full window flushes it and reads the next chunk header
// (2 bytes, discarded), abandoning the rest of the flag byte; a window
// filled by a copy, or by the last item of a flag byte, is overwritten
// instead. The end flushes the partial window. A truncated copy token, a
// missing chunk header after a flushed window, or no output at all fail.
// Every byte produced, flushed or not, is charged to budget; ErrLimit when it
// runs out.
func vbaInflate(ctx context.Context, c []byte, budget *int64) ([]byte, error) {
	if len(c) < 3 {
		return nil, errVBAInflate
	}
	var win [4096]byte
	var out []byte
	pos, flushed := 0, 0
	charge := func(n int) bool {
		*budget -= int64(n)
		return *budget >= 0
	}
	end := func() ([]byte, error) {
		out = append(out, win[:pos&4095]...)
		if len(out) == 0 {
			return nil, errVBAInflate
		}
		return out, nil
	}
	i := 3
	for i < len(c) {
		if i&0xffff == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		flags := c[i]
		i++
		for bit := 0; bit < 8; bit++ {
			if flags>>bit&1 == 0 {
				if pos > 0 && pos&4095 == 0 && pos != flushed {
					out = append(out, win[:]...)
					flushed = pos
					if len(c)-i < 2 {
						return nil, errVBAInflate
					}
					i += 2
					break
				}
				if i >= len(c) {
					return end()
				}
				if !charge(1) {
					return nil, ErrLimit
				}
				win[pos&4095] = c[i]
				i++
				pos++
			} else {
				if len(c)-i < 2 {
					return nil, errVBAInflate
				}
				t := binary.LittleEndian.Uint16(c[i:])
				i += 2
				bc := 4
				if d := pos & 4095; d > 16 {
					bc = bits.Len(uint(d - 1))
				}
				n := int(t&(0xffff>>bc)) + 3
				off := int(t>>(16-bc)) + 1
				if !charge(n) {
					return nil, ErrLimit
				}
				for k := 0; k < n; k++ {
					win[pos&4095] = win[(pos-off)&4095]
					pos++
				}
			}
		}
	}
	return end()
}

// vbaNormalize is the reference's normalization of converted module text:
// a double quote toggles a string, across lines; a tab becomes a space.
// Outside strings A-Z are lower-cased, a run of spaces becomes one space and
// a run right after such a space is dropped even across line ends, and " _"
// joins lines when a line feed follows. After " _" and a carriage return the
// bytes up to the next line feed are kept as they are and that line feed
// removes the two bytes before it.
func vbaNormalize(b []byte) []byte {
	out := make([]byte, 0, len(b))
	inString, space := false, false
	for i := 0; i < len(b); i++ {
		x := b[i]
		if x == '"' {
			inString = !inString
			space = false
			out = append(out, x)
			continue
		}
		if x == '\t' {
			x = ' '
		}
		if !inString {
			if x == '_' && len(out) > 0 && out[len(out)-1] == ' ' && i+1 < len(b) {
				switch b[i+1] {
				case '\n':
					i++
					continue
				case '\r':
					j := bytes.IndexByte(b[i+1:], '\n')
					if j < 0 {
						return append(out, b[i:]...)
					}
					j += i + 1
					out = append(out, b[i:j]...)
					out = out[:len(out)-2]
					i = j
					continue
				}
			}
			if 'A' <= x && x <= 'Z' {
				x += 'a' - 'A'
			}
			if x == ' ' {
				if space {
					continue
				}
				space = true
			} else if x != '\r' && x != '\n' {
				space = false
			}
		}
		out = append(out, x)
	}
	return out
}

// vbaUTF16 converts a record's UTF-16LE text. ok is false where the
// reference's output is not known: an odd length or an unpaired surrogate.
func vbaUTF16(b []byte) (string, bool) {
	if len(b)%2 != 0 {
		return "", false
	}
	u := make([]uint16, len(b)/2)
	for i := range u {
		u[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	for i := 0; i < len(u); i++ {
		switch {
		case u[i] >= 0xd800 && u[i] <= 0xdbff:
			if i+1 == len(u) || u[i+1] < 0xdc00 || u[i+1] > 0xdfff {
				return "", false
			}
			i++
		case u[i] >= 0xdc00 && u[i] <= 0xdfff:
			return "", false
		}
	}
	return string(utf16.Decode(u)), true
}

// vbaDumpLimits bound a project dump.
const vbaRecordLimit = 1 << 20

var vbaSysKind = map[uint32]string{0: "Windows 16 bit", 1: "Windows 32 bit", 2: "Macintosh", 3: "Windows 64 bit"}

const (
	vbaDumpHeader     = "REM VBA project extracted from Microsoft Office document\n\n"
	vbaModuleSep      = "\nREM ##################################################\n"
	vbaStringError    = "<error decoding string>"
	vbaModuleError    = "\n<Error decoding module data>\n"
	vbaModuleCookieID = 0x2c
)

// vbaDump is the reference's dump of one project (a storage holding a dir
// stream). scanned is false when the reference stops reading the directory:
// it then does not scan the dump.
type vbaDump struct {
	text    []byte
	scanned bool
}

// errVBADumpUnknown marks a dump this engine cannot reproduce exactly.
var errVBADumpUnknown = errors.New("vba: project dump not reproducible")

// vbaProjectDump builds the reference's dump from the stream named dir and
// the other streams of its storage (keyed by upper-case name). The dir
// stream is decompressed with the reference's decompressor; records are read
// in order, project records by id (unknown ids skipped) and module records
// in their fixed sequence, whose deviation stops the reader.
func vbaProjectDump(ctx context.Context, dir []byte, streams map[string][]byte, budget *int64) (vbaDump, error) {
	d, err := vbaInflate(ctx, dir, budget)
	if err != nil {
		if errors.Is(err, errVBAInflate) {
			return vbaDump{}, errVBADumpUnknown
		}
		return vbaDump{}, err
	}
	// Before a PROJECTCODEPAGE record the reference converts with 1252.
	w := &vbaDumpWriter{ctx: ctx, d: d, cp: 1252, budget: budget}
	w.out = append(w.out, vbaDumpHeader...)
	scanned, err := w.run(streams)
	if err != nil {
		return vbaDump{}, err
	}
	if !w.charge() {
		return vbaDump{}, ErrLimit
	}
	return vbaDump{text: w.out, scanned: scanned}, nil
}

type vbaDumpWriter struct {
	ctx    context.Context
	d, out []byte
	p      int
	cp     uint16
	// budget is charged for the decompression work and the dump's size.
	budget  *int64
	charged int
}

// charge charges the dump's growth since the last call.
func (w *vbaDumpWriter) charge() bool {
	*w.budget -= int64(len(w.out) - w.charged)
	w.charged = len(w.out)
	return *w.budget >= 0
}

// errVBAStop is the reference's reader stopping: the dump is not scanned.
var errVBAStop = errors.New("vba: directory reader stops")

// record reads the next record. end reports the end of the data at a record
// boundary.
func (w *vbaDumpWriter) record() (id uint16, data []byte, end bool, err error) {
	if w.p == len(w.d) {
		return 0, nil, true, nil
	}
	if len(w.d)-w.p < 6 {
		return 0, nil, false, errVBADumpUnknown
	}
	id = binary.LittleEndian.Uint16(w.d[w.p:])
	size := binary.LittleEndian.Uint32(w.d[w.p+2:])
	w.p += 6
	if id == 0x09 {
		// PROJECTVERSION: Reserved must be 4; VersionMajor and VersionMinor
		// follow (6 bytes).
		if size != 4 {
			return 0, nil, false, errVBAStop
		}
		size = 6
	}
	if uint64(size) > uint64(len(w.d)-w.p) {
		return 0, nil, false, errVBAStop
	}
	data = w.d[w.p : w.p+int(size)]
	w.p += int(size)
	return id, data, false, nil
}

func (w *vbaDumpWriter) peek() (uint16, bool) {
	if len(w.d)-w.p < 2 {
		return 0, false
	}
	return binary.LittleEndian.Uint16(w.d[w.p:]), true
}

// mbcs converts a record's text in the project codepage; a conversion the
// reference fails writes its error string.
func (w *vbaDumpWriter) mbcs(b []byte) ([]byte, bool, error) {
	out, ok, known := vbaConvert(b, w.cp)
	if !known {
		return nil, false, errVBADumpUnknown
	}
	if !ok {
		return []byte(vbaStringError), false, nil
	}
	return out, true, nil
}

func (w *vbaDumpWriter) printf(format string, args ...any) {
	w.out = fmt.Appendf(w.out, format, args...)
}

func u32Of(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, errVBADumpUnknown
	}
	return binary.LittleEndian.Uint32(b), nil
}

func u16Of(b []byte) (uint16, error) {
	if len(b) != 2 {
		return 0, errVBADumpUnknown
	}
	return binary.LittleEndian.Uint16(b), nil
}

// run writes the project records and the modules. It reports whether the
// reader read the whole directory (the dump is scanned).
func (w *vbaDumpWriter) run(streams map[string][]byte) (bool, error) {
	for records := 0; ; records++ {
		if records > vbaRecordLimit || !w.charge() {
			return false, ErrLimit
		}
		if records&1023 == 0 {
			if err := w.ctx.Err(); err != nil {
				return false, err
			}
		}
		id, data, end, err := w.record()
		switch {
		case errors.Is(err, errVBAStop):
			return false, nil
		case err != nil:
			return false, err
		case end:
			return true, nil
		}
		switch id {
		case 0x01:
			v, err := u32Of(data)
			if err != nil {
				return false, err
			}
			if name, ok := vbaSysKind[v]; ok {
				w.printf("REM PROJECTSYSKIND: %s\n", name)
			} else {
				w.printf("REM PROJECTSYSKIND: Unknown (0x%x)\n", v)
			}
		case 0x02, 0x14:
			v, err := u32Of(data)
			if err != nil {
				return false, err
			}
			if id == 0x02 {
				w.printf("REM PROJECTLCID: 0x%08x\n", v)
			} else {
				w.printf("REM PROJECTLCIDINVOKE: 0x%08x\n", v)
			}
		case 0x03:
			v, err := u16Of(data)
			if err != nil {
				return false, err
			}
			w.cp = v
			w.printf("REM PROJECTCODEPAGE: 0x%04x\n", v)
		case 0x04, 0x05:
			s, _, err := w.mbcs(data)
			if err != nil {
				return false, err
			}
			if id == 0x04 {
				w.printf("REM PROJECTNAME: %s\n", s)
			} else {
				w.printf("REM PROJECTDOCSTRING: %s\n", s)
			}
		case 0x40:
			s, ok := vbaUTF16(data)
			if !ok {
				return false, errVBADumpUnknown
			}
			w.printf("REM PROJECTDOCSTRINGUNICODE: %s\n", s)
		case 0x06:
			s, _, err := w.mbcs(data)
			if err != nil {
				return false, err
			}
			w.printf("REM PROJECTHELPFILEPATH: %s\n", s)
			// The second help file path is read only right after the
			// first, as UTF-16.
			if next, ok := w.peek(); ok && next == 0x3d {
				_, data2, _, err := w.record()
				if errors.Is(err, errVBAStop) {
					return false, nil
				} else if err != nil {
					return false, err
				}
				if !bytes.Equal(data, data2) {
					w.printf("REM WARNING: PROJECTHELPFILEPATH and PROJECTHELPFILEPATH2 contents differ\n")
				}
				s2, ok := vbaUTF16(data2)
				if !ok {
					return false, errVBADumpUnknown
				}
				w.printf("REM PROJECTHELPFILEPATH2: %s\n", s2)
			} else {
				w.printf("REM WARNING: PROJECTHELPFILEPATH is not followed by PROJECTHELPFILEPATH2\n")
			}
		case 0x07, 0x08:
			v, err := u32Of(data)
			if err != nil {
				return false, err
			}
			if id == 0x07 {
				w.printf("REM PROJECTHELPCONTEXT: 0x%04x\n", v)
			} else {
				w.printf("REM PROJECTLIBFLAGS: 0x%04x\n", v)
			}
		case 0x09:
			w.printf("REM PROJECTVERSION: %d.%d\n", binary.LittleEndian.Uint32(data), binary.LittleEndian.Uint16(data[4:]))
		case 0x0f:
			v, err := u16Of(data)
			if err != nil {
				return false, err
			}
			w.printf("REM PROJECTMODULES: %d\n", v)
		case 0x13:
			v, err := u16Of(data)
			if err != nil {
				return false, err
			}
			w.printf("REM PROJECTCOOKIE: 0x%04x\n", v)
		case 0x19:
			done, err := w.module(data, streams)
			if err != nil || !done {
				return false, err
			}
		case 0x10:
			return true, nil
		}
	}
}

// vbaModuleSequence is the record order the reference's reader requires
// after MODULENAME.
var vbaModuleSequence = [...]uint16{0x47, 0x1a, 0x32, 0x1c, 0x48, 0x31, 0x1e, vbaModuleCookieID}

// module writes one module from its MODULENAME record. done is false when
// the reader stops.
func (w *vbaDumpWriter) module(name []byte, streams map[string][]byte) (done bool, err error) {
	s, nameOK, err := w.mbcs(name)
	if err != nil {
		return false, err
	}
	w.printf("\n\nREM MODULENAME: %s", s)
	var streamName string
	var offset uint32
	var mbcsStream, mbcsDoc []byte
	for _, want := range vbaModuleSequence {
		id, data, end, err := w.record()
		switch {
		case end:
			return false, errVBADumpUnknown
		case errors.Is(err, errVBAStop):
			return false, nil
		case err != nil:
			return false, err
		case id != want:
			return false, nil
		}
		switch id {
		case 0x47, 0x32, 0x48:
			u, ok := vbaUTF16(data)
			if !ok {
				return false, errVBADumpUnknown
			}
			label, mbcs, mbcsOK := "MODULENAME", name, nameOK
			switch id {
			case 0x32:
				label, mbcs = "MODULESTREAMNAME", mbcsStream
				streamName = strings.ToUpper(u)
			case 0x48:
				label, mbcs = "MODULEDOCSTRING", mbcsDoc
			}
			if id != 0x47 {
				_, mbcsOK, err = w.mbcs(mbcs)
				if err != nil {
					return false, err
				}
			}
			w.printf("\nREM %sUNICODE: %s", label, u)
			if mbcsOK {
				if converted, _, _ := vbaConvert(mbcs, w.cp); string(converted) != u {
					w.printf("\nREM WARNING: %s and %sUNICODE differ", label, label)
				}
			}
		case 0x1a, 0x1c:
			s, _, err := w.mbcs(data)
			if err != nil {
				return false, err
			}
			if id == 0x1a {
				mbcsStream = data
				w.printf("\nREM MODULESTREAMNAME: %s", s)
			} else {
				mbcsDoc = data
				w.printf("\nREM MODULEDOCSTRING: %s", s)
			}
		case 0x31, 0x1e:
			v, err := u32Of(data)
			if err != nil {
				return false, err
			}
			if id == 0x31 {
				offset = v
				w.printf("\nREM MODULEOFFSET: 0x%08x", v)
			} else {
				w.printf("\nREM MODULEHELPCONTEXT: 0x%08x", v)
			}
		case vbaModuleCookieID:
			v, err := u16Of(data)
			if err != nil {
				return false, err
			}
			w.printf("\nREM MODULECOOKIE: 0x%04x", v)
		}
	}
	id, data, end, err := w.record()
	switch {
	case end:
		return false, errVBADumpUnknown
	case errors.Is(err, errVBAStop):
		return false, nil
	case err != nil:
		return false, err
	case id == 0x21 && len(data) == 0:
		w.printf("\nREM MODULETYPE: Procedural")
	case id == 0x22 && len(data) == 0:
		w.printf("\nREM MODULETYPE: Class")
	case id == 0x21 || id == 0x22:
		return false, errVBADumpUnknown
	default:
		return false, nil
	}
	// Optional flags, then the terminator. Other orders were not observed.
	for _, flag := range []uint16{0x25, 0x28} {
		if next, ok := w.peek(); ok && next == flag {
			_, data, _, err := w.record()
			if err != nil || len(data) != 0 {
				return false, errVBADumpUnknown
			}
			if flag == 0x25 {
				w.printf("\nREM MODULEREADONLY")
			} else {
				w.printf("\nREM MODULEPRIVATE")
			}
		}
	}
	id, _, end, err = w.record()
	switch {
	case end || errors.Is(err, errVBAStop):
		return false, errVBADumpUnknown
	case err != nil:
		return false, err
	case id == 0x25 || id == 0x28:
		return false, errVBADumpUnknown
	case id != 0x2b:
		return false, nil
	}
	w.out = append(w.out, vbaModuleSep...)
	text, err := w.moduleText(streams[streamName], offset)
	if err != nil {
		return false, err
	}
	w.out = append(w.out, text...)
	return true, nil
}

// moduleText is a module's source as the dump holds it: decompressed by the
// reference's decompressor from the module offset, converted from the
// project codepage and normalized. A missing stream, an offset past its end
// or a failed decompression give no text; a failed conversion the error
// line.
func (w *vbaDumpWriter) moduleText(stream []byte, offset uint32) ([]byte, error) {
	if stream == nil || uint64(offset) >= uint64(len(stream)) {
		return nil, nil
	}
	raw, err := vbaInflate(w.ctx, stream[offset:], w.budget)
	if errors.Is(err, errVBAInflate) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	converted, ok, known := vbaConvert(raw, w.cp)
	switch {
	case !known:
		return nil, errVBADumpUnknown
	case !ok:
		return []byte(vbaModuleError), nil
	}
	return vbaNormalize(converted), nil
}
