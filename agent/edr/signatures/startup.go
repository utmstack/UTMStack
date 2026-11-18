package signatures

import (
	"context"
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// prepareStartup runs the first loaded BC_STARTUP program, as ClamAV does
// while preparing its bytecode engine. A JIT-disable request leaves this Go
// interpreter unchanged; a full-disable request suppresses every callback in
// the generation. A failed or unsupported startup program must not silently
// leave the generation's other bytecodes enabled. A non-magic return is only
// a warning upstream outside test mode, so it does not reject this feed.
func (d *Database) prepareStartup() error {
	d.bytecodeDisabled = false
	if len(d.startupHooks) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := d.startupHooks[0].RunWithInput(ctx, 0, nil, nil, bytecode.MaxRunSteps)
	if err != nil {
		return fmt.Errorf("startup bytecode could not prepare this feed generation: %w", err)
	}
	if result.BytecodeDisableStatus == 2 {
		d.bytecodeDisabled = true
		d.preclassHooks = nil
		d.unknownPreclass = false
		d.pdfHooks = nil
		clear(d.unknownHooks[:])
		for target := range d.bareHooks {
			d.bareHooks[target] = nil
		}
	}
	return nil
}
