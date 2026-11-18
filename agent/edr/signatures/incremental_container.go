package signatures

import (
	"crypto/sha256"
	"errors"
	"maps"
	"path/filepath"
	"slices"
)

// tryContainerSourceRefresh handles existing CVD/CLD sources whose member sets
// and order remain unchanged and whose changed members are NDB/LDB/HDB/HSB/CBC.
// Every changed archive is still checked with the ordinary limits.
func tryContainerSourceRefresh(files map[string][]byte, limits Limits, options compileOptions) (*Database, bool, error) {
	prior := options.prior
	if prior == nil || prior.limits != limits || prior.sourceDigests == nil || prior.containerMembers == nil || prior.memberStats == nil {
		return nil, false, nil
	}
	version, names, total, err := sourceIdentity(files, limits)
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
			if ext != ".cvd" && ext != ".cld" {
				return nil, false, nil
			}
			changed = append(changed, name)
		}
	}
	if len(changed) == 0 {
		return nil, false, nil
	}
	return refreshContainerDeltaSources(files, limits, options, version, names, digests, total, changed)
}

// refreshContainerDeltaSources applies each verified archive delta in load
// order. Intermediate generations are private to this call; only the final
// generation is published. Any ineligible step falls back to the full compiler.
func refreshContainerDeltaSources(files map[string][]byte, limits Limits, options compileOptions,
	version string, names []string, digests map[string][sha256.Size]byte, total int64, changed []string) (*Database, bool, error) {
	if len(changed) == 0 || options.prior == nil {
		return nil, false, nil
	}
	changedSet := make(map[string]bool, len(changed))
	for _, name := range changed {
		changedSet[name] = true
		if _, exists := files[name]; !exists {
			return nil, false, nil
		}
	}
	current := options.prior
	applied := 0
	for _, name := range current.sourceOrder {
		if !changedSet[name] {
			continue
		}
		step := options
		step.prior = current
		if applied < len(changed)-1 {
			// Only the final generation must satisfy the base-source rule.
			step.baseSources = nil
		}
		next, handled, err := refreshContainerDeltaSource(files[name], limits, step, version, names, digests, total, name)
		if err != nil || !handled {
			return nil, handled, err
		}
		current = next
		applied++
	}
	if applied != len(changed) {
		return nil, false, nil
	}
	return current, true, nil
}

// refreshContainerDeltaSource verifies the new archive against the prior member
// manifest, then replaces changed rule, hash and bytecode members in archive order.
func refreshContainerDeltaSource(data []byte, limits Limits, options compileOptions, version string,
	names []string, digests map[string][sha256.Size]byte, total int64, containerName string) (*Database, bool, error) {
	prior := options.prior
	if prior == nil || prior.containerMembers == nil || prior.memberStats == nil ||
		len(prior.patternKeys) != len(prior.patterns) ||
		len(names) != len(prior.sourceDigests) || !slices.Equal(loadOrder(names), prior.sourceOrder) {
		return nil, false, nil
	}
	oldMembers := prior.containerMembers[containerName]
	if len(oldMembers) == 0 {
		return nil, false, nil
	}
	// The full compiler charges all selected compressed bytes and all expanded
	// members against the same limits. Reserve unchanged members before opening
	// the changed archive so the incremental path cannot bypass that budget.
	budget := archiveBudget{limits: limits, total: total, files: len(names)}
	for name, members := range prior.containerMembers {
		if name == containerName {
			continue
		}
		for _, m := range members {
			if m.size < 0 || m.size > limits.MaxTotalBytes-budget.total {
				return nil, false, nil
			}
			budget.total += m.size
			budget.files++
		}
	}
	if budget.files > limits.MaxFiles {
		return nil, false, nil
	}
	changedMembers, newMembers, info, eligible, err := inspectContainerDelta(containerName, data, &budget, oldMembers)
	if err != nil {
		return nil, true, err
	}
	if !eligible {
		return nil, false, nil
	}
	oldSourceStats := prior.sourceStats[containerName]
	if oldSourceStats == nil || !oldSourceStats.ruleSeen {
		return nil, false, nil
	}
	type replacement struct {
		full  string
		ext   string
		data  []byte
		old   *sourceCompileStats
		fresh *sourceCompileStats
		rules []rule
	}
	replacements := make([]*replacement, 0, len(changedMembers))
	byName := make(map[string]*replacement, len(changedMembers))
	hashChanged := false
	bytecodeChanged := false
	lastEnd := -1
	for _, m := range changedMembers {
		full := containerName + "/" + m.name
		ext := filepath.Ext(m.name)
		if ext == ".hdb" || ext == ".hsb" {
			hashChanged = true
			if prior.memberHashes == nil {
				return nil, false, nil
			}
		}
		if ext == ".cbc" {
			bytecodeChanged = true
		}
		old := prior.memberStats[full]
		if old == nil || !old.ruleSeen || old.ruleStart < lastEnd || old.ruleStart < 0 ||
			old.ruleEnd < old.ruleStart || old.ruleEnd > len(prior.rules) {
			return nil, false, nil
		}
		r := &replacement{full: full, ext: ext, data: m.data, old: old}
		replacements = append(replacements, r)
		byName[full] = r
		lastEnd = old.ruleEnd
	}
	next := *prior
	next.report = prior.Report()
	for _, r := range replacements {
		next.report.Loaded -= r.old.loaded
		for reason, count := range r.old.skipped {
			next.report.Skipped[reason] -= count
			if next.report.Skipped[reason] == 0 {
				delete(next.report.Skipped, reason)
			}
		}
		for reason, count := range r.old.unknown {
			next.report.Unknown[reason] -= count
			if next.report.Unknown[reason] == 0 {
				delete(next.report.Unknown, reason)
			}
		}
	}
	if next.report.Loaded < 0 || next.report.Loaded > limits.MaxSignatures {
		return nil, false, nil
	}
	// This compiler starts with unchanged sources already counted, including
	// sources that follow this archive. Near the limit, a skipped line in an
	// earlier changed member could therefore fail here even though the full
	// compiler would accept it in load order. Let the full compiler decide.
	possibleLoaded := next.report.Loaded
	for _, r := range replacements {
		lines := ruleLines(r.data)
		if r.ext == ".cbc" {
			lines = 1
		}
		if lines >= limits.MaxSignatures-possibleLoaded {
			return nil, false, nil
		}
		possibleLoaded += lines
	}
	next.sourceStats = maps.Clone(prior.sourceStats)
	next.memberStats = maps.Clone(prior.memberStats)
	if bytecodeChanged {
		next.bytecodeSources = maps.Clone(prior.bytecodeSources)
		next.startupHooks = slices.Clone(prior.startupHooks)
		next.preclassHooks = slices.Clone(prior.preclassHooks)
		next.pdfHooks = slices.Clone(prior.pdfHooks)
		for target := range next.bareHooks {
			next.bareHooks[target] = slices.Clone(prior.bareHooks[target])
		}
		for _, r := range replacements {
			if r.ext != ".cbc" {
				continue
			}
			state := next.bytecodeSources[r.full]
			if state.storage > next.bytecodeBytes {
				return nil, false, nil
			}
			next.bytecodeBytes -= state.storage
			delete(next.bytecodeSources, r.full)
		}
	}
	next.memberHashes = maps.Clone(prior.memberHashes)
	next.sourceHashes = maps.Clone(prior.sourceHashes)
	if hashChanged {
		// Changed members compile into a new source wrapper; the prior source
		// and all unchanged member tables remain immutable for active scans.
		next.sourceHashes[containerName] = newDetectionHashTable()
	}
	next.sourceDigests = maps.Clone(prior.sourceDigests)
	next.containerMembers = maps.Clone(prior.containerMembers)
	next.containerMembers[containerName] = newMembers
	sourceStats := *oldSourceStats
	sourceStats.skipped = maps.Clone(oldSourceStats.skipped)
	sourceStats.unknown = maps.Clone(oldSourceStats.unknown)
	for _, r := range replacements {
		sourceStats.loaded -= r.old.loaded
		sourceStats.detections -= r.old.detections
		for reason, count := range r.old.skipped {
			sourceStats.skipped[reason] -= count
			if sourceStats.skipped[reason] == 0 {
				delete(sourceStats.skipped, reason)
			}
		}
		for reason, count := range r.old.unknown {
			sourceStats.unknown[reason] -= count
			if sourceStats.unknown[reason] == 0 {
				delete(sourceStats.unknown, reason)
			}
		}
	}
	next.sourceStats[containerName] = &sourceStats
	next.patterns = append([]bodyPattern(nil), prior.patterns...)
	next.patternKeys = append([][16]byte(nil), prior.patternKeys...)
	next.retiredPatterns = append(bitset(nil), prior.retiredPatterns...)
	next.freshReusedPatterns, next.reusePatternCursor = nil, 0
	next.rules = append([]rule(nil), prior.rules...)
	next.patternIDs = make(map[[16]byte]int, len(prior.patternKeys))
	next.expressionIDs = make(map[expressionKey]*expression)
	for id, key := range prior.patternKeys {
		if id>>6 < len(prior.retiredPatterns) && prior.retiredPatterns.has(id) {
			continue
		}
		next.patternIDs[key] = id
	}
	beforeFiles := len(next.report.Files)
	errs := &CompileError{}
	addError := func(file string, line int, err error) {
		if len(errs.Diagnostics) < 100 {
			errs.Diagnostics = append(errs.Diagnostics, Diagnostic{File: file, Line: line, Reason: err.Error()})
		}
	}
	newCapacity := len(prior.rules)
	for _, r := range replacements {
		fresh := &sourceCompileStats{skipped: make(map[string]int), unknown: make(map[string]int)}
		r.fresh = fresh
		next.memberStats[r.full] = fresh
		if r.ext == ".hdb" || r.ext == ".hsb" {
			delete(next.memberHashes, r.full)
		}
		next.activeSource, next.activeMember = containerName, r.full
		beforeRules, beforeLoaded := len(next.rules), next.report.Loaded
		next.compileSource(source{name: r.full, ext: r.ext, data: r.data, nested: true}, addError)
		if len(errs.Diagnostics) > 0 {
			return nil, true, errs
		}
		r.rules = append([]rule(nil), next.rules[beforeRules:]...)
		fresh.loaded = next.report.Loaded - beforeLoaded
		if r.ext == ".cbc" && fresh.loaded != r.old.loaded {
			// CBC sources are only present in Report.Files when selected.
			// A change in selection requires the full compiler's file order.
			return nil, false, nil
		}
		if r.ext == ".ndb" || r.ext == ".ldb" || r.ext == ".cbc" {
			fresh.detections = len(r.rules)
			if r.ext == ".cbc" && (next.bytecodeSources[r.full].bareHook != nil || next.bytecodeSources[r.full].preclassHook != nil) {
				fresh.detections++
			}
		} else {
			if len(r.rules) != 0 {
				return nil, false, nil
			}
			if shard := next.memberHashes[r.full]; shard != nil {
				fresh.detections = shard.len()
			}
		}
		sourceStats.loaded += fresh.loaded
		sourceStats.detections += fresh.detections
		newCapacity += len(r.rules) - (r.old.ruleEnd - r.old.ruleStart)
	}
	if newCapacity < 0 || newCapacity > limits.MaxSignatures {
		return nil, false, nil
	}
	finalRules := make([]rule, 0, newCapacity)
	cursor := 0
	for _, r := range replacements {
		finalRules = append(finalRules, prior.rules[cursor:r.old.ruleStart]...)
		finalRules = append(finalRules, r.rules...)
		cursor = r.old.ruleEnd
	}
	finalRules = append(finalRules, prior.rules[cursor:]...)
	next.rules = finalRules
	next.report.Files = next.report.Files[:beforeFiles]
	delta := 0
	afterContainer := false
	for _, sourceName := range next.sourceOrder {
		if sourceName == containerName {
			for _, m := range newMembers {
				full := containerName + "/" + m.name
				if r := byName[full]; r != nil {
					r.fresh.ruleStart = r.old.ruleStart + delta
					r.fresh.ruleEnd = r.fresh.ruleStart + len(r.rules)
					r.fresh.ruleSeen = true
					delta += len(r.rules) - (r.old.ruleEnd - r.old.ruleStart)
				} else if delta != 0 {
					old := next.memberStats[full]
					if old == nil || !old.ruleSeen {
						return nil, false, nil
					}
					shifted := *old
					shifted.ruleStart += delta
					shifted.ruleEnd += delta
					next.memberStats[full] = &shifted
				}
			}
			sourceStats.ruleEnd += delta
			afterContainer = true
			continue
		}
		if afterContainer && delta != 0 {
			old := next.sourceStats[sourceName]
			if old == nil || !old.ruleSeen {
				return nil, false, nil
			}
			shifted := *old
			shifted.ruleStart += delta
			shifted.ruleEnd += delta
			next.sourceStats[sourceName] = &shifted
			for _, m := range next.containerMembers[sourceName] {
				full := sourceName + "/" + m.name
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
	}
	if !afterContainer {
		return nil, false, nil
	}
	next.sourceDigests[containerName] = digests[containerName]
	containerFound := false
	for i := range next.report.Containers {
		if next.report.Containers[i].File == containerName {
			next.report.Containers[i] = info
			containerFound = true
			break
		}
	}
	if !containerFound {
		return nil, false, nil
	}
	if hashChanged {
		containerHashes := next.sourceHashes[containerName]
		containerHashes.shards = nil
		for _, m := range newMembers {
			if shard := next.memberHashes[containerName+"/"+m.name]; shard != nil {
				containerHashes.shards = append(containerHashes.shards, shard)
			}
		}
		if len(containerHashes.shards) == 0 {
			delete(next.sourceHashes, containerName)
		}
		next.hashes = newWholeHashIndex()
		for _, sourceName := range next.sourceOrder {
			if shard := next.sourceHashes[sourceName]; shard != nil {
				next.hashes.shards = append(next.hashes.shards, shard)
			}
		}
		if err := next.hashes.freeze(); err != nil {
			return nil, true, err
		}
	}
	if bytecodeChanged {
		if err := next.rebuildBytecodeRuntime(); err != nil {
			return nil, true, err
		}
		if err := next.prepareStartup(); err != nil {
			return nil, true, err
		}
	}
	next.activeSource, next.activeMember, next.patternIDs, next.expressionIDs = "", "", nil, nil
	// Logical rules with icon-group descriptors are resolved only after all
	// unchanged IDB members and newly compiled logical rules are present.
	next.resolveIconGroups()
	next.normalizeReportMaps()
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
