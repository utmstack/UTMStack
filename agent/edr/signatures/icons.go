package signatures

// Icon-group rules (IconGroup1/IconGroup2) compare a program's icons with the
// icon database (.idb). A rule's icon condition holds when one icon of the
// program's first group ID, across its resource languages, is close enough
// to one entry that carries both
// of the rule's group names (a missing name is any). Checked on the
// reference: entries of other groups never count, whatever their order, and
// an entry only compares with icons in its 16/24/32-pixel width bucket.
//
// Closeness is ClamAV 1.5.4's confidence formula, including its different
// six-area edge and noedge comparison for grayscale icons. The hash measures
// are computed in iconhash.go and iconedge.go. A valid parsed icon and IDB
// entry therefore have a definite comparison. Bitmap reads use the complete
// physical file, including uint32 header-offset wrap. Modeled directory
// failures terminate the walk; scan budgets return errors.

// iconEntry is one icon database entry: its group names and hash.
type iconEntry struct {
	groups [2]string
	hash   iconHash
}

// resolveIconGroups runs once every file is loaded. The reference compares a
// program's icon only with icon entries in both of a rule's groups; a group
// name no entry carries (or no icon database) leaves nothing to compare, so
// the rule can never match.
func (d *Database) resolveIconGroups() {
	for i := range d.rules {
		r := &d.rules[i]
		if r.icons == nil || d.iconCandidate(*r.icons) {
			continue
		}
		r.cloneCold()
		r.dead = true
		if r.unknown == UnknownIconGroup {
			r.unknown = ""
			if d.report.Unknown[UnknownIconGroup]--; d.report.Unknown[UnknownIconGroup] <= 0 {
				delete(d.report.Unknown, UnknownIconGroup)
			}
		}
	}
}

func inIconGroups(groups, have [2]string) bool {
	in := func(want, have string) bool { return want == "" || want == "*" || want == have }
	return in(groups[0], have[0]) && in(groups[1], have[1])
}

func (d *Database) iconCandidate(groups [2]string) bool {
	for _, e := range d.iconEntries {
		if inIconGroups(groups, e.groups) {
			return true
		}
	}
	return false
}

// iconsComparable reports whether the reference would compare the object's
// icons: always unknown for a header this engine cannot model.
func (info *objectInfo) iconsComparable(data []byte) bool {
	switch info.status {
	case exeUnknown:
		return true
	case exeParsed:
		if info.groupIcon == 0 {
			info.groupIcon = -1
			if info.meta.hasRes && info.meta.hasGroupIcon(data) {
				info.groupIcon = 1
			}
		}
		return info.groupIcon > 0
	}
	return false
}

// iconCondition decides a rule's icon condition on the object.
func (d *Database) iconCondition(info *objectInfo, data []byte, groups [2]string) logicState {
	state, _ := d.iconConditionWithWork(info, data, groups, nil)
	return state
}

func (d *Database) iconConditionWithWork(info *objectInfo, data []byte, groups [2]string, work *matchWork) (logicState, error) {
	if info.status != exeParsed {
		if info.iconsComparable(data) {
			return logicUnknown, nil
		}
		return logicFalse, nil
	}
	// The budgeted group walk below also determines whether a group exists.
	// A separate unbudgeted preflight could otherwise repeat a hostile walk.
	if !info.meta.hasRes {
		return logicFalse, nil
	}
	if info.iconState == 0 {
		info.iconState = -1
		matched := false
		icons, ok, err := info.meta.programIconsWithWork(data, work, func(icon programIcon) (bool, error) {
			state, err := d.iconMatchesGroups(icon, groups, work)
			matched = state == logicTrue
			return matched, err
		})
		if matched {
			// Only this condition was decided. If a later rule must run,
			// it needs its own complete walk rather than this matched prefix.
			info.iconState = 0
			return logicTrue, nil
		}
		if err != nil {
			return logicUnknown, err
		}
		info.icons = icons
		if ok {
			info.iconState = 1
		}
	}
	state := logicFalse
	for _, icon := range info.icons {
		now, err := d.iconMatchesGroups(icon, groups, work)
		if err != nil || now == logicTrue {
			return now, err
		}
		if now == logicUnknown {
			state = logicUnknown
		}
	}
	if info.iconState < 0 {
		return logicUnknown, nil
	}
	return state, nil
}

// Charge both IDB candidate selection and confidence work to the same scan
// budget, including cached images compared by subsequent logical rules.
func (d *Database) iconMatchesGroups(icon programIcon, groups [2]string, work *matchWork) (logicState, error) {
	state := logicFalse
	for _, e := range d.iconEntries {
		if work != nil {
			if err := work.step(1); err != nil {
				return logicUnknown, err
			}
		}
		if e.hash.width != iconComparisonWidth(icon.hash.width) || !inIconGroups(groups, e.groups) {
			continue
		}
		if work != nil {
			if err := work.step(512); err != nil {
				return logicUnknown, err
			}
		}
		now := iconConfidence(icon, e.hash)
		if now == logicTrue {
			return now, nil
		}
		if now == logicUnknown {
			state = logicUnknown
		}
	}
	return state, nil
}

// The value tolerances: a fifth of 255, and of 4072, the largest
// colorfulness.
const (
	iconTolerance8  = 51
	iconTolerance12 = 814
)

// iconMatchFrom is the lowest colored-comparison confidence at a width.
func iconMatchFrom(width int) int { return 80 - width/2 }

// The reference indexes its three IDB buckets by (scaledWidth>>3)-2.
// Fast scaling can also yield 18, 21, 27 or 30 pixels. Geometry uses that
// actual width, while bucket selection and colored thresholds use 16/24/32.
func iconComparisonWidth(width int) int { return width &^ 7 }

// iconConfidence decides whether an icon matches an entry of its width bucket.
func iconConfidence(icon programIcon, e iconHash) logicState {
	g := icon.hash
	grayEntry := e.spread[3] == 0
	bright := iconPointScore(g.bright[:], e.bright[:], g.width, iconTolerance8)
	dark := iconPointScore(g.dark[:], e.dark[:], g.width, iconTolerance8)
	var confidence, threshold int
	if icon.grayMode && grayEntry {
		edge := iconSixPointScore(g.edge, g.color, e.edge, e.color, g.width)
		noedge := iconSixPointScore(g.noedge, g.gray, e.noedge, e.gray, g.width)
		confidence = (bright + dark + 2*edge + noedge) / 6
		threshold = 70
	} else {
		edge := iconPointScore(g.edge[:], e.edge[:], g.width, iconTolerance8)
		noedge := iconPointScore(g.noedge[:], e.noedge[:], g.width, iconTolerance8)
		color, gray := 0, 0
		if !icon.grayMode && !grayEntry {
			color = iconPointScore(g.color[:], e.color[:], g.width, iconTolerance12)
			gray = iconPointScore(g.gray[:], e.gray[:], g.width, iconTolerance12)
		}
		spread := 0
		for i := range g.spread {
			spread += max(0, 100-10*abs(g.spread[i]-e.spread[i]))
		}
		confidence = (color + 2*(gray+bright+noedge)/3 + dark + edge + spread/4) / 6
		threshold = iconMatchFrom(iconComparisonWidth(g.width))
	}
	if confidence >= threshold {
		return logicTrue
	}
	return logicFalse
}

func iconSixPointScore(haveA, haveB, wantA, wantB [3]iconArea, width int) int {
	var have, want [6]iconArea
	copy(have[:3], haveA[:])
	copy(have[3:], haveB[:])
	copy(want[:3], wantA[:])
	copy(want[3:], wantB[:])
	return iconPointScore(have[:], want[:], width, iconTolerance8)
}

func iconPointScore(have, want []iconArea, width, tolerance int) int {
	cut := (width / 4) * 3 / 4
	sum := 0
	for _, a := range have {
		best := 0
		for _, b := range want {
			dx, dy := a.x-b.x, a.y-b.y
			if dist := isqrt(dx*dx + dy*dy); abs(a.v-b.v) <= tolerance && dist <= cut {
				best = max(best, 100-dist*60/cut)
			}
		}
		sum += best
	}
	return sum / len(have)
}

func isqrt(n int) int {
	r := 0
	for (r+1)*(r+1) <= n {
		r++
	}
	return r
}
