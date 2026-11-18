package signatures

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// ignoreList holds the entries of .ign and .ign2 files. A named rule is never
// loaded: detection rules (.db, .ndb, .ldb, hash files including the .fp/.sfp
// allowlists) and bytecode, whose logical signature name drops the whole
// program. Phishing rules are not named and are never affected.
type ignoreList struct {
	entries map[string]*[md5.Size]byte // stored name -> required rule-line MD5, nil for any
}

// ignoreLineBuffer is the reference loader's line buffer size. The loader
// reuses one buffer and one MD5 pointer across lines, so a name-only line after
// a "name:md5" line keeps pointing at whatever the buffer then holds there;
// load reproduces that byte for byte instead of guessing what an update meant.
const ignoreLineBuffer = 8192

// load adds one ignore-list file. Lines are "name", "name:md5-of-the-rule-line"
// or the old "database:line:name" form, of which only the name is used. Names
// shorter than 3 bytes are space-padded. An empty name (including a blank
// line), a malformed MD5 or more than three fields is an error, as in the
// reference. nested files come from a container, whose lines the reference
// copies without the trailing newline. When one name appears more than once,
// the latest entry decides, which is what the reference does unless the names
// collide in its pattern table.
func (l *ignoreList) load(b []byte, nested bool) (int, error) {
	if l.entries == nil {
		l.entries = map[string]*[md5.Size]byte{}
	}
	var buf [ignoreLineBuffer]byte
	cstring := func(at int) string {
		end := bytes.IndexByte(buf[at:], 0)
		if end < 0 {
			end = len(buf) - at
		}
		return string(buf[at : at+end])
	}
	hash := -1
	line := 0
	for len(b) > 0 {
		raw := b
		newline := false
		if i := bytes.IndexByte(b, '\n'); i >= 0 {
			raw, b, newline = b[:i], b[i+1:], true
		} else {
			b = nil
		}
		line++
		n := copy(buf[:], raw)
		if len(raw)+2 > len(buf) {
			return line, errors.New("ignore-list line exceeds the line buffer")
		}
		if newline && !nested {
			buf[n] = '\n'
			n++
		}
		buf[n] = 0
		if buf[0] == '#' {
			continue
		}
		end := bytes.IndexByte(buf[:], 0)
		for end > 0 && (buf[end-1] == '\n' || buf[end-1] == '\r') {
			end--
			buf[end] = 0
		}
		starts := []int{0}
		for pos := 0; len(starts) <= 3; {
			colon := bytes.IndexByte(buf[pos:end], ':')
			if colon < 0 {
				break
			}
			buf[pos+colon] = 0
			pos += colon + 1
			starts = append(starts, pos)
		}
		name := 0
		switch len(starts) {
		case 1:
		case 2:
			hash = starts[1]
		case 3:
			name = starts[2]
		default:
			return line, errors.New("ignore-list line has more than three fields")
		}
		signame := cstring(name)
		if signame == "" {
			return line, errors.New("ignore-list line has no signature name")
		}
		if len(signame) < 3 {
			copy(buf[:], signame)
			buf[3] = 0
			for i := len(signame); i < 3; i++ {
				buf[i] = ' '
			}
			signame = cstring(0)
		}
		var digest *[md5.Size]byte
		if hash >= 0 {
			text := cstring(hash)
			decoded, err := hex.DecodeString(text)
			if len(text) != 2*md5.Size || err != nil {
				return line, fmt.Errorf("malformed ignore-list MD5 %q", text)
			}
			digest = (*[md5.Size]byte)(decoded)
		}
		l.entries[signame] = digest
	}
	return line, nil
}

// normalizeSignatureName is the name form the reference looks up: a trailing
// ".{...}" clause or ".UNOFFICIAL" suffix is removed and names shorter than 3
// bytes are space-padded.
func normalizeSignatureName(name string) string {
	size := len(name)
	switch {
	case size > 3 && name[size-1] == '}':
		i := strings.Index(name, ".{")
		if i < 0 {
			return name
		}
		size = i
	case size > 11:
		if !strings.HasSuffix(name, ".UNOFFICIAL") {
			return name
		}
		size -= len(".UNOFFICIAL")
	case size > 2:
		return name
	}
	return name[:size] + strings.Repeat(" ", max(0, 3-size))
}

// ignored reports whether a rule named name is dropped. entry is what the
// reference hashes for MD5 entries: the whole rule line with only its line
// ending removed, or the name for bytecode.
func (l *ignoreList) ignored(name, entry string) bool {
	if l == nil || len(l.entries) == 0 {
		return false
	}
	digest, ok := l.entries[normalizeSignatureName(name)]
	if !ok {
		return false
	}
	if digest == nil {
		return true
	}
	if i := strings.IndexByte(entry, 0); i >= 0 {
		entry = entry[:i]
	}
	return md5.Sum([]byte(entry)) == *digest
}

// ignoredBytes is the hash-loader path. A scanner's line buffer is reused on
// its next Scan, so this method only compares the bytes and retains none.
func (l *ignoreList) ignoredBytes(name, entry []byte) bool {
	if l == nil || len(l.entries) == 0 {
		return false
	}
	digest, ok := l.entries[normalizeSignatureName(string(name))]
	if !ok {
		return false
	}
	if digest == nil {
		return true
	}
	if i := bytes.IndexByte(entry, 0); i >= 0 {
		entry = entry[:i]
	}
	return md5.Sum(entry) == *digest
}
