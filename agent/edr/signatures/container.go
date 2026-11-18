package signatures

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
)

type archiveBudget struct {
	limits Limits
	total  int64
	files  int
}

// member is one file of a signature container. Members keep archive order,
// which is the order the reference loads them in.
type member struct {
	name string
	data []byte
}

func unpackContainer(name string, b []byte, budget *archiveBudget) ([]member, ContainerInfo, error) {
	var files []member
	info, err := walkContainer(name, b, budget, func(memberName string, size int64, r io.Reader, authenticated bool) error {
		var content []byte
		var readErr error
		if authenticated {
			// Only a publisher-authenticated header may drive exact allocation.
			content = make([]byte, int(size))
			_, readErr = io.ReadFull(r, content)
		} else {
			// Local CLDs and unauthenticated CVDs may lie about tar sizes.
			content, readErr = io.ReadAll(io.LimitReader(r, size+1))
		}
		if readErr != nil || int64(len(content)) != size {
			return errors.New("truncated signature archive member")
		}
		files = append(files, member{memberName, content})
		return nil
	})
	if err != nil {
		return nil, info, err
	}
	return files, info, nil
}

// walkContainer validates the same framing and expansion limits for both
// full materialization and bounded member-at-a-time compilation. The visitor
// receives exactly one tar member; unread bytes are drained before the next.
func walkContainer(name string, b []byte, budget *archiveBudget, visit func(string, int64, io.Reader, bool) error) (ContainerInfo, error) {
	info := ContainerInfo{File: name}
	if len(b) < 512 {
		return info, errors.New("truncated signature container")
	}
	header := strings.TrimRight(string(b[:512]), " \x00\r\n")
	fields := strings.Split(header, ":")
	if len(fields) != 9 || fields[0] != "ClamAV-VDB" {
		return info, errors.New("invalid signature container header")
	}
	info.Version = fields[2]
	for _, i := range []int{2, 3, 4, 8} {
		if _, e := parseSize(fields[i]); e != nil || fields[i] == "*" {
			return info, errors.New("invalid numeric signature container header field")
		}
	}
	if strings.EqualFold(path.Ext(name), ".cvd") {
		expected, e := hex.DecodeString(fields[5])
		if e != nil || len(expected) != md5.Size {
			return info, errors.New("invalid signature container checksum")
		}
		actual := md5.Sum(b[512:])
		if !bytes.Equal(expected, actual[:]) {
			return info, errors.New("signature container checksum mismatch")
		}
		info.IntegrityVerified = true
		// The header digest now equals the payload digest; a valid publisher
		// signature over it authenticates the whole container.
		info.Authenticated = verifyContainerSignature(fields[5], fields[6]) == nil
	}
	// Incrementally patched CLD files do not carry a valid publisher signature or
	// original payload checksum. They require external artifact authentication.
	var reader io.Reader = bytes.NewReader(b[512:])
	var z *gzip.Reader
	if len(b) > 513 && b[512] == 0x1f && b[513] == 0x8b {
		var e error
		z, e = gzip.NewReader(reader)
		if e != nil {
			return info, fmt.Errorf("invalid compressed signature container: %w", e)
		}
		defer z.Close()
		reader = z
	}
	limited := &io.LimitedReader{R: reader, N: budget.limits.MaxTotalBytes - budget.total + 1}
	tr := tar.NewReader(limited)
	count := 0
	seen := make(map[string]bool)
	for {
		h, e := tr.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			return info, fmt.Errorf("invalid signature archive: %w", e)
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA {
			return info, errors.New("signature archive contains a non-regular member")
		}
		if h.Name == "" || path.Base(h.Name) != h.Name || strings.ContainsAny(h.Name, "\\\x00:") || h.Name == "." || h.Name == ".." {
			return info, errors.New("signature archive member has unsafe path")
		}
		if seen[h.Name] {
			return info, errors.New("duplicate signature archive member")
		}
		seen[h.Name] = true
		if h.Size < 0 || uint64(h.Size) >= uint64(^uint(0)>>1) || h.Size > budget.limits.MaxFileBytes || h.Size > budget.limits.MaxTotalBytes-budget.total {
			return info, errors.New("signature archive expansion size limit exceeded")
		}
		budget.files++
		if budget.files > budget.limits.MaxFiles {
			return info, errors.New("signature archive file count limit exceeded")
		}
		part := &io.LimitedReader{R: tr, N: h.Size}
		if e = visit(h.Name, h.Size, part, info.Authenticated); e != nil {
			return info, e
		}
		if _, e = io.Copy(io.Discard, part); e != nil || part.N != 0 {
			return info, errors.New("truncated signature archive member")
		}
		budget.total += h.Size
		count++
	}
	// Drain through the decompressor to verify its trailer/checksum and reject
	// trailing expansion bombs after the tar end markers.
	_, e := io.Copy(io.Discard, limited)
	if e != nil {
		return info, fmt.Errorf("signature container checksum/trailer error: %w", e)
	}
	if limited.N <= 0 {
		return info, errors.New("signature archive expanded stream limit exceeded")
	}
	if count == 0 {
		return info, errors.New("empty signature archive")
	}
	return info, nil
}
