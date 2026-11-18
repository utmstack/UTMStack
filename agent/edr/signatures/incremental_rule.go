package signatures

import (
	"crypto/sha256"
	"errors"
	"maps"
	"path/filepath"
	"slices"
)

// tryRuleSourceRefresh is intentionally limited to one changed top-level NDB.
// Other changes still use the ordinary full compiler until their cross-source
// dependencies and index ordering have been validated separately.
func tryRuleSourceRefresh(files map[string][]byte, limits Limits, options compileOptions) (*Database, bool, error) {
	prior := options.prior
	if prior == nil || prior.limits != limits || prior.sourceDigests == nil || prior.sourceStats == nil {
		return nil, false, nil
	}
	version, names, _, err := sourceIdentity(files, limits)
	if err != nil {
		return nil, true, err
	}
	if len(names) != len(prior.sourceDigests) || !slices.Equal(loadOrder(names), prior.sourceOrder) {
		return nil, false, nil
	}
	var changed string
	digests := make(map[string][sha256.Size]byte, len(names))
	for _, name := range names {
		old, exists := prior.sourceDigests[name]
		if !exists {
			return nil, false, nil
		}
		digest := sha256.Sum256(files[name])
		digests[name] = digest
		if digest != old {
			if changed != "" || filepath.Ext(name) != ".ndb" {
				return nil, false, nil
			}
			changed = name
		}
	}
	if changed == "" {
		return nil, false, nil
	}
	return refreshRuleSource(files, limits, options, version, names, digests, changed)
}

// refreshRuleSource reuses immutable compiled patterns from the prior engine,
// splices one newly compiled NDB rule range into source order, and reconstructs
// only the matcher indexes. Every retained source byte was verified above.
func refreshRuleSource(files map[string][]byte, limits Limits, options compileOptions, version string,
	names []string, digests map[string][sha256.Size]byte, name string) (*Database, bool, error) {
	prior := options.prior
	if len(names) != len(prior.sourceDigests) || !slices.Equal(loadOrder(names), prior.sourceOrder) ||
		len(prior.patternKeys) != len(prior.patterns) {
		return nil, false, nil
	}
	oldStats := prior.sourceStats[name]
	if oldStats == nil || !oldStats.ruleSeen || oldStats.ruleStart < 0 ||
		oldStats.ruleEnd < oldStats.ruleStart || oldStats.ruleEnd > len(prior.rules) {
		return nil, false, nil
	}
	next := *prior
	next.report = prior.Report()
	next.report.Loaded -= oldStats.loaded
	if next.report.Loaded < 0 || next.report.Loaded > limits.MaxSignatures {
		return nil, false, nil
	}
	// Unchanged later sources are included in Loaded before this source is
	// recompiled. Near the limit, only the full compiler can preserve the
	// exact load-order diagnostic for skipped or ignored lines.
	if ruleLines(files[name]) >= limits.MaxSignatures-next.report.Loaded {
		return nil, false, nil
	}
	for reason, count := range oldStats.skipped {
		next.report.Skipped[reason] -= count
		if next.report.Skipped[reason] == 0 {
			delete(next.report.Skipped, reason)
		}
	}
	for reason, count := range oldStats.unknown {
		next.report.Unknown[reason] -= count
		if next.report.Unknown[reason] == 0 {
			delete(next.report.Unknown, reason)
		}
	}
	next.sourceDigests = maps.Clone(prior.sourceDigests)
	next.sourceStats = maps.Clone(prior.sourceStats)
	next.memberStats = maps.Clone(prior.memberStats)
	stats := &sourceCompileStats{skipped: make(map[string]int), unknown: make(map[string]int)}
	next.sourceStats[name] = stats
	next.patterns = append([]bodyPattern(nil), prior.patterns...)
	next.patternKeys = append([][16]byte(nil), prior.patternKeys...)
	next.retiredPatterns = append(bitset(nil), prior.retiredPatterns...)
	next.freshReusedPatterns, next.reusePatternCursor = nil, 0
	// Compiling this one source can append at most one rule per source line.
	// Own the old rules with enough room for those additions so the compiler
	// and the subsequent splice do not each allocate another full rule table.
	next.rules = make([]rule, len(prior.rules), len(prior.rules)+ruleLines(files[name]))
	copy(next.rules, prior.rules)
	next.patternIDs = make(map[[16]byte]int, len(prior.patternKeys))
	for id, key := range prior.patternKeys {
		if id>>6 < len(prior.retiredPatterns) && prior.retiredPatterns.has(id) {
			continue
		}
		next.patternIDs[key] = id
	}
	next.activeSource = name
	beforeRules, beforeLoaded, beforeFiles := len(next.rules), next.report.Loaded, len(next.report.Files)
	errs := &CompileError{}
	addError := func(file string, line int, err error) {
		if len(errs.Diagnostics) < 100 {
			errs.Diagnostics = append(errs.Diagnostics, Diagnostic{File: file, Line: line, Reason: err.Error()})
		}
	}
	next.compileSource(source{name: name, ext: ".ndb", data: files[name]}, addError)
	if len(errs.Diagnostics) > 0 {
		return nil, true, errs
	}
	newRules := append([]rule(nil), next.rules[beforeRules:]...)
	finalRules := len(prior.rules) - (oldStats.ruleEnd - oldStats.ruleStart) + len(newRules)
	copy(next.rules[oldStats.ruleStart+len(newRules):], next.rules[oldStats.ruleEnd:beforeRules])
	copy(next.rules[oldStats.ruleStart:], newRules)
	next.rules = next.rules[:finalRules:finalRules]
	next.report.Files = next.report.Files[:beforeFiles] // same source stays in its original report position
	next.normalizeReportMaps()
	stats.loaded = next.report.Loaded - beforeLoaded
	stats.detections = len(newRules)
	stats.ruleStart, stats.ruleEnd, stats.ruleSeen = oldStats.ruleStart, oldStats.ruleStart+len(newRules), true
	delta := len(newRules) - (oldStats.ruleEnd - oldStats.ruleStart)
	after := false
	for _, sourceName := range next.sourceOrder {
		if sourceName == name {
			after = true
			continue
		}
		if !after || delta == 0 {
			continue
		}
		old := next.sourceStats[sourceName]
		if old == nil || !old.ruleSeen {
			return nil, false, nil
		}
		shifted := *old
		shifted.ruleStart += delta
		shifted.ruleEnd += delta
		next.sourceStats[sourceName] = &shifted
		for _, member := range next.containerMembers[sourceName] {
			full := sourceName + "/" + member.name
			oldMember := next.memberStats[full]
			if oldMember == nil || !oldMember.ruleSeen {
				return nil, false, nil
			}
			shiftedMember := *oldMember
			shiftedMember.ruleStart += delta
			shiftedMember.ruleEnd += delta
			next.memberStats[full] = &shiftedMember
		}
	}
	next.sourceDigests[name] = digests[name]
	next.patternIDs = nil
	next.activeSource = ""
	if next.report.Loaded == 0 {
		return nil, true, errors.New("UTMStack EDR signature database contains no supported signatures")
	}
	if options.baseSources != nil {
		baseDetects := false
		for sourceName := range options.baseSources {
			if sourceStats := next.sourceStats[sourceName]; sourceStats != nil && sourceStats.detections > 0 {
				baseDetects = true
				break
			}
			if shard := next.sourceHashes[sourceName]; shard != nil && shard.len() > 0 {
				baseDetects = true
				break
			}
		}
		if !baseDetects {
			return nil, true, ErrNoDetections
		}
	}
	activePatterns := next.activePatternsForRules(next.rules)
	if err := next.rebuildAnchorIndex(len(prior.patterns), activePatterns); err != nil {
		return nil, true, err
	}
	if err := next.retireInactivePatterns(activePatterns); err != nil {
		return nil, true, err
	}
	next.freshReusedPatterns, next.reusePatternCursor = nil, 0
	next.buildRuleIndex()
	next.textFileMode = false
	next.measureTextRoot()
	next.version = version
	return &next, true, nil
}
