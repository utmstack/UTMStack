package bytecode

import (
	"context"
	"errors"
	"testing"
)

func TestAPIMapFixedValuesAndLifecycle(t *testing.T) {
	var mem memory
	s := &apiState{}
	aa, bb, cc := apiTestBytes(t, &mem, []byte("aa")), apiTestBytes(t, &mem, []byte("bb")), apiTestBytes(t, &mem, []byte("cc"))
	one, two := apiTestBytes(t, &mem, []byte("one")), apiTestBytes(t, &mem, []byte("two"))
	apiTestCall(t, s, &mem, 57, 0, 2, 3)
	apiTestCall(t, s, &mem, 62, 3, 0) // fixed-size maps answer before find
	apiTestCall(t, s, &mem, 58, 1, aa, 2, 0)
	apiTestCall(t, s, &mem, 58, 0, aa, 2, 0)
	apiTestCall(t, s, &mem, 59, 0, one, 3, 0)
	apiTestCall(t, s, &mem, 61, 1, aa, 2, 0)
	old, err := s.invoke(context.Background(), 63, []uint64{0, 3}, &mem, func(uint64) error { return nil })
	if err != nil || old == 0 {
		t.Fatalf("first value pointer = %#x, %v", old, err)
	}
	if got, err := mem.read(old, 3); err != nil || string(got) != "one" {
		t.Fatalf("first value = %q, %v", got, err)
	}
	apiTestCall(t, s, &mem, 58, 1, bb, 2, 0)
	if err := mem.check(old, 1, false); !errors.Is(err, ErrRuntime) {
		t.Fatalf("value pointer survived realloc: %v", err)
	}
	apiTestCall(t, s, &mem, 59, 0, two, 3, 0)
	apiTestCall(t, s, &mem, 61, 0, cc, 2, 0) // miss retains last_find
	first, err := s.invoke(context.Background(), 63, []uint64{0, 3}, &mem, func(uint64) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if got, err := mem.read(first, 3); err != nil || string(got) != "one" {
		t.Fatalf("last_find changed after miss: %q, %v", got, err)
	}
	apiTestCall(t, s, &mem, 61, 1, bb, 2, 0)
	second, err := s.invoke(context.Background(), 63, []uint64{0, 3}, &mem, func(uint64) error { return nil })
	if err != nil || second != first+3 {
		t.Fatalf("contiguous fixed value pointer = %#x, %v; first %#x", second, err, first)
	}
	if got, err := mem.read(second, 3); err != nil || string(got) != "two" {
		t.Fatalf("second value = %q, %v", got, err)
	}
	apiTestCall(t, s, &mem, 60, 1, bb, 2, 0)
	if got, err := mem.read(second, 3); err != nil || string(got) != "\x00\x00\x00" {
		t.Fatalf("removed value not zeroed: %q, %v", got, err)
	}
	apiTestCall(t, s, &mem, 60, 0, bb, 2, 0)
	apiTestCall(t, s, &mem, 64, 0, 0)
	if err := mem.check(first, 1, false); !errors.Is(err, ErrRuntime) {
		t.Fatalf("map value survived done: %v", err)
	}
	apiTestCall(t, s, &mem, 57, 0, 1, 1) // last ID is reused
}

func TestAPIMapVariableValuesAndInvalidCalls(t *testing.T) {
	var mem memory
	s := &apiState{}
	key, missing := apiTestBytes(t, &mem, []byte("A")), apiTestBytes(t, &mem, []byte("B"))
	first, second := apiTestBytes(t, &mem, []byte("X")), apiTestBytes(t, &mem, []byte("YZ"))
	apiTestCall(t, s, &mem, 57, apiFailure, 0, 0)
	apiTestCall(t, s, &mem, 57, 0, 1, 0)
	apiTestCall(t, s, &mem, 62, apiFailure, 0)
	apiTestCall(t, s, &mem, 58, 1, key, 1, 0)
	apiTestCall(t, s, &mem, 61, 1, key, 1, 0)
	apiTestCall(t, s, &mem, 62, 0, 0)
	apiTestCall(t, s, &mem, 63, 0, 0, 0)
	apiTestCall(t, s, &mem, 59, 0, first, 1, 0)
	apiTestCall(t, s, &mem, 62, 1, 0)
	ptr, err := s.invoke(context.Background(), 63, []uint64{0, 1}, &mem, func(uint64) error { return nil })
	if err != nil || ptr == 0 {
		t.Fatalf("dynamic value pointer = %#x, %v", ptr, err)
	}
	apiTestCall(t, s, &mem, 59, 0, second, 2, 0)
	if err := mem.check(ptr, 1, false); !errors.Is(err, ErrRuntime) {
		t.Fatalf("replaced dynamic value remained live: %v", err)
	}
	apiTestCall(t, s, &mem, 62, 2, 0)
	apiTestCall(t, s, &mem, 63, 0, 0, 1)
	current, err := s.invoke(context.Background(), 63, []uint64{0, 2}, &mem, func(uint64) error { return nil })
	if err != nil || current == 0 {
		t.Fatalf("replacement value pointer = %#x, %v", current, err)
	}
	apiTestCall(t, s, &mem, 59, apiFailure, first, 0, 0)
	if err := mem.check(current, 1, false); !errors.Is(err, ErrRuntime) {
		t.Fatalf("zero-sized failed update retained old value: %v", err)
	}
	apiTestCall(t, s, &mem, 62, 2, 0) // failure does not reset stored size
	apiTestCall(t, s, &mem, 63, 0, 0, 2)
	apiTestCall(t, s, &mem, 59, 0, second, 2, 0)
	apiTestCall(t, s, &mem, 61, 0, missing, 1, 0)
	apiTestCall(t, s, &mem, 62, 2, 0)
	apiTestCall(t, s, &mem, 60, 1, key, 1, 0)
	apiTestCall(t, s, &mem, 62, 0, 0)
	apiTestCall(t, s, &mem, 63, 0, 0, 0)
	apiTestCall(t, s, &mem, 58, apiFailure, second, 2, 0)
	apiTestCall(t, s, &mem, 58, apiFailure, key, 1, 7)
	apiTestCall(t, s, &mem, 64, 0, 0)
	apiTestCall(t, s, &mem, 62, apiFailure, 7)
	apiTestCall(t, s, &mem, 63, 0, 7, 1)
	if _, err := s.invoke(context.Background(), 57, []uint64{uint64(uint32(0xffffffff)), 1}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
		t.Fatalf("negative key size = %v", err)
	}
}

func TestAPIMapFixedGrowthUsesFinalLiveBudget(t *testing.T) {
	var mem memory
	s := &apiState{}
	first, second := apiTestBytes(t, &mem, []byte("A")), apiTestBytes(t, &mem, []byte("B"))
	const valueSize = 11 << 20
	apiTestCall(t, s, &mem, 57, 0, 1, valueSize)
	apiTestCall(t, s, &mem, 58, 1, first, 1, 0)
	old := s.maps[0].fixedPtr
	apiTestCall(t, s, &mem, 58, 1, second, 1, 0)
	if mem.used != 2*valueSize+2 || s.maps[0].fixedPtr == old || !errors.Is(mem.check(old, 1, false), ErrRuntime) {
		t.Fatalf("map realloc: used %d, old %#x, new %#x", mem.used, old, s.maps[0].fixedPtr)
	}
}

func TestAPIMapWireAndBoundedWork(t *testing.T) {
	types := []string{
		memoryType(1, 3, 32, 32, 32), memoryType(1, 4, 32, 65, 32, 32),
		memoryType(1, 2, 32, 32), memoryType(1, 3, 65, 32, 32),
	}
	apis := []string{
		memoryAPI(57, 69, "map_new"), memoryAPI(58, 70, "map_addkey"),
		memoryAPI(59, 70, "map_setvalue"), memoryAPI(60, 70, "map_remove"),
		memoryAPI(61, 70, "map_find"), memoryAPI(62, 71, "map_getvaluesize"),
		memoryAPI(63, 72, "map_getvalue"), memoryAPI(64, 71, "map_done"),
	}
	m := memoryFixture{types: types, apis: apis, apiMax: 64,
		functions: []compatFunction{{locals: []int{32, 32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 57, compatConstant(1, 32), compatConstant(7, 32)),
			memoryAPICall(1, 62, compatNum(0)), compatRet(32, compatNum(1)),
		}}}},
	}
	result, err := memoryParse(t, m).RunWithInput(context.Background(), 0, nil, nil, 1000)
	if err != nil || result.Value != 7 {
		t.Fatalf("serialized map API = %+v, %v", result, err)
	}
	var mem memory
	s := &apiState{}
	key := apiTestBytes(t, &mem, []byte("A"))
	apiTestCall(t, s, &mem, 57, 0, 1, 1)
	if _, err := s.invoke(context.Background(), 58, []uint64{key, 1, 0}, &mem, func(uint64) error { return ErrLimit }); !errors.Is(err, ErrLimit) || s.mapSlots != 0 {
		t.Fatalf("work limit = %v, slots %d", err, s.mapSlots)
	}
	s.maps = make([]*bytecodeMap, maxMaps)
	if _, err := s.invoke(context.Background(), 57, []uint64{1, 1}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrLimit) {
		t.Fatalf("map count limit = %v", err)
	}
}
