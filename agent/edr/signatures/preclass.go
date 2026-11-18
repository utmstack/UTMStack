package signatures

import (
	"context"
	"fmt"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

type preclassHook struct {
	program *bytecode.Program
	order   int
	bare    bool
}

// HasPreclassHooks reports unconditional or compiled kind-260 hooks. A
// triggered unsupported hook can also flag the phase through MatchContext.
func (d *Database) HasPreclassHooks() bool {
	return len(d.preclassHooks) != 0 || d.unknownPreclass
}

// HasLegacyMetadataRules reports whether root 13 has selected patterns. The
// reference enters its deprecated serialized-JSON pass only when metadata
// collection is active and this root has patterns.
func (d *Database) HasLegacyMetadataRules() bool { return d.legacyMetadataRules }

// MatchLegacyMetadata scans a complete serialization through ordinary typing
// and matching, as the reference does with a nested CL_TYPE_ANY buffer. A
// partial tree cannot safely be matched as JSON: its missing entries change
// both byte offsets and the meaning of negative logical expressions.
func (d *Database) MatchLegacyMetadata(ctx context.Context, serialized []byte, complete bool, mc MatchContext) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if !d.legacyMetadataRules {
		return "", nil
	}
	if !complete || serialized == nil {
		return "", fmt.Errorf("%w: %w", ErrRuleIncomplete, unknownBecause(UnknownMetadataJSON))
	}
	typ, typed := d.DetermineType(serialized)
	if !typed {
		return "", fmt.Errorf("%w: %w", ErrRuleIncomplete, unknownBecause(UnknownMetadataJSON))
	}
	mc.Type, mc.Mode = typ, FileMode
	// The nested pass disables new metadata collection, but its JSON API
	// still reads the existing scan-wide root. The caller supplies the
	// suspended collection options and the original root ancestry.
	mc.PreclassMatched = nil
	if mc.Layers != nil {
		mc.Layers.Push(Layer{Type: typ, Size: int64(len(serialized))})
		defer mc.Layers.Pop()
	}
	res, err := d.MatchObject(ctx, serialized, mc)
	return res.Name, err
}

// MatchPreclass runs the root metadata hook phase after the root scan. The
// flagged set contains only programs whose logical triggers matched at the
// root layer; bare programs run unconditionally. The caller must supply the
// root input, a metadata tree and a bounded recursive child scanner. An
// absent metadata tree makes reached programs undecided rather than clean.
func (d *Database) MatchPreclass(ctx context.Context, root []byte, flagged map[*bytecode.Program]bool, mc MatchContext) (string, error) {
	if len(d.preclassHooks) == 0 && !d.unknownPreclass && !flagged[nil] {
		return "", nil
	}
	w := &matchWork{
		ctx: ctx, limit: d.limits.matchLimit(len(root)),
		bytecodeOptions: mc.BytecodeOptions, metadata: mc.Metadata,
		metadataPartial: mc.MetadataPartial, scanExtracted: mc.ScanExtracted, scanExtractedTyped: mc.ScanExtractedTyped,
	}
	var pending error
	if d.unknownPreclass || flagged[nil] {
		pending = unknownBecause(UnknownBytecode)
	}
	for _, hook := range d.preclassHooks {
		if !hook.bare && !flagged[hook.program] {
			continue
		}
		if mc.Metadata == nil {
			if pending == nil {
				pending = unknownBecause(UnknownBytecode)
			}
			continue
		}
		name, value, cause, err := runBytecode(w, hook.program, root, bytecode.MatchContext{Hook: true})
		if name != "" || err != nil {
			return name, err
		}
		if cause != nil && pending == nil {
			pending = cause
		}
		if value == 0xcea5e { // ClamAV maps BC_PRECLASS CL_BREAK to success.
			break
		}
	}
	if pending != nil {
		return "", fmt.Errorf("%w: %w", ErrRuleIncomplete, pending)
	}
	return "", nil
}
