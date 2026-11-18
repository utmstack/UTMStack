package signatures

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

// sourceCompileStats belongs to one top-level source, including members of an
// archive. Rule ranges give later source-segment refreshes exact provenance.
type sourceCompileStats struct {
	loaded     int
	detections int
	skipped    map[string]int
	unknown    map[string]int
	ruleStart  int
	ruleEnd    int
	ruleSeen   bool
}

// normalizeReportMaps preserves the full compiler's distinction between no
// unknown rules (nil map) and an unknown rule later resolved away, such as a
// logical icon-group rule with no matching IDB entry (allocated empty map).
func (d *Database) normalizeReportMaps() {
	if len(d.report.Skipped) == 0 {
		d.report.Skipped = nil
	}
	if len(d.report.Unknown) != 0 {
		return
	}
	for _, stats := range d.sourceStats {
		for _, count := range stats.unknown {
			if count > 0 {
				if d.report.Unknown == nil {
					d.report.Unknown = make(map[string]int)
				}
				return
			}
		}
	}
	d.report.Unknown = nil
}

func (d *Database) addWholeHash(digest []byte, size int64, name []byte) {
	if d.sourceHashes != nil && d.activeSource != "" {
		shard := d.sourceHashes[d.activeSource]
		if shard == nil {
			shard = newDetectionHashTable()
			d.sourceHashes[d.activeSource] = shard
			d.hashes.shards = append(d.hashes.shards, shard)
		}
		if d.activeMember != "" && d.memberHashes != nil {
			memberShard := d.memberHashes[d.activeMember]
			if memberShard == nil {
				memberShard = newDetectionHashTable()
				if d.activeHashReserve != nil {
					memberShard.reserve(*d.activeHashReserve)
				}
				d.memberHashes[d.activeMember] = memberShard
				shard.shards = append(shard.shards, memberShard)
			}
			memberShard.addBytes(digest, size, name)
			return
		}
		shard.addBytes(digest, size, name)
		return
	}
	d.hashes.addBytes(digest, size, name)
}

// tryHashSourceRefresh preserves every non-hash matcher byte-for-byte. It is
// deliberately narrow: any changed rule, archive, selection file, file set,
// limit, or untracked prior database sends the caller to the full compiler.
func tryHashSourceRefresh(files map[string][]byte, limits Limits, options compileOptions) (*Database, bool, error) {
	prior := options.prior
	if prior == nil || prior.limits != limits || prior.sourceDigests == nil ||
		prior.sourceStats == nil || prior.sourceHashes == nil {
		return nil, false, nil
	}
	version, names, _, err := sourceIdentity(files, limits)
	if err != nil {
		return nil, true, err
	}
	if len(names) != len(prior.sourceDigests) || !slices.Equal(loadOrder(names), prior.sourceOrder) {
		return nil, false, nil
	}
	var changed []string
	digests := make(map[string][sha256.Size]byte, len(names))
	for _, name := range names {
		old, exists := prior.sourceDigests[name]
		if !exists {
			return nil, false, nil
		}
		digest := sha256.Sum256(files[name])
		digests[name] = digest
		if digest != old {
			ext := filepath.Ext(name)
			if ext != ".hdb" && ext != ".hsb" {
				return nil, false, nil
			}
			changed = append(changed, name)
		}
	}
	if len(changed) == 0 {
		return nil, false, nil
	}
	return refreshHashSources(files, limits, options, version, names, digests, changed)
}

// tryIncrementalDirectoryRefresh streams unchanged large archives through the
// content verifier instead of retaining another source copy alongside the old
// engine. Changed hash lists, one changed top-level NDB, or a changed CVD/CLD
// whose changed members are NDB/LDB/HDB/HSB/CBC may be eligible.
func tryIncrementalDirectoryRefresh(dir string, limits Limits, options compileOptions) (*Database, bool, error) {
	prior := options.prior
	if prior == nil || prior.limits != limits || prior.sourceDigests == nil ||
		prior.sourceStats == nil || prior.sourceHashes == nil {
		return nil, false, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, true, err
	}
	identity := sha256.New()
	digests := make(map[string][sha256.Size]byte, len(prior.sourceDigests))
	changedData := make(map[string][]byte)
	var names, changed []string
	var total int64
	buf := make([]byte, 64<<10)
	for _, entry := range entries {
		name := entry.Name()
		if (entry.IsDir() && name == "generations") || metadataFile(name) {
			continue
		}
		if err := ValidateFileName(name); err != nil {
			return nil, true, err
		}
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			return nil, true, fmt.Errorf("UTMStack EDR signature path is not a regular file: %s", name)
		}
		info, err := entry.Info()
		if err != nil {
			return nil, true, err
		}
		if !info.Mode().IsRegular() {
			return nil, true, fmt.Errorf("UTMStack EDR invalid signature file: %s", name)
		}
		if info.Size() < 0 || info.Size() > limits.MaxFileBytes || info.Size() > limits.MaxTotalBytes-total {
			return nil, true, errors.New("UTMStack EDR signature database size limit exceeded")
		}
		names = append(names, name)
		if len(names) > limits.MaxFiles {
			return nil, true, errors.New("UTMStack EDR signature file count limit exceeded")
		}
		old, exists := prior.sourceDigests[name]
		if !exists || len(names) > len(prior.sourceDigests) {
			return nil, false, nil
		}
		total += info.Size()
		fmt.Fprintf(identity, "%d:%s:%d:", len(name), name, info.Size())
		file, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return nil, true, err
		}
		sourceHash := sha256.New()
		n, readErr := io.CopyBuffer(io.MultiWriter(identity, sourceHash), io.LimitReader(file, info.Size()), buf)
		if readErr == nil && n != info.Size() {
			readErr = errors.New("signature file changed or truncated while reading")
		}
		if readErr == nil {
			var extra [1]byte
			_, readErr = io.ReadFull(file, extra[:])
			if readErr == io.EOF {
				readErr = nil
			} else if readErr == nil {
				readErr = errors.New("signature file changed while reading")
			}
		}
		closeErr := file.Close()
		if readErr != nil {
			return nil, true, readErr
		}
		if closeErr != nil {
			return nil, true, closeErr
		}
		var digest [sha256.Size]byte
		copy(digest[:], sourceHash.Sum(nil))
		digests[name] = digest
		if digest != old {
			ext := filepath.Ext(name)
			if ext != ".hdb" && ext != ".hsb" && ext != ".ndb" && ext != ".cvd" && ext != ".cld" {
				return nil, false, nil
			}
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				return nil, true, err
			}
			if int64(len(data)) != info.Size() || sha256.Sum256(data) != digest {
				return nil, true, errors.New("signature file changed while verifying hash refresh")
			}
			changedData[name] = data
			changed = append(changed, name)
		}
	}
	if len(names) != len(prior.sourceDigests) || !slices.Equal(loadOrder(names), prior.sourceOrder) || len(changed) == 0 {
		return nil, false, nil
	}
	version := "go-v1-" + hex.EncodeToString(identity.Sum(nil))
	if allHashSources(changed) {
		return refreshHashSources(changedData, limits, options, version, names, digests, changed)
	}
	if len(changed) == 1 && filepath.Ext(changed[0]) == ".ndb" {
		return refreshRuleSource(changedData, limits, options, version, names, digests, changed[0])
	}
	if allContainerSources(changed) {
		return refreshContainerDeltaSources(changedData, limits, options, version, names, digests, total, changed)
	}
	return nil, false, nil
}

func allContainerSources(names []string) bool {
	if len(names) == 0 {
		return false
	}
	for _, name := range names {
		ext := filepath.Ext(name)
		if ext != ".cvd" && ext != ".cld" {
			return false
		}
	}
	return true
}

func allHashSources(names []string) bool {
	for _, name := range names {
		ext := filepath.Ext(name)
		if ext != ".hdb" && ext != ".hsb" {
			return false
		}
	}
	return true
}

func refreshHashSources(files map[string][]byte, limits Limits, options compileOptions, version string,
	names []string, digests map[string][sha256.Size]byte, changed []string) (*Database, bool, error) {
	prior := options.prior
	if len(names) != len(prior.sourceDigests) || !slices.Equal(loadOrder(names), prior.sourceOrder) {
		return nil, false, nil
	}
	next := *prior
	next.hashes = newWholeHashIndex()
	next.report = prior.Report()
	next.sourceHashes = maps.Clone(prior.sourceHashes)
	next.sourceDigests = maps.Clone(prior.sourceDigests)
	next.sourceStats = maps.Clone(prior.sourceStats)
	next.activeSource = ""
	for _, name := range changed {
		oldStats := prior.sourceStats[name]
		if oldStats == nil {
			return nil, false, nil
		}
		otherLoaded := next.report.Loaded - oldStats.loaded
		if otherLoaded < 0 || otherLoaded > limits.MaxSignatures {
			return nil, false, nil
		}
		temp := newDatabase(limits, prior.level)
		temp.limits.MaxSignatures = limits.MaxSignatures - otherLoaded
		temp.ignore, temp.dconf = prior.ignore, prior.dconf
		temp.sourceHashes = make(map[string]*hashTable, 1)
		temp.sourceStats = map[string]*sourceCompileStats{name: {skipped: make(map[string]int), unknown: make(map[string]int)}}
		temp.activeSource = name
		errs := &CompileError{}
		addError := func(file string, line int, err error) {
			if len(errs.Diagnostics) < 100 {
				errs.Diagnostics = append(errs.Diagnostics, Diagnostic{File: file, Line: line, Reason: err.Error()})
			}
		}
		src := source{name: name, ext: filepath.Ext(name), data: files[name]}
		if prepared := options.prepared[name]; prepared != nil {
			temp.compilePreparedHashList(src, prepared, addError)
		} else {
			temp.compileSource(src, addError)
		}
		if len(errs.Diagnostics) > 0 {
			return nil, true, errs
		}
		shard := temp.sourceHashes[name]
		if shard == nil {
			shard = newDetectionHashTable()
		}
		if err := shard.freeze(); err != nil {
			return nil, true, err
		}
		stats := temp.sourceStats[name]
		stats.loaded = temp.report.Loaded
		stats.detections = shard.len()
		// A hash-list source contains no body rules. Its position between
		// neighboring rule ranges remains unchanged after this refresh.
		stats.ruleStart, stats.ruleEnd, stats.ruleSeen = oldStats.ruleStart, oldStats.ruleEnd, oldStats.ruleSeen
		next.sourceHashes[name] = shard
		next.sourceStats[name] = stats
		next.sourceDigests[name] = digests[name]
		next.report.Loaded += stats.loaded - oldStats.loaded
		for reason, count := range oldStats.skipped {
			next.report.Skipped[reason] -= count
			if next.report.Skipped[reason] == 0 {
				delete(next.report.Skipped, reason)
			}
		}
		for reason, count := range stats.skipped {
			if next.report.Skipped == nil {
				next.report.Skipped = make(map[string]int)
			}
			next.report.Skipped[reason] += count
		}
	}
	if next.report.Loaded == 0 {
		return nil, true, errors.New("UTMStack EDR signature database contains no supported signatures")
	}
	next.normalizeReportMaps()
	if next.report.Loaded > limits.MaxSignatures {
		return nil, true, errors.New("signature count limit exceeded")
	}
	if options.baseSources != nil {
		baseDetects := false
		for name := range options.baseSources {
			if stats := next.sourceStats[name]; stats != nil && stats.detections > 0 {
				baseDetects = true
				break
			}
			if shard := next.sourceHashes[name]; shard != nil && shard.len() > 0 {
				baseDetects = true
				break
			}
		}
		if !baseDetects {
			return nil, true, ErrNoDetections
		}
	}
	for _, name := range next.sourceOrder {
		if shard := next.sourceHashes[name]; shard != nil {
			next.hashes.shards = append(next.hashes.shards, shard)
		}
	}
	if err := next.hashes.freeze(); err != nil {
		return nil, true, err
	}
	next.version = version
	// Equal-class aliases retain the publisher source preference; candidate
	// classification follows the reference. The old tables, rules, and anchor
	// index remain immutable for concurrent scanner readers.
	return &next, true, nil
}
