package signatures

import (
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

func TestBytecodeCompiledStorageBudgetIsAtomic(t *testing.T) {
	// A compact wire program expands into types, instructions, and initialized
	// global storage. The compiler must account for the representation it keeps.
	f := callbackWithLabel([]byte("UTMStack."+strings.Repeat("A", 1024)), true)
	wire := f.encode()
	p, err := bytecode.Parse(wire)
	if err != nil {
		t.Fatal(err)
	}
	storage := p.StorageBytes()
	if storage <= uint64(len(wire)) {
		t.Fatalf("fixture does not expand: wire=%d storage=%d", len(wire), storage)
	}
	if _, err := CompileWithLimits(map[string][]byte{"a.cbc": wire}, Limits{MaxTotalBytes: int64(storage - 1)}); err == nil {
		t.Fatal("accepted expanded bytecode over storage budget")
	}
	if db, err := CompileWithLimits(map[string][]byte{"a.cbc": wire}, Limits{MaxTotalBytes: int64(storage)}); err != nil || db == nil {
		t.Fatalf("rejected bytecode at storage budget: %v", err)
	}
	if _, err := CompileWithLimits(map[string][]byte{"a.cbc": wire, "b.cbc": wire}, Limits{MaxTotalBytes: int64(2*storage - 1)}); err == nil {
		t.Fatal("accepted aggregate bytecode over storage budget")
	}
}
