package signatures

import (
	"errors"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// A CBC source contributes at most one compiled program or one unsupported
// hook. The map is immutable after publication and retains enough provenance
// to replace one archive member without retaining its old runtime hooks.
type bytecodeSourceState struct {
	storage         uint64
	startupHook     *bytecode.Program
	preclassHook    *bytecode.Program
	unknownPreclass bool
	pdfHook         *bytecode.Program
	bareHook        *bytecode.Program
	unknownTarget   int // target + 1; zero means no unconditional unknown hook
}

func (d *Database) recordBytecodeSource(state bytecodeSourceState) {
	name := d.activeMember
	if name == "" {
		name = d.activeSource
	}
	if name == "" {
		return
	}
	if d.bytecodeSources == nil {
		d.bytecodeSources = make(map[string]bytecodeSourceState)
	}
	d.bytecodeSources[name] = state
}

// rebuildBytecodeRuntime restores hook order and global hook tables from
// source order after a CBC member replacement. Rules are already owned by the
// replacement generation; programs from unchanged sources remain shared.
func (d *Database) rebuildBytecodeRuntime() error {
	var storage uint64
	d.startupHooks = nil
	d.preclassHooks = nil
	d.unknownPreclass = false
	d.pdfHooks = nil
	d.bareHooks = [maxTargets][]rule{}
	d.unknownHooks = [maxTargets]string{}
	d.hookCount = 0
	visit := func(name string, stats *sourceCompileStats) {
		state, ok := d.bytecodeSources[name]
		if ok {
			storage += state.storage
			if state.startupHook != nil {
				d.startupHooks = append(d.startupHooks, state.startupHook)
			}
			if state.preclassHook != nil {
				d.preclassHooks = append(d.preclassHooks, preclassHook{program: state.preclassHook, order: d.hookCount, bare: true})
				d.hookCount++
			}
			if state.unknownPreclass {
				d.unknownPreclass = true
			}
			if state.pdfHook != nil {
				d.pdfHooks = append(d.pdfHooks, state.pdfHook)
			}
			if state.unknownTarget > 0 && state.unknownTarget <= maxTargets {
				d.unknownHooks[state.unknownTarget-1] = UnknownBytecode
			}
			if state.bareHook != nil {
				target := hookTargets[state.bareHook.Kind]
				d.bareHooks[target] = append(d.bareHooks[target], rule{ruleCold: &ruleCold{bytecode: state.bareHook, hook: uint16(state.bareHook.Kind), hookOrder: d.hookCount}})
				d.hookCount++
			}
		}
		if stats != nil && stats.ruleSeen {
			for i := stats.ruleStart; i < stats.ruleEnd; i++ {
				if d.rules[i].hook != 0 {
					d.rules[i].cloneCold()
					d.rules[i].hookOrder = d.hookCount
					if d.rules[i].hook == 260 && d.rules[i].bytecode != nil {
						d.preclassHooks = append(d.preclassHooks, preclassHook{program: d.rules[i].bytecode, order: d.hookCount})
					}
					d.hookCount++
				}
			}
		}
	}
	for _, sourceName := range d.sourceOrder {
		if members := d.containerMembers[sourceName]; len(members) > 0 {
			for _, member := range members {
				name := sourceName + "/" + member.name
				visit(name, d.memberStats[name])
			}
		} else {
			visit(sourceName, d.sourceStats[sourceName])
		}
	}
	if storage != d.bytecodeBytes {
		return errors.New("bytecode source storage accounting differs from compiled generation")
	}
	return nil
}
