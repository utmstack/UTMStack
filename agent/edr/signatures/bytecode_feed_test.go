package signatures

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// Set EDR_OFFICIAL_SIGNATURE_DIR to a directory holding the published
// bytecode.cvd to run every official program this engine loads on a small PE
// with real x86 code at its entry point and on a small ELF, the way a
// triggered rule runs them (same step allowance). None may fail at run time:
// a failed run leaves its rule undecided.
func TestOfficialBytecodeRunsOnSmallObjects(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_SIGNATURE_DIR")
	if dir == "" {
		t.Skip("EDR_OFFICIAL_SIGNATURE_DIR not set")
	}
	b, err := os.ReadFile(filepath.Join(dir, "bytecode.cvd"))
	if err != nil {
		t.Fatal(err)
	}
	budget := archiveBudget{limits: Limits{}.defaults(), total: int64(len(b)), files: 1}
	members, _, err := unpackContainer("bytecode.cvd", b, &budget)
	if err != nil {
		t.Fatal(err)
	}
	pe := testPE()
	code := []byte{
		0x55, 0x8b, 0xec, 0x83, 0xec, 0x20, 0x53, 0x56, 0x57, // push ebp; mov ebp, esp; sub esp, 32; push ebx, esi, edi
		0x8b, 0x45, 0x08, 0x8b, 0x4d, 0x0c, 0x03, 0xc1, 0x89, 0x45, 0xfc, // mov eax, [ebp+8]; mov ecx, [ebp+12]; add eax, ecx; mov [ebp-4], eax
		0xe8, 0x00, 0x00, 0x00, 0x00, 0x5f, 0x5e, 0x5b, // call $+5; pop edi, esi, ebx
		0x8b, 0xe5, 0x5d, 0xc2, 0x08, 0x00, // mov esp, ebp; pop ebp; ret 8
	}
	copy(pe[512:], append(code, bytes.Repeat([]byte{0xcc}, 512-len(code))...))
	meta := parsePE(pe)
	if meta == nil {
		t.Fatal("test PE not parsed")
	}
	elf := elf64(binary.LittleEndian, 1, 0x400100, [][4]uint64{{1, 0, 0x400000, 0x1000}}, 3, 64)
	objects := []struct {
		name string
		data []byte
		pe   *bytecode.PEInfo
	}{{"PE", pe, meta.bytecodePE()}, {"ELF", elf, nil}}
	hookHeader, st := readHookHeader(pe, true)
	if st != hookParsed {
		t.Fatalf("test PE hook reading: %d", st)
	}
	ran, hooks := 0, 0
	for _, m := range members {
		if !strings.HasSuffix(m.name, ".cbc") {
			continue
		}
		p, err := bytecode.Parse(m.data)
		if err == nil && p.ValidateHook(FunctionalityLevel) == nil {
			// A PE hook runs in the hook phase with the hook context.
			matches := bytecode.MatchContext{PE: hookHeader.bytecodePE(), Hook: true}
			if _, err := p.RunWithMatchContext(context.Background(), 0, nil, pe, matches, 1_000_000); err != nil {
				t.Errorf("hook %s on the PE: %v", m.name, err)
			}
			hooks++
			continue
		}
		if err == nil {
			err = p.ValidateLogical(FunctionalityLevel)
		}
		if err != nil {
			continue // loaded as unknown when reached
		}
		for _, o := range objects {
			matches := bytecode.EmptyMatchContext()
			matches.PE = o.pe
			if _, err := p.RunWithMatchContext(context.Background(), 0, nil, o.data, matches, 1_000_000); err != nil {
				t.Errorf("%s on the %s: %v", m.name, o.name, err)
			}
		}
		ran++
	}
	t.Logf("%d official programs ran on both objects, %d hook programs on the PE", ran, hooks)
	if ran == 0 || hooks == 0 {
		t.Fatal("no official program ran")
	}
}
