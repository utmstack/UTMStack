package signatures

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"path/filepath"
	"strings"
)

type hashGroupReserve struct {
	count, nameBytes int
}

type hashReserve [3]hashGroupReserve

func hashReserveFormat(ext string) bool {
	switch ext {
	case ".hdb", ".hsb", ".mdb", ".msb", ".fp", ".sfp":
		return true
	}
	return false
}

func patternReserveFormat(ext string) bool {
	switch ext {
	case ".ldb", ".ndb", ".db":
		return true
	}
	return false
}

func patternCandidates(ext string, line []byte) int {
	line = bytes.TrimSpace(line)
	if len(line) == 0 || line[0] == '#' {
		return 0
	}
	if ext == ".ldb" {
		return max(0, bytes.Count(line, []byte{';'})-2)
	}
	return 1
}

func (reserve *hashReserve) observe(ext string, line []byte) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 || line[0] == '#' {
		return
	}
	first := bytes.IndexByte(line, ':')
	if first < 0 {
		return
	}
	second := bytes.IndexByte(line[first+1:], ':')
	if second < 0 {
		return
	}
	second += first + 1
	digest := line[:first]
	if ext == ".mdb" || ext == ".msb" {
		digest = line[first+1 : second]
	}
	name := line[second+1:]
	if end := bytes.IndexByte(name, ':'); end >= 0 {
		name = name[:end]
	}
	if len(name) == 0 || len(name) > 256 {
		return
	}
	index := -1
	switch len(digest) {
	case 32:
		index = 0
	case 40:
		index = 1
	case 64:
		index = 2
	}
	if index >= 0 {
		reserve[index].count++
		reserve[index].nameBytes += int(hashNameEncodedSize(len(name)))
	}
}

func (reserve hashReserve) bounded(limits Limits) bool {
	var count, bytes int64
	for i, group := range reserve {
		count += int64(group.count)
		bytes += int64(group.nameBytes) + int64(group.count)*int64([3]int{20, 24, 36}[i])
	}
	return count > 0 && count <= int64(limits.MaxSignatures) && bytes <= 256<<20
}

// preflightContainer authenticates and bounds every member without keeping
// large expanded members live. Only global selection inputs must survive to
// the first compilation pass.
func preflightContainer(name string, b []byte, budget *archiveBudget) ([]source, []containerMemberInfo, ContainerInfo, error) {
	var sources []source
	var members []containerMemberInfo
	buf := make([]byte, 64<<10)
	info, err := walkContainer(name, b, budget, func(memberName string, size int64, r io.Reader, authenticated bool) error {
		ext := strings.ToLower(filepath.Ext(memberName))
		full := name + "/" + memberName
		var data []byte
		var digest [sha256.Size]byte
		var reserve hashReserve
		candidates := 0
		if ext == ".ign" || ext == ".ign2" || ext == ".cfg" {
			var err error
			data, err = readContainerMember(r, size, authenticated)
			if err != nil {
				return err
			}
			digest = sha256.Sum256(data)
		} else if authenticated && size >= 8<<20 && hashReserveFormat(ext) {
			h := sha256.New()
			scanner := bufio.NewScanner(io.TeeReader(r, h))
			scanner.Buffer(make([]byte, 4096), budget.limits.MaxLineBytes)
			for scanner.Scan() {
				reserve.observe(ext, scanner.Bytes())
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			copy(digest[:], h.Sum(nil))
			if !reserve.bounded(budget.limits) {
				reserve = hashReserve{}
			}
		} else if authenticated && size >= 8<<20 && patternReserveFormat(ext) {
			h := sha256.New()
			scanner := bufio.NewScanner(io.TeeReader(r, h))
			scanner.Buffer(make([]byte, 4096), budget.limits.MaxLineBytes)
			tooManyCandidates := false
			for scanner.Scan() {
				if !tooManyCandidates {
					candidates += patternCandidates(ext, scanner.Bytes())
					if candidates > 2_000_000 {
						candidates = 0 // disable hints for pathological members
						tooManyCandidates = true
					}
				}
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			copy(digest[:], h.Sum(nil))
		} else {
			h := sha256.New()
			n, err := io.CopyBuffer(h, r, buf)
			if err != nil || n != size {
				return errors.New("truncated signature archive member")
			}
			copy(digest[:], h.Sum(nil))
		}
		sources = append(sources, source{full, ext, data, true})
		members = append(members, containerMemberInfo{name: memberName, digest: digest, size: size, reserve: reserve, patternCandidates: candidates})
		return nil
	})
	if err != nil {
		return nil, nil, info, err
	}
	return sources, members, info, nil
}

func readContainerMember(r io.Reader, size int64, authenticated bool) ([]byte, error) {
	var data []byte
	var err error
	if authenticated {
		data = make([]byte, int(size))
		_, err = io.ReadFull(r, data)
	} else {
		data, err = io.ReadAll(io.LimitReader(r, size+1))
	}
	if err != nil || int64(len(data)) != size {
		return nil, errors.New("truncated signature archive member")
	}
	return data, nil
}

func streamableTextSource(ext string) bool {
	switch ext {
	case ".hdb", ".hsb", ".mdb", ".msb", ".fp", ".sfp", ".ndb", ".ldb", ".db", ".pdb", ".wdb", ".gdb", ".idb", ".crb", ".cdb":
		return true
	}
	return false
}

// replayContainer compiles each member as the validated archive advances.
// A digest comparison protects the preflight-to-compile boundary, including
// when a compiler intentionally reads only part of a member.
func (d *Database) replayContainer(name string, b []byte, budget *archiveBudget, sources []source, expected []containerMemberInfo,
	prepared map[string]*PreparedHashList, compile func(source, io.Reader)) error {
	index := 0
	_, err := walkContainer(name, b, budget, func(memberName string, size int64, r io.Reader, authenticated bool) error {
		if index >= len(sources) || index >= len(expected) || sources[index].name != name+"/"+memberName || expected[index].name != memberName || expected[index].size != size {
			return errors.New("signature container member changed during compilation")
		}
		src := sources[index]
		index++
		hint := expected[index-1].reserve
		if authenticated {
			switch src.ext {
			case ".mdb", ".msb":
				d.sections.reserve(hint)
			case ".fp", ".sfp":
				d.allow.reserve(hint)
			case ".hdb", ".hsb":
				d.activeHashReserve = &hint
			}
		}
		h := sha256.New()
		tee := io.TeeReader(r, h)
		if src.ext == ".ign" || src.ext == ".ign2" || src.ext == ".cfg" || src.ext == ".info" {
			compile(src, nil)
		} else if streamableTextSource(src.ext) && !d.dconfDisabled(src.ext) && prepared[src.name] == nil {
			compile(src, tee)
		} else {
			data, readErr := readContainerMember(tee, size, authenticated)
			if readErr != nil {
				return readErr
			}
			src.data = data
			compile(src, nil)
		}
		d.activeHashReserve = nil
		if _, err := io.Copy(io.Discard, tee); err != nil {
			return err
		}
		var digest [sha256.Size]byte
		copy(digest[:], h.Sum(nil))
		if digest != expected[index-1].digest {
			return errors.New("signature container member changed during compilation")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if index != len(sources) || index != len(expected) {
		return errors.New("signature container member changed during compilation")
	}
	return nil
}
