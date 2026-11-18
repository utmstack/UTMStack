package bytecode

import (
	"bytes"
	"errors"
	"math"
	"testing"
)

func TestVirtualMemoryIsolationAndInitialization(t *testing.T) {
	var m memory
	ptr, err := m.alloc(16, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.read(ptr, 1); !errors.Is(err, ErrRuntime) {
		t.Fatal("uninitialized", err)
	}
	if err := m.store(ptr+3, 4, 0x12345678); err != nil {
		t.Fatal(err)
	}
	data, err := m.read(ptr+3, 4)
	if err != nil || !bytes.Equal(data, []byte{0x78, 0x56, 0x34, 0x12}) {
		t.Fatal(data, err)
	}
	if _, err := m.read(ptr+2, 5); !errors.Is(err, ErrRuntime) {
		t.Fatal("partially initialized", err)
	}
	for _, bad := range []uint64{0, 1, ptr + 16, ptr + (1 << 32), math.MaxUint64} {
		if _, err := m.read(bad, 1); !errors.Is(err, ErrRuntime) {
			t.Fatalf("read %#x: %v", bad, err)
		}
		if err := m.store(bad, 1, 1); !errors.Is(err, ErrRuntime) {
			t.Fatalf("write %#x: %v", bad, err)
		}
	}
	if _, err := m.read(ptr+16, 0); err != nil {
		t.Fatal("one-past zero read", err)
	}
	if err := m.fill(ptr, 0x80, 16); err != nil {
		t.Fatal(err)
	}
	if value, err := m.load(ptr+1, 8); err != nil || value != 0x8080808080808080 {
		t.Fatal(value, err)
	}
	m.release(ptr)
	if _, err := m.read(ptr, 1); !errors.Is(err, ErrRuntime) {
		t.Fatal("retired", err)
	}
	other, err := m.alloc(16, false)
	if err != nil || other == ptr {
		t.Fatal("region ID reuse", other, err)
	}
	if _, err := m.read(ptr, 1); !errors.Is(err, ErrRuntime) {
		t.Fatal("resurrected", err)
	}
	locked, err := m.alloc(1, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.write(locked, []byte{1}); !errors.Is(err, ErrRuntime) {
		t.Fatal("readonly", err)
	}
}

func TestVirtualMemoryLimitsAndPointerArithmetic(t *testing.T) {
	var m memory
	if _, err := m.alloc(maxMemoryBytes+1, false); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	ptr, _ := m.alloc(1, false)
	for _, delta := range []int64{-1, math.MaxUint32 + 1, math.MaxInt64, math.MinInt64} {
		if _, err := pointerAdd(ptr, delta); !errors.Is(err, ErrRuntime) {
			t.Fatal(delta, err)
		}
	}
	if next, err := pointerAdd(ptr+8, -4); err != nil || next != ptr+4 {
		t.Fatal(next, err)
	}
	if delta, err := pointerDifference(ptr, ptr+4); err != nil || delta != math.MaxUint32-3 {
		t.Fatal(delta, err)
	}
	if _, err := pointerDifference(ptr, ptr+(1<<32)); !errors.Is(err, ErrRuntime) {
		t.Fatal("different regions", err)
	}
	if _, err := pointerDifference(ptr+math.MaxUint32, ptr); !errors.Is(err, ErrRuntime) {
		t.Fatal("overflow", err)
	}
}

func FuzzVirtualMemory(f *testing.F) {
	f.Add(uint64(0), uint16(0), uint16(1), byte(42))
	f.Add(uint64(1)<<32, uint16(127), uint16(2), byte(0))
	f.Fuzz(func(t *testing.T, ptr uint64, offset, length uint16, value byte) {
		var m memory
		base, err := m.alloc(128, false)
		if err != nil {
			t.Fatal(err)
		}
		_ = m.fill(base+uint64(offset), value, int(length))
		_, _ = m.read(ptr, int(length))
		_, _ = pointerAdd(ptr, int64(offset)-32768)
		_, _ = pointerDifference(ptr, base)
		m.release(base)
		if _, err := m.read(base, 0); !errors.Is(err, ErrRuntime) {
			t.Fatal(err)
		}
	})
}

func TestVirtualMemoryReusesRetiredRegionsWithoutResurrection(t *testing.T) {
	// Live memory, not allocation history, is bounded. A retired region's
	// slot can be reused, but every pointer to the retired region stays
	// invalid because each reuse carries a new generation.
	var m memory
	global, err := m.alloc(8, false)
	if err != nil || m.store(global, 1, 0x5a) != nil {
		t.Fatal(err)
	}
	first, err := m.alloc(64, false)
	if err != nil {
		t.Fatal(err)
	}
	m.release(first)
	seen := map[uint64]bool{global >> 32: true, first >> 32: true}
	for i := 0; i < maxMemoryRegions+5000; i++ {
		ptr, err := m.alloc(64, false)
		if err != nil {
			t.Fatalf("allocation %d after retirement: %v", i, err)
		}
		if seen[ptr>>32] {
			t.Fatalf("allocation %d reused pointer identity %#x", i, ptr)
		}
		seen[ptr>>32] = true
		if err := m.store(ptr, 1, 1); err != nil {
			t.Fatal(err)
		}
		m.release(ptr)
	}
	if _, err := m.read(first, 0); !errors.Is(err, ErrRuntime) {
		t.Fatal("retired region resurrected", err)
	}
	if len(m.regions) > 4 || m.used != 8 {
		t.Fatalf("retired regions retained: %d slots, %d live bytes", len(m.regions), m.used)
	}
	if value, err := m.load(global, 1); err != nil || value != 0x5a {
		t.Fatal("live region disturbed by churn", value, err)
	}
	// A slot whose generation space is exhausted is retired for good rather
	// than wrapping to a generation that stale pointers could still hold.
	var e memory
	a, err := e.alloc(1, false)
	if err != nil {
		t.Fatal(err)
	}
	e.release(a)
	slot := a >> 32 & 0xffff
	e.generations[slot] = maxRegionGeneration - 1
	last, err := e.alloc(1, false)
	if err != nil || last>>32&0xffff != slot || last>>48 != maxRegionGeneration {
		t.Fatalf("final generation: %#x, %v", last, err)
	}
	e.release(last)
	next, err := e.alloc(1, false)
	if err != nil || next>>32&0xffff == slot {
		t.Fatalf("exhausted slot reused: %#x, %v", next, err)
	}
	for _, stale := range []uint64{a, last} {
		if _, err := e.read(stale, 0); !errors.Is(err, ErrRuntime) {
			t.Fatalf("stale pointer %#x resolved: %v", stale, err)
		}
	}
	// The live-byte bound is unchanged.
	if _, err := e.alloc(maxMemoryBytes, false); !errors.Is(err, ErrLimit) {
		t.Fatal("live memory bound", err)
	}
}
