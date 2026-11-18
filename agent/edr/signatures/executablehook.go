package signatures

import (
	"slices"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

func (d *Database) executableHooks(target int, flagged []*rule) []*rule {
	if len(d.bareHooks[target]) == 0 {
		return flagged
	}
	all := make([]*rule, 0, len(flagged)+len(d.bareHooks[target]))
	all = append(all, flagged...)
	for i := range d.bareHooks[target] {
		all = append(all, &d.bareHooks[target][i])
	}
	return all
}

// runOtherExecutableHooks executes triggered ELF and Mach-O unpacker hooks
// after the raw rules. The reference invokes their hook points even if its
// header reader cannot parse the file. A fresh hook context has zero match
// counts/offsets and no PE record; extracted output uses the caller's
// bounded recursive child scanner.
func (d *Database) runOtherExecutableHooks(w *matchWork, data []byte, flagged []*rule, target int) (string, error, error) {
	kind := uint16(261)
	if target == 9 {
		kind = 262
	}
	slices.SortFunc(flagged, func(a, b *rule) int { return a.hookOrder - b.hookOrder })
	var pending error
	for _, r := range flagged {
		if r.hook != kind {
			continue
		}
		name, value, cause, err := runBytecode(w, r.bytecode, data, bytecode.MatchContext{Hook: true})
		if w.hookTrace != nil {
			w.hookTrace(r.name, r.hook, value, cause)
		}
		if name != "" || err != nil {
			return name, nil, err
		}
		if cause != nil && pending == nil {
			pending = cause
		}
	}
	return "", pending, nil
}
