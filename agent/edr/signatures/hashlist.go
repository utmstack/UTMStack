package signatures

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// PreparedHashList is an opaque, decoded hash-list source. Its digest bytes are
// bound to the exact normalized source bytes; the compiler rejects a different
// source rather than assigning its identity to unrelated matching rules.
type PreparedHashList struct {
	digests         []byte
	width, rowBytes int
	ext             string
	sourceBytes     int
	sourceSHA256    [sha256.Size]byte
}

// ErrEmptyHashList means a syntactically valid list has no indicators after
// blank/comment and empty-object digest filtering. A publisher may omit it.
var ErrEmptyHashList = errors.New("hash list contains no usable indicators")

// PrepareHashList validates whole MD5/SHA1/SHA256 digests and returns both the
// existing durable HDB/HSB representation and compact decoded input for native
// compilation. The durable bytes and names remain compatible with old signed
// generations. Empty-object indicators are filtered as before: the reference
// engine never matches signatures against objects of five bytes or fewer.
func PrepareHashList(data []byte, algorithm string, maxBytes int) ([]byte, *PreparedHashList, error) {
	return prepareHashList(data, algorithm, maxBytes, true)
}

// NormalizeHashList retains the durable format without keeping an additional
// decoded representation when a publisher or a legacy store needs bytes only.
func NormalizeHashList(data []byte, algorithm string, maxBytes int) ([]byte, error) {
	out, _, err := prepareHashList(data, algorithm, maxBytes, false)
	return out, err
}

func prepareHashList(data []byte, algorithm string, maxBytes int, keepDecoded bool) ([]byte, *PreparedHashList, error) {
	p := &PreparedHashList{}
	var empty []byte
	switch algorithm {
	case "MD5":
		p.width, p.ext = md5.Size, ".hdb"
		h := md5.Sum(nil)
		empty = h[:]
	case "SHA1":
		p.width, p.ext = sha1.Size, ".hsb"
		h := sha1.Sum(nil)
		empty = h[:]
	case "SHA256":
		p.width, p.ext = sha256.Size, ".hsb"
		h := sha256.Sum256(nil)
		empty = h[:]
	default:
		return nil, nil, errors.New("unsupported hash-list algorithm")
	}
	if maxBytes <= 0 || int64(maxBytes) > DefaultLimits().MaxFileBytes {
		return nil, nil, errors.New("invalid normalized hash-list size limit")
	}
	var out bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 256), 4096)
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		digest, err := hex.DecodeString(s)
		if err != nil || len(digest) != p.width {
			return nil, nil, errors.New("invalid digest in hash list")
		}
		if bytes.Equal(digest, empty) {
			continue
		}
		s = strings.ToLower(s)
		start := out.Len()
		out.WriteString(s)
		out.WriteString(":*:UTMStack.Intel.")
		out.WriteString(algorithm)
		out.WriteByte('.')
		out.WriteString(s)
		out.WriteString(":73\n")
		if out.Len() > maxBytes {
			return nil, nil, errors.New("normalized signature artifact exceeds size limit")
		}
		p.rowBytes = out.Len() - start
		if keepDecoded {
			p.digests = append(p.digests, digest...)
		}
	}
	if scanner.Err() != nil {
		return nil, nil, errors.New("hash-list line exceeds size limit")
	}
	if out.Len() == 0 {
		return nil, nil, ErrEmptyHashList
	}
	if !keepDecoded {
		return out.Bytes(), nil, nil
	}
	p.sourceBytes = out.Len()
	p.sourceSHA256 = sha256.Sum256(out.Bytes())
	return out.Bytes(), p, nil
}

func (d *Database) compilePreparedHashList(src source, p *PreparedHashList, addError func(string, int, error)) {
	algorithm := ""
	switch p.width {
	case md5.Size:
		if p.ext == ".hdb" {
			algorithm = "MD5"
		}
	case sha1.Size:
		if p.ext == ".hsb" {
			algorithm = "SHA1"
		}
	case sha256.Size:
		if p.ext == ".hsb" {
			algorithm = "SHA256"
		}
	}
	wantRowBytes := p.width*4 + len(":*:UTMStack.Intel.") + len(algorithm) + 1 + len(":73\n")
	if algorithm == "" || p.rowBytes != wantRowBytes || p.sourceBytes <= 0 ||
		len(p.digests) == 0 || len(p.digests)%p.width != 0 ||
		p.sourceBytes%p.rowBytes != 0 || p.sourceBytes/p.rowBytes != len(p.digests)/p.width ||
		src.nested || src.ext != p.ext || len(src.data) != p.sourceBytes || sha256.Sum256(src.data) != p.sourceSHA256 {
		addError(src.name, 0, errors.New("prepared hash list does not match signature source"))
		return
	}
	d.report.Files = append(d.report.Files, src.name)
	_, windowErr := d.hashWindow([]string{"", "", "", "73"})
	for at := 0; at < len(p.digests); at += p.width {
		n := at / p.width
		if d.report.Loaded >= d.limits.MaxSignatures {
			addError(src.name, n+1, errors.New("signature count limit exceeded"))
			return
		}
		if reason, skipped := skipReason(windowErr); skipped {
			d.skipped(reason, 1)
			continue
		}
		if windowErr != nil {
			addError(src.name, n+1, windowErr)
			return
		}
		start := n * p.rowBytes
		entry := string(src.data[start : start+p.rowBytes-1])
		name := string(src.data[start+p.width*2+3 : start+p.rowBytes-4])
		if d.ignore.ignored(name, entry) {
			d.skipped(SkipIgnoreList, 1)
			continue
		}
		d.addWholeHash(p.digests[at:at+p.width], -1, []byte(name))
		d.report.Loaded++
	}
}
