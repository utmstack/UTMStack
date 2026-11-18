package bytecode

// bytecodeMap models the key table and the independently indexed value
// storage in ClamAV 1.5.4's cli_map. Deleted value slots remain addressable
// through a previous last_find; only map_done retires the whole object.
type bytecodeMap struct {
	keysize, valuesize int
	keys               map[string]int
	values             []mapValue
	lastInsert         int
	lastFind           int
	fixedPtr           uint64
}

type mapValue struct {
	ptr  uint64 // used only for variable-sized values
	size int
}

func (s *apiState) bytecodeMap(id uint64) (*bytecodeMap, bool) {
	i := int32(id)
	if i < 0 || int(i) >= len(s.maps) {
		return nil, false
	}
	if s.maps[i] == nil {
		return nil, true
	}
	return s.maps[i], true
}

func (s *apiState) mapAPI(id int, args []uint64, mem *memory, bytesWork func(uint64) error) (uint64, error) {
	if id == 57 {
		keysize, valuesize := int32(args[0]), int32(args[1])
		if keysize == 0 {
			return apiFailure, nil
		}
		// cli_map_init rejects these, but bytecode_api.c ignores that error
		// and leaves the allocated slot uninitialized. It has no safe result.
		if keysize < 0 || valuesize < 0 {
			return 0, ErrRuntime
		}
		if keysize > maxMapKeys || valuesize > maxMemoryBytes {
			return 0, ErrLimit
		}
		if len(s.maps) >= maxMaps {
			return 0, ErrLimit
		}
		n := len(s.maps)
		s.maps = append(s.maps, &bytecodeMap{keysize: int(keysize), valuesize: int(valuesize), keys: make(map[string]int), lastInsert: -1, lastFind: -1})
		return uint64(n), nil
	}
	mapID := args[0]
	if id == 58 || id == 59 || id == 60 || id == 61 {
		mapID = args[2]
	}
	m, present := s.bytecodeMap(mapID)
	if !present {
		if id == 63 {
			return 0, nil
		}
		return apiFailure, nil
	}
	if m == nil {
		return 0, ErrRuntime
	}
	if mem == nil && (id == 58 || id == 59 || id == 60 || id == 61 || id == 63 || id == 64) {
		return 0, ErrRuntime
	}
	switch id {
	case 58, 60, 61:
		n := int(uint32(args[1]))
		if n != m.keysize {
			return apiFailure, nil
		}
		if err := bytesWork(uint64(n)); err != nil {
			return 0, err
		}
		data, err := mem.read(args[0], n)
		if err != nil {
			return 0, err
		}
		key := string(data)
		index, found := m.keys[key]
		if id == 61 {
			if !found {
				return 0, nil // last_find is retained on a miss
			}
			m.lastFind = index
			return 1, nil
		}
		if id == 60 {
			if !found {
				return 0, nil
			}
			delete(m.keys, key)
			s.mapKeys -= n
			if m.valuesize != 0 {
				if err := mem.fill(m.fixedPtr+uint64(index*m.valuesize), 0, m.valuesize); err != nil {
					return 0, err
				}
			} else {
				if m.values[index].ptr != 0 {
					mem.release(m.values[index].ptr)
				}
				m.values[index] = mapValue{}
			}
			return 1, nil
		}
		if found {
			m.lastInsert = index
			return 0, nil
		}
		if s.mapSlots >= maxMapSlots || n > maxMapKeys-s.mapKeys {
			return 0, ErrLimit
		}
		if m.valuesize != 0 {
			if m.valuesize > maxMemoryBytes/(len(m.values)+1) {
				return 0, ErrLimit
			}
			newSize := (len(m.values) + 1) * m.valuesize
			if err := bytesWork(uint64(newSize)); err != nil {
				return 0, err
			}
			var ptr uint64
			var err error
			if m.fixedPtr == 0 {
				ptr, err = mem.alloc(newSize, false)
			} else {
				ptr, err = mem.replace(m.fixedPtr, newSize)
			}
			if err != nil {
				return 0, err
			}
			if err := mem.fill(ptr+uint64(len(m.values)*m.valuesize), 0, m.valuesize); err != nil {
				mem.release(ptr)
				return 0, err
			}
			m.fixedPtr = ptr
		}
		index = len(m.values)
		m.values = append(m.values, mapValue{})
		m.keys[key] = index
		m.lastInsert = index
		s.mapSlots++
		s.mapKeys += n
		return 1, nil
	case 59:
		n := int(uint32(args[1]))
		if m.lastInsert < 0 || m.lastInsert >= len(m.values) || m.valuesize != 0 && n != m.valuesize {
			return apiFailure, nil
		}
		if err := bytesWork(uint64(n)); err != nil {
			return 0, err
		}
		data, err := mem.read(args[0], n)
		if err != nil {
			return 0, err
		}
		if m.valuesize != 0 {
			return 0, mem.write(m.fixedPtr+uint64(m.lastInsert*m.valuesize), data)
		}
		if n == 0 {
			// cli_map_setvalue frees the old dynamic value before its
			// cli_max_malloc(0) fails. The stored size is not reset.
			if old := m.values[m.lastInsert].ptr; old != 0 {
				mem.release(old)
				m.values[m.lastInsert].ptr = 0
			}
			return apiFailure, nil
		}
		ptr, err := mem.alloc(n, false)
		if err != nil {
			return 0, err
		}
		if err := mem.write(ptr, data); err != nil {
			mem.release(ptr)
			return 0, err
		}
		if old := m.values[m.lastInsert].ptr; old != 0 {
			mem.release(old)
		}
		m.values[m.lastInsert] = mapValue{ptr: ptr, size: n}
		return 0, nil
	case 62:
		if m.valuesize != 0 {
			return uint64(m.valuesize), nil
		}
		if m.lastFind < 0 || m.lastFind >= len(m.values) {
			return apiFailure, nil
		}
		return uint64(m.values[m.lastFind].size), nil
	case 63:
		want := int(uint32(args[1]))
		if m.valuesize != 0 {
			if want != m.valuesize || m.lastFind < 0 || m.lastFind >= len(m.values) {
				return 0, nil
			}
			return m.fixedPtr + uint64(m.lastFind*m.valuesize), nil
		}
		if m.lastFind < 0 || m.lastFind >= len(m.values) || want != m.values[m.lastFind].size {
			return 0, nil
		}
		return m.values[m.lastFind].ptr, nil
	case 64:
		for _, value := range m.values {
			if value.ptr != 0 {
				mem.release(value.ptr)
			}
		}
		if m.fixedPtr != 0 {
			mem.release(m.fixedPtr)
		}
		s.mapSlots -= len(m.values)
		s.mapKeys -= len(m.keys) * m.keysize
		i := int(int32(mapID))
		s.maps[i] = nil
		if i == len(s.maps)-1 {
			s.maps = s.maps[:i]
		}
		return 0, nil
	default:
		return 0, ErrUnsupported
	}
}
