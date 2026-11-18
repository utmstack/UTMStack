package bytecode

import (
	"encoding/binary"
	"math"
)

const maxMemoryBytes = 32 << 20

// maxMemoryRegions keeps slot 0xffff free: unregisteredRegion names it, so a
// pointer composed against it is never valid.
const maxMemoryRegions = 0xffff
const unregisteredRegion = uint64(0xffff) << 32
const maxRegionGeneration = 0xffff

// A virtual pointer is generation<<48 | slot<<32 | offset, where slot names a
// region table entry (zero is null) and generation its current occupant. It
// never contains a Go address. Releasing a region lets a later allocation
// reuse the slot with the next generation, so pointers into a retired stack
// frame stay invalid even after another frame occupies the slot. A slot whose
// generations are exhausted is never reused. Only live bytes and live slots
// are bounded: the work budget already charges every allocated frame byte.
type memoryRegion struct {
	data        []byte
	initialized []uint64
	readOnly    bool
	borrowed    bool
}

// borrowReadOnly registers an already-owned input snapshot without copying it
// or charging the VM allocation pool a second time. The caller retains the
// snapshot for the lifetime of this memory table.
func (m *memory) borrowReadOnly(data []byte) (uint64, error) {
	if len(data) == 0 {
		return 0, ErrRuntime
	}
	if len(m.regions) == 0 {
		m.regions, m.generations = append(m.regions, nil), append(m.generations, 0)
	}
	slot := len(m.regions)
	if last := len(m.free) - 1; last >= 0 {
		slot, m.free = m.free[last], m.free[:last]
		m.generations[slot]++
	} else {
		if slot >= maxMemoryRegions {
			return 0, ErrLimit
		}
		m.regions, m.generations = append(m.regions, nil), append(m.generations, 0)
	}
	initialized := make([]uint64, (len(data)+63)/64)
	for i := range initialized {
		initialized[i] = ^uint64(0)
	}
	m.regions[slot] = &memoryRegion{data: data, initialized: initialized, readOnly: true, borrowed: true}
	return (uint64(m.generations[slot])<<16 | uint64(slot)) << 32, nil
}

type memory struct {
	regions     []*memoryRegion
	generations []uint16
	free        []int
	used        int
}

func (m *memory) alloc(n int, readOnly bool) (uint64, error) {
	if n < 0 || n > maxMemoryBytes-m.used {
		return 0, ErrLimit
	}
	if len(m.regions) == 0 {
		m.regions, m.generations = append(m.regions, nil), append(m.generations, 0)
	}
	slot := len(m.regions)
	if last := len(m.free) - 1; last >= 0 {
		slot, m.free = m.free[last], m.free[:last]
		m.generations[slot]++
	} else {
		if slot >= maxMemoryRegions {
			return 0, ErrLimit
		}
		m.regions, m.generations = append(m.regions, nil), append(m.generations, 0)
	}
	m.regions[slot] = &memoryRegion{data: make([]byte, n), initialized: make([]uint64, (n+63)/64), readOnly: readOnly}
	m.used += n
	return (uint64(m.generations[slot])<<16 | uint64(slot)) << 32, nil
}

// replace grows an owned region while retiring all pointers into its old
// backing block. Capacity is checked against final live bytes; copying may
// temporarily hold both blocks, just as a moving realloc may.
func (m *memory) replace(ptr uint64, n int) (uint64, error) {
	oldSlot, ok := m.slot(ptr)
	if !ok || m.regions[oldSlot].borrowed || m.regions[oldSlot].readOnly {
		return 0, ErrRuntime
	}
	old := m.regions[oldSlot]
	if n < 0 || n > maxMemoryBytes-(m.used-len(old.data)) {
		return 0, ErrLimit
	}
	region := &memoryRegion{data: make([]byte, n), initialized: make([]uint64, (n+63)/64)}
	copy(region.data, old.data)
	copy(region.initialized, old.initialized)
	slot := len(m.regions)
	if last := len(m.free) - 1; last >= 0 {
		slot, m.free = m.free[last], m.free[:last]
		m.generations[slot]++
	} else {
		if slot >= maxMemoryRegions {
			return 0, ErrLimit
		}
		m.regions, m.generations = append(m.regions, nil), append(m.generations, 0)
	}
	m.regions[slot] = region
	m.used += n
	m.release(ptr)
	return (uint64(m.generations[slot])<<16 | uint64(slot)) << 32, nil
}

// slot resolves a pointer to the table entry of the live region it names.
func (m *memory) slot(ptr uint64) (int, bool) {
	slot, generation := int(ptr>>32&0xffff), uint16(ptr>>48)
	if slot == 0 || slot >= len(m.regions) || m.regions[slot] == nil || m.generations[slot] != generation {
		return 0, false
	}
	return slot, true
}

func (m *memory) release(ptr uint64) {
	slot, ok := m.slot(ptr)
	if !ok {
		return
	}
	if !m.regions[slot].borrowed {
		m.used -= len(m.regions[slot].data)
	}
	m.regions[slot] = nil
	if m.generations[slot] < maxRegionGeneration {
		m.free = append(m.free, slot)
	}
}

// protect makes an initialized live region read-only.
func (m *memory) protect(ptr uint64) {
	if slot, ok := m.slot(ptr); ok {
		m.regions[slot].readOnly = true
	}
}

func (m *memory) span(ptr uint64, n int, write bool) (*memoryRegion, int, error) {
	slot, ok := m.slot(ptr)
	if n < 0 || !ok {
		return nil, 0, ErrRuntime
	}
	r, off := m.regions[slot], uint64(uint32(ptr))
	if off > uint64(len(r.data)) || uint64(n) > uint64(len(r.data))-off || write && r.readOnly {
		return nil, 0, ErrRuntime
	}
	return r, int(off), nil
}

func (m *memory) check(ptr uint64, n int, write bool) error {
	_, _, err := m.span(ptr, n, write)
	return err
}

func (m *memory) read(ptr uint64, n int) ([]byte, error) {
	r, off, err := m.span(ptr, n, false)
	if err != nil {
		return nil, err
	}
	for i := off; i < off+n; i++ {
		if r.initialized[i/64]&(uint64(1)<<uint(i%64)) == 0 {
			return nil, ErrRuntime
		}
	}
	return r.data[off : off+n], nil
}

func (m *memory) write(ptr uint64, data []byte) error {
	r, off, err := m.span(ptr, len(data), true)
	if err != nil {
		return err
	}
	copy(r.data[off:], data)
	r.mark(off, len(data))
	return nil
}

func (r *memoryRegion) mark(off, n int) {
	for i := off; i < off+n; i++ {
		r.initialized[i/64] |= uint64(1) << uint(i%64)
	}
}

func (m *memory) fill(ptr uint64, b byte, n int) error {
	r, off, err := m.span(ptr, n, true)
	if err != nil {
		return err
	}
	for i := off; i < off+n; i++ {
		r.data[i] = b
	}
	r.mark(off, n)
	return nil
}

func (m *memory) load(ptr uint64, n int) (uint64, error) {
	b, err := m.read(ptr, n)
	if err != nil {
		return 0, err
	}
	switch n {
	case 1:
		return uint64(b[0]), nil
	case 2:
		return uint64(binary.LittleEndian.Uint16(b)), nil
	case 4:
		return uint64(binary.LittleEndian.Uint32(b)), nil
	case 8:
		return binary.LittleEndian.Uint64(b), nil
	default:
		return 0, ErrRuntime
	}
}
func (m *memory) store(ptr uint64, n int, v uint64) error {
	if n != 1 && n != 2 && n != 4 && n != 8 {
		return ErrRuntime
	}
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	return m.write(ptr, b[:n])
}

func pointerAdd(ptr uint64, delta int64) (uint64, error) {
	if ptr>>32 == 0 {
		return 0, ErrRuntime
	}
	off := int64(uint32(ptr))
	if delta < -off || delta > math.MaxUint32-off {
		return 0, ErrRuntime
	}
	return ptr&0xffffffff00000000 | uint64(off+delta), nil
}

func pointerDifference(a, b uint64) (uint64, error) {
	if a>>32 == 0 || a>>32 != b>>32 {
		return 0, ErrRuntime
	}
	d := int64(uint32(a)) - int64(uint32(b))
	if d < math.MinInt32 || d > math.MaxInt32 {
		return 0, ErrRuntime
	}
	return uint64(uint32(int32(d))), nil
}

// referencePointer applies the reference interpreter's pointer check
// (bytecode_vm.c ptr_torealptr): a live object, an offset inside it (the end
// itself is refused even for zero bytes), and n bytes within it.
func (m *memory) referencePointer(ptr uint64, n uint32) error {
	slot, ok := m.slot(ptr)
	if !ok {
		return ErrRuntime
	}
	size, off := uint64(len(m.regions[slot].data)), uint64(uint32(ptr))
	if off >= size || uint64(n) > size || off+uint64(n) > size {
		return ErrRuntime
	}
	return nil
}
