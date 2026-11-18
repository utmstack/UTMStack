package signatures

import "errors"

// activePatternsForRules marks the compiled patterns referenced by selected
// rules or file typing. Shared patterns remain active when any retained rule
// or type entry uses them.
func (d *Database) activePatternsForRules(rules []rule) bitset {
	active := newBitset(len(d.patterns))
	for i := range rules {
		for _, id := range rules[i].patterns {
			active.set(id)
		}
	}
	for _, pattern := range d.ftypes.patterns {
		active.set(pattern.id)
	}
	return active
}

// retireInactivePatterns releases compiled bodies no selected rule or file
// type uses. The pattern table itself keeps stable IDs, so published rule and
// file-type references do not need a costly global remap. The caller owns a
// clone of patterns; the previous generation remains immutable.
func (d *Database) retireInactivePatterns(active bitset) error {
	if len(active) < (len(d.patterns)+63)/64 {
		return errors.New("invalid active-pattern retirement boundary")
	}
	retired := newBitset(len(d.patterns))
	copy(retired, d.retiredPatterns)
	for id := range d.patterns {
		if active.has(id) {
			if retired.has(id) {
				return errors.New("active rule references a retired pattern")
			}
			continue
		}
		d.patterns[id] = bodyPattern{}
		retired.set(id)
	}
	d.retiredPatterns = retired
	return nil
}

type anchorGroupKey struct {
	literal string
	fold    bool
}

// rebuildAnchorIndex constructs an index for active compiled patterns without
// mutating the published prior index. Pattern IDs below priorPatternCount have
// had their redundant anchors released; anchorGroups supply those members.
// Freshly compiled patterns after that boundary, or in reused retired slots,
// still hold their anchors.
func (d *Database) rebuildAnchorIndex(priorPatternCount int, active bitset) error {
	if priorPatternCount < 0 || priorPatternCount > len(d.patterns) || len(active) < (len(d.patterns)+63)/64 {
		return errors.New("invalid active-pattern index boundary")
	}
	groups := make([]*anchorGroup, 0, len(d.anchorGroups))
	unanchored := make([]int, 0, len(d.unanchored))
	for _, id := range d.unanchored {
		if id < priorPatternCount && active.has(id) {
			unanchored = append(unanchored, id)
		}
	}
	var byKey map[anchorGroupKey]int
	var owned map[anchorGroupKey]bool
	if priorPatternCount < len(d.patterns) || len(d.freshReusedPatterns) > 0 {
		byKey = make(map[anchorGroupKey]int, len(d.anchorGroups))
		owned = make(map[anchorGroupKey]bool)
	}
	for _, old := range d.anchorGroups {
		kept := 0
		for _, member := range old.members {
			if int(member.id) < priorPatternCount && active.has(int(member.id)) {
				kept++
			}
		}
		if kept == 0 {
			continue
		}
		group := old
		isOwned := false
		if kept != len(old.members) {
			copyGroup := *old
			copyGroup.members = make([]anchorMember, 0, kept)
			for _, member := range old.members {
				if int(member.id) < priorPatternCount && active.has(int(member.id)) {
					copyGroup.members = append(copyGroup.members, member)
				}
			}
			group, isOwned = &copyGroup, true
		}
		if byKey != nil {
			key := anchorGroupKey{literal: string(group.literal), fold: group.fold}
			byKey[key] = len(groups)
			owned[key] = isOwned
		}
		groups = append(groups, group)
	}
	addFresh := func(id int) {
		if !active.has(id) {
			return
		}
		pattern := &d.patterns[id]
		if pattern.compiledAnchorCount() == 0 {
			if pattern.everywhere() {
				unanchored = append(unanchored, id)
			}
			return
		}
		for i := 0; i < pattern.compiledAnchorCount(); i++ {
			anchor := pattern.compiledAnchor(i)
			key := anchorGroupKey{literal: string(anchor.literal), fold: anchor.fold}
			position, found := byKey[key]
			if !found {
				group := &anchorGroup{literal: anchor.literal, fold: anchor.fold, keyAt: keyWindow(anchor.literal, anchor.fold)}
				position = len(groups)
				groups = append(groups, group)
				byKey[key], owned[key] = position, true
			} else if !owned[key] {
				copyGroup := *groups[position]
				copyGroup.members = append([]anchorMember(nil), copyGroup.members...)
				groups[position], owned[key] = &copyGroup, true
			}
			groups[position].members = append(groups[position].members, newAnchorMember(id, anchor))
		}
	}
	for id := range d.freshReusedPatterns.members() {
		if id < priorPatternCount {
			addFresh(id)
		}
	}
	for id := priorPatternCount; id < len(d.patterns); id++ {
		addFresh(id)
	}
	var index anchorIndex
	for _, group := range groups {
		index.add(group)
	}
	for _, bucket := range index.long {
		bucket.split(4, &index)
	}
	for n, table := range index.short {
		for _, bucket := range table {
			bucket.split(n + 1, &index)
		}
	}
	for id := priorPatternCount; id < len(d.patterns); id++ {
		if active.has(id) {
			d.patterns[id].releaseCompiledAnchors()
		}
	}
	for id := range d.freshReusedPatterns.members() {
		if id < priorPatternCount && active.has(id) {
			d.patterns[id].releaseCompiledAnchors()
		}
	}
	d.index, d.anchorGroups, d.unanchored = index, groups, unanchored
	return nil
}
