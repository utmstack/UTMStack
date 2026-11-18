package signatures

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// GDB full hashes are checked without a host-prefix gate, as in ClamAV
// phishcheck.c (the prefix short circuit is disabled there). The prefix records
// still load so feeds and audits can account for them.
func (d *Database) addSafeBrowsing(line string) error {
	line, err := phishingLevel(line, d.functionality())
	if err != nil {
		return err
	}
	fields := strings.Split(line, ":")
	if len(fields) != 3 {
		return errors.New("invalid GDB signature field count")
	}
	class, kind, value := fields[0], fields[1], fields[2]
	if class != "S" && class != "S1" && class != "S2" {
		return errors.New("unsupported GDB verdict class")
	}
	if kind == "P" {
		if len(value) != 8 {
			return errors.New("GDB host prefix must be four bytes")
		}
		decoded, err := hex.DecodeString(value)
		if err != nil {
			return fmt.Errorf("invalid GDB host prefix: %w", err)
		}
		var prefix [4]byte
		copy(prefix[:], decoded)
		if d.phishing.hostPrefixes == nil {
			d.phishing.hostPrefixes = make(map[[4]byte]struct{})
		}
		d.phishing.hostPrefixes[prefix] = struct{}{}
		return nil
	}
	if kind != "F" && (kind != "W" || class != "S") {
		return errors.New("unsupported GDB hash type")
	}
	if len(value) != 64 {
		return errors.New("GDB full hash must be 32 bytes")
	}
	decoded, err := hex.DecodeString(value)
	if err != nil {
		return fmt.Errorf("invalid GDB full hash: %w", err)
	}
	var digest [32]byte
	copy(digest[:], decoded)
	if d.phishing.hashes == nil {
		d.phishing.hashes = make(map[[32]byte]byte)
	}
	verdict := byte('0')
	if kind == "W" {
		verdict = 'W'
	} else if len(class) == 2 {
		verdict = class[1]
	}
	prior, exists := d.phishing.hashes[digest]
	if exists {
		if prior == 'W' || verdict != 'W' {
			return nil // first hash wins, except a local allow entry overrides it
		}
		d.phishing.hashDetections--
	} else if verdict != 'W' {
		d.phishing.hashDetections++
	}
	d.phishing.hashes[digest] = verdict
	return nil
}

const safeBrowsingURLLimit = 1024

// canonicalSafeBrowsingURL returns the host and path that the GDB hash loop
// combines. Forms with ambiguous byte-level semantics remain incomplete rather
// than silently clean. No DNS lookup or request is made.
func canonicalSafeBrowsingURL(raw string) (host, path string, eligible bool, err error) {
	if raw == "" {
		return "", "", false, nil
	}
	colon := strings.IndexByte(raw, ':')
	if colon < 0 {
		return "", "", false, nil
	}
	// The pinned isURL fast path is case-sensitive. A mailto:// URL also
	// reaches its hash check, despite the usual mailto: form not doing so.
	scheme := raw[:colon]
	if scheme != "http" && scheme != "https" && scheme != "ftp" && scheme != "mailto" {
		return "", "", false, nil
	}
	if !strings.HasPrefix(raw[colon+1:], "//") {
		return "", "", true, fmt.Errorf("%w: GDB URL without canonical scheme slashes", ErrHTMLIncomplete)
	}
	if len(raw) > safeBrowsingURLLimit {
		return "", "", true, fmt.Errorf("%w: GDB URL length limit exceeded", ErrHTMLIncomplete)
	}
	authority := strings.TrimLeft(raw[colon+1:], "/")
	if authority == "" {
		return "", "", false, nil
	}
	end := strings.IndexAny(authority, ":/?")
	if end < 0 {
		end = len(authority)
	}
	if at := strings.LastIndexByte(authority[:end], '@'); at >= 0 {
		authority = authority[at+1:]
	}
	if strings.ContainsAny(authority, "\\\x00") {
		return "", "", true, fmt.Errorf("%w: ambiguous GDB URL authority", ErrHTMLIncomplete)
	}
	// The pinned decoder strips the first leading escape separately, then
	// repeatedly decodes escapes in the remainder only while a new '%' is
	// produced. Each successful rewrite shortens the buffer, so no arbitrary
	// pass limit is needed.
	authority = decodeSafeBrowsingPercent(authority)
	authority = strings.TrimLeft(authority, "/")
	authority = strings.TrimLeft(authority, ".")
	if authority == "" {
		return "", "", false, nil
	}
	if authority[0] == '/' {
		return "", "", true, fmt.Errorf("%w: mixed leading GDB URL separators", ErrHTMLIncomplete)
	}
	authority, err = normalizeSafeBrowsingDots(authority)
	if err != nil {
		return "", "", true, err
	}
	if strings.Contains(authority, "\x00") {
		return "", "", true, fmt.Errorf("%w: ambiguous GDB URL delimiter", ErrHTMLIncomplete)
	}
	hostEnd := strings.IndexAny(authority, ":/?")
	if hostEnd < 0 {
		hostEnd = len(authority)
	}
	host = authority[:hostEnd]
	if host == "" {
		return "", "", true, fmt.Errorf("%w: ambiguous GDB hostname", ErrHTMLIncomplete)
	}
	host = strings.ToLower(escapeSafeBrowsingBytes(host))
	if len(host) > safeBrowsingURLLimit {
		return "", "", true, fmt.Errorf("%w: GDB canonical hostname length limit", ErrHTMLIncomplete)
	}
	rest := authority[hostEnd:]
	if rest == "" || rest == "/" {
		return host, "", true, nil
	}
	if !strings.ContainsAny(rest[:1], "/:?") {
		return "", "", true, fmt.Errorf("%w: ambiguous GDB URL path", ErrHTMLIncomplete)
	}
	// The pinned reference inserts a slash before the first colon, question
	// mark or slash after the host, then hashes the bytes following it. This
	// makes a port or a query without a slash part of the path component.
	path = rest[1:]
	if strings.Contains(path, "\x00") {
		return "", "", true, fmt.Errorf("%w: GDB path normalization pending", ErrHTMLIncomplete)
	}
	path = escapeSafeBrowsingBytes(path)
	if len(host)+len(path) > safeBrowsingURLLimit {
		return "", "", true, fmt.Errorf("%w: GDB canonical URL length limit", ErrHTMLIncomplete)
	}
	return host, path, true, nil
}

func escapeSafeBrowsingBytes(s string) string {
	var escaped strings.Builder
	const digits = "0123456789ABCDEF"
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c <= 32 || c >= 127 || c == '%' || c == '#' {
			escaped.WriteByte('%')
			escaped.WriteByte(digits[c>>4])
			escaped.WriteByte(digits[c&15])
		} else {
			escaped.WriteByte(c)
		}
	}
	return escaped.String()
}

func decodeSafeBrowsingPercent(s string) string {
	end := len(s)
	if end <= 1 {
		return s
	}
	buf := make([]byte, end+3)
	copy(buf, s)
	start := 0
	if buf[0] == '%' {
		buf[2] = safeBrowsingHexValue(buf[1], buf[2])
		start = 2
	}
	passStart := start + 1
	for {
		producedPercent := false
		for pos := passStart; pos+3 <= end; pos++ {
			if buf[pos] != '%' || hexNibble(buf[pos+1]) < 0 || hexNibble(buf[pos+2]) < 0 {
				continue
			}
			buf[pos] = safeBrowsingHexValue(buf[pos+1], buf[pos+2])
			producedPercent = producedPercent || buf[pos] == '%'
			copy(buf[pos+1:], buf[pos+3:end+1]) // includes the terminal zero
			end -= 2
		}
		if !producedPercent {
			break
		}
	}
	return string(buf[start:end])
}

func safeBrowsingHexValue(hi, lo byte) byte {
	if hi == '0' && lo == '0' {
		return 1 // the pinned cloak check keeps %00 visible as 0x01
	}
	upper := hexNibble(hi)
	lower := hexNibble(lo)
	if upper < 0 {
		upper = 0
	}
	if lower < 0 {
		lower = 0
	}
	return byte(upper<<4 | lower)
}

// The pinned canonicalizer walks a mutable byte buffer once. After a rewrite
// it advances from the old position, which can leave a later dot segment in
// place. Preserving that behavior matters because GDB hashes exact URL bytes.
func normalizeSafeBrowsingDots(s string) (string, error) {
	buf := make([]byte, len(s)+3) // room for the reference's terminal lookahead
	copy(buf, s)
	end := len(s)
	last := strings.IndexByte(s, '/')
	for pos := 0; pos < end; pos++ {
		if pos+2 < end && buf[pos] == '/' && buf[pos+1] == '.' {
			switch {
			case buf[pos+2] == '/':
				if pos+3 < end {
					copy(buf[pos+1:], buf[pos+3:end])
				}
				end -= 2
			case buf[pos+2] == '.' && (buf[pos+3] == '/' || buf[pos+3] == 0) && last >= 0:
				if pos+4 < end {
					copy(buf[last+1:], buf[pos+4:end])
				}
				end -= 3 + pos - last
			}
		}
		if buf[pos] == '/' {
			last = pos
		}
	}
	return string(buf[:end]), nil
}

func (p phishingRules) matchSafeBrowsingURL(raw string, w *matchWork) (string, error) {
	host, path, eligible, err := canonicalSafeBrowsingURL(raw)
	if err != nil || !eligible {
		return "", err
	}
	labels := strings.Split(host, ".")
	if len(labels) == 0 {
		return "", nil
	}
	var hosts []string
	for components := 2; components <= 5 && components <= len(labels); components++ {
		hosts = append(hosts, strings.Join(labels[len(labels)-components:], "."))
	}
	if len(hosts) == 0 || hosts[len(hosts)-1] != host {
		hosts = append(hosts, host)
	}
	// The C matcher visits directory prefixes, root, queryless path and full
	// path in that order. Duplicate candidates are harmless but omitted here.
	paths := make([]string, 0, 7)
	directoryLimit := 4
	if strings.IndexByte(path, '?') >= 0 {
		// The separate queryless candidate occupies one of the reference's
		// six path slots, leaving room for only three directory prefixes.
		directoryLimit = 3
	}
	for search := 1; search < len(path) && len(paths) < directoryLimit; {
		i := strings.IndexByte(path[search:], '/')
		if i < 0 {
			break
		}
		i += search
		paths = append(paths, path[:i+1])
		search = i + 2 // the reference starts one byte after each prefix
	}
	for i, j := 0, len(paths)-1; i < j; i, j = i+1, j-1 {
		paths[i], paths[j] = paths[j], paths[i]
	}
	paths = append(paths, "")
	if query := strings.IndexByte(path, '?'); query >= 0 {
		paths = append(paths, path[:query])
	}
	paths = append(paths, path)
	for _, candidatePath := range paths {
		for _, candidateHost := range hosts {
			if err := w.step(int64(len(candidateHost) + len(candidatePath) + 1)); err != nil {
				return "", err
			}
			digest := sha256.Sum256([]byte(candidateHost + "/" + candidatePath))
			switch p.hashes[digest] {
			case '1':
				return "Heuristics.Phishing.URL.Blocked", nil
			case '2':
				return "Heuristics.Safebrowsing.Suspected-phishing_safebrowsing.clamav.net", nil
			case '0':
				return "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net", nil
			}
		}
	}
	return "", nil
}
