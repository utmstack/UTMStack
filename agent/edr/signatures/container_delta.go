package signatures

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
)

// inspectContainerDelta checks the same container framing, checksum, tar
// structure and expansion limits as unpackContainer, but hashes unchanged
// members as streams. Changed NDB/LDB, whole-file hash, and CBC members are
// retained for compilation.
// A structurally valid archive outside this narrow delta shape returns false
// so the ordinary full compiler can handle it.
func inspectContainerDelta(name string, b []byte, budget *archiveBudget, old []containerMemberInfo) (
	changed []member, members []containerMemberInfo, info ContainerInfo, eligible bool, err error,
) {
	info.File = name
	if len(b) < 512 {
		return changed, nil, info, true, errors.New("truncated signature container")
	}
	header := strings.TrimRight(string(b[:512]), " \x00\r\n")
	fields := strings.Split(header, ":")
	if len(fields) != 9 || fields[0] != "ClamAV-VDB" {
		return changed, nil, info, true, errors.New("invalid signature container header")
	}
	info.Version = fields[2]
	for _, i := range []int{2, 3, 4, 8} {
		if _, e := parseSize(fields[i]); e != nil || fields[i] == "*" {
			return changed, nil, info, true, errors.New("invalid numeric signature container header field")
		}
	}
	if strings.EqualFold(path.Ext(name), ".cvd") {
		expected, e := hex.DecodeString(fields[5])
		if e != nil || len(expected) != md5.Size {
			return changed, nil, info, true, errors.New("invalid signature container checksum")
		}
		actual := md5.Sum(b[512:])
		if !bytes.Equal(expected, actual[:]) {
			return changed, nil, info, true, errors.New("signature container checksum mismatch")
		}
		info.IntegrityVerified = true
		info.Authenticated = verifyContainerSignature(fields[5], fields[6]) == nil
	}
	var reader io.Reader = bytes.NewReader(b[512:])
	if len(b) > 513 && b[512] == 0x1f && b[513] == 0x8b {
		z, e := gzip.NewReader(reader)
		if e != nil {
			return changed, nil, info, true, fmt.Errorf("invalid compressed signature container: %w", e)
		}
		defer z.Close()
		reader = z
	}
	limited := &io.LimitedReader{R: reader, N: budget.limits.MaxTotalBytes - budget.total + 1}
	tr := tar.NewReader(limited)
	seen := make(map[string]bool)
	members = make([]containerMemberInfo, 0, len(old))
	buf := make([]byte, 64<<10)
	for {
		h, e := tr.Next()
		if e == io.EOF {
			break
		}
		if e != nil {
			return changed, nil, info, true, fmt.Errorf("invalid signature archive: %w", e)
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA {
			return changed, nil, info, true, errors.New("signature archive contains a non-regular member")
		}
		if h.Name == "" || path.Base(h.Name) != h.Name || strings.ContainsAny(h.Name, "\\\x00:") || h.Name == "." || h.Name == ".." {
			return changed, nil, info, true, errors.New("signature archive member has unsafe path")
		}
		if seen[h.Name] {
			return changed, nil, info, true, errors.New("duplicate signature archive member")
		}
		seen[h.Name] = true
		if h.Size < 0 || uint64(h.Size) >= uint64(^uint(0)>>1) || h.Size > budget.limits.MaxFileBytes || h.Size > budget.limits.MaxTotalBytes-budget.total {
			return changed, nil, info, true, errors.New("signature archive expansion size limit exceeded")
		}
		budget.files++
		if budget.files > budget.limits.MaxFiles {
			return changed, nil, info, true, errors.New("signature archive file count limit exceeded")
		}
		if len(members) >= len(old) || old[len(members)].name != h.Name {
			return nil, nil, info, false, nil
		}
		var digest [sha256.Size]byte
		var content []byte
		ext := filepath.Ext(h.Name)
		if ext == ".ndb" || ext == ".ldb" || ext == ".hdb" || ext == ".hsb" || ext == ".cbc" {
			if info.Authenticated {
				content = make([]byte, int(h.Size))
				_, e = io.ReadFull(tr, content)
			} else {
				content, e = io.ReadAll(io.LimitReader(tr, h.Size+1))
			}
			if e != nil || int64(len(content)) != h.Size {
				return changed, nil, info, true, errors.New("truncated signature archive member")
			}
			digest = sha256.Sum256(content)
		} else {
			hash := sha256.New()
			n, readErr := io.CopyBuffer(hash, io.LimitReader(tr, h.Size), buf)
			if readErr != nil || n != h.Size {
				return changed, nil, info, true, errors.New("truncated signature archive member")
			}
			copy(digest[:], hash.Sum(nil))
		}
		budget.total += h.Size
		members = append(members, containerMemberInfo{name: h.Name, digest: digest, size: h.Size})
		prior := old[len(members)-1]
		if digest != prior.digest || h.Size != prior.size {
			if ext != ".ndb" && ext != ".ldb" && ext != ".hdb" && ext != ".hsb" && ext != ".cbc" {
				return nil, nil, info, false, nil
			}
			changed = append(changed, member{name: h.Name, data: content})
		}
	}
	if len(members) != len(old) || len(changed) == 0 {
		return nil, nil, info, false, nil
	}
	_, e := io.Copy(io.Discard, limited)
	if e != nil {
		return changed, nil, info, true, fmt.Errorf("signature container checksum/trailer error: %w", e)
	}
	if limited.N <= 0 {
		return changed, nil, info, true, errors.New("signature archive expanded stream limit exceeded")
	}
	return changed, members, info, true, nil
}
