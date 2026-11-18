package bytecode

import (
	"fmt"
	"strings"
)

// ValidateLogical restricts feed activation to the completed logical-callback
// profile. Parse/Run also support standalone generic test programs; their
// successful execution does not make them valid scanner hooks.
func (p *Program) ValidateLogical(functionality uint64) error {
	if p.Kind != 256 || p.LogicalSignature == "" {
		return fmt.Errorf("logical bytecode hook required: %w", ErrUnsupported)
	}
	if p.MinFunctionality > functionality || p.MaxFunctionality != 0 && p.MaxFunctionality < functionality {
		return fmt.Errorf("bytecode functionality window: %w", ErrUnsupported)
	}
	if len(p.functions) == 0 || p.functions[0].args != 0 || p.functions[0].result != 32 {
		return fmt.Errorf("logical bytecode entry prototype: %w", ErrUnsupported)
	}
	// Declared write and extract_new calls are accepted; a reached call
	// requires the enclosing scanner's recursive child callback.
	return nil
}

// ValidateStartup accepts an engine-preparation program. Unlike scanner
// callbacks, BC_STARTUP has no logical trigger and never scans an object.
func (p *Program) ValidateStartup(functionality uint64) error {
	if p.Kind != 1 || strings.Contains(p.LogicalSignature, ";") {
		return fmt.Errorf("startup bytecode without a logical trigger required: %w", ErrUnsupported)
	}
	if p.MinFunctionality > functionality || p.MaxFunctionality != 0 && p.MaxFunctionality < functionality {
		return fmt.Errorf("bytecode functionality window: %w", ErrUnsupported)
	}
	if len(p.functions) == 0 || p.functions[0].args != 0 || p.functions[0].result != 32 {
		return fmt.Errorf("startup bytecode entry prototype: %w", ErrUnsupported)
	}
	return nil
}

// ValidateHook accepts the hook programs this engine runs: kind 257 (PE
// unpacker), 259 (all PE), 260 (root metadata preclass), 261 (ELF unpacker), or 262 (Mach-O unpacker)
// with or without a logical trigger, and kind 258 (PDF) without one. A run that writes
// or finalizes extracted output without a child scanner stops, and one that
// asks for the PDF parser's object model stops with ErrPDFObjects.
func (p *Program) ValidateHook(functionality uint64) error {
	switch {
	case p.Kind == 257 || p.Kind == 259 || p.Kind == 260 || p.Kind == 261 || p.Kind == 262:
	case p.Kind == 258 && !strings.Contains(p.LogicalSignature, ";"):
		// a PDF hook, run on every PDF object: its line holds at most a
		// name, no trigger
	default:
		return fmt.Errorf("executable or PDF hook bytecode required: %w", ErrUnsupported)
	}
	if p.MinFunctionality > functionality || p.MaxFunctionality != 0 && p.MaxFunctionality < functionality {
		return fmt.Errorf("bytecode functionality window: %w", ErrUnsupported)
	}
	if len(p.functions) == 0 || p.functions[0].args != 0 || p.functions[0].result != 32 {
		return fmt.Errorf("hook bytecode entry prototype: %w", ErrUnsupported)
	}
	return nil
}
