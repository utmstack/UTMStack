package bytecode

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"testing"
)

func TestJSONAPIHandlesAndValues(t *testing.T) {
	state := &apiState{scanning: true, jsonRoot: map[string]any{
		"WordDocument": map[string]any{"Attributes": map[string]any{"version": "1"}},
		"BinaryData":   map[string]any{"Count": int32(2)},
		"Values":       []any{"Alice", true, int32(12), nil},
		"Null":         nil,
	}}
	var mem memory
	name := func(s string) uint64 { return apiTestBytes(t, &mem, []byte(s)) }
	apiTestCall(t, state, nil, 92, 1)
	apiTestCall(t, state, &mem, 93, 0, name("Missing"), 7, 0)
	word := name("WordDocument")
	apiTestCall(t, state, &mem, 93, 1, word, 12, 0)
	apiTestCall(t, state, nil, 94, 4, 1)
	apiTestCall(t, state, nil, 99, 0, 1) // json-c treats objects as false
	apiTestCall(t, state, &mem, 93, 2, name("Attributes"), 10, 1)
	apiTestCall(t, state, &mem, 93, 3, name("version"), 7, 2)
	apiTestCall(t, state, nil, 94, 6, 3)
	apiTestCall(t, state, nil, 97, 1, 3)
	output, err := mem.alloc(4, false)
	if err != nil {
		t.Fatal(err)
	}
	apiTestCall(t, state, &mem, 98, 2, output, 4, 3)
	got, err := mem.read(output, 2)
	if err != nil || string(got) != "1\x00" {
		t.Fatalf("JSON string bytes: %q, %v", got, err)
	}
	apiTestCall(t, state, &mem, 93, 4, name("Values"), 6, 0)
	apiTestCall(t, state, nil, 95, 4, 4)
	apiTestCall(t, state, nil, 99, 0, 4) // even a nonempty array is false
	apiTestCall(t, state, nil, 96, 5, 0, 4)
	apiTestCall(t, state, nil, 94, 6, 5)
	apiTestCall(t, state, nil, 96, 6, 1, 4)
	apiTestCall(t, state, nil, 99, 1, 6)
	apiTestCall(t, state, nil, 96, 7, 2, 4)
	apiTestCall(t, state, nil, 100, 12, 7)
	apiTestCall(t, state, nil, 96, 0, 3, 4) // null array item
	apiTestCall(t, state, nil, 96, 0, 4, 4) // outside array
	apiTestCall(t, state, nil, 95, jsonSigned(-2), 1)
	apiTestCall(t, state, nil, 94, apiFailure, 999)
	apiTestCall(t, state, &mem, 93, apiFailure, word, 12, 999)
	apiTestCall(t, state, &mem, 93, 8, name("BinaryData"), 10, 0)
	apiTestCall(t, state, &mem, 93, 9, name("Count"), 5, 8)
	apiTestCall(t, state, nil, 100, 2, 9)
	apiTestCall(t, state, &mem, 93, 10, name("Null"), 4, 0)
	apiTestCall(t, state, nil, 94, 0, 10)
	apiTestCall(t, state, &mem, 93, apiFailure, name("Missing"), 7, 10)
}

func TestJSONAPIUnavailableAndBounds(t *testing.T) {
	apiTestCall(t, &apiState{}, nil, 92, 0)
	apiTestCall(t, &apiState{scanning: true}, nil, 92, 0)
	apiTestCall(t, &apiState{scanning: true}, nil, 94, apiFailure, 0)
	for id := 93; id <= 100; id++ {
		got, err := (&apiState{scanning: true}).invokeJSON(id, nil, nil, 0, func(uint64) error { return nil })
		if err != nil || got != apiFailure {
			t.Fatalf("inactive JSON API %d = %d, %v; want -1", id, got, err)
		}
	}
	state := &apiState{jsonRoot: map[string]any{"S": "abcdef", "Max": int64(math.MaxInt64)}}
	var mem memory
	key := apiTestBytes(t, &mem, []byte("S"))
	missing := apiTestBytes(t, &mem, []byte("Unknown"))
	state.jsonPartial = true
	if _, err := state.invoke(context.Background(), 93, []uint64{missing, 7, 0}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("absent partial metadata: %v", err)
	}
	apiTestCall(t, state, &mem, 93, 1, key, 1, 0)
	output, err := mem.alloc(3, false)
	if err != nil {
		t.Fatal(err)
	}
	apiTestCall(t, state, &mem, 98, 3, output, 3, 1)
	got, err := mem.read(output, 3)
	if err != nil || string(got) != "ab\x00" {
		t.Fatalf("truncated JSON string: %q, %v", got, err)
	}
	if _, err := state.invoke(context.Background(), 98, []uint64{output, 0, 1}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
		t.Fatalf("zero output buffer: %v", err)
	}
	state.jsonObjs = make([]any, maxJSONHandles)
	state.jsonObjs[0] = state.jsonRoot
	apiTestCall(t, state, &mem, 93, apiFailure, key, 1, 0)
}

func TestJSONAPIVMDispatch(t *testing.T) {
	fixture := memoryFixture{
		types:  []string{memoryType(1, 1, 32)},
		apiMax: 92, apis: []string{memoryAPI(92, 69, "json_is_active")},
		functions: []compatFunction{{locals: []int{32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 92), compatRet(32, compatNum(0)),
		}}}},
	}
	program := memoryParse(t, fixture)
	if got, err := program.Run(context.Background(), 0, nil, 1000); err != nil || got != 0 {
		t.Fatalf("unbound metadata = %d, %v", got, err)
	}
	matches := EmptyMatchContext()
	matches.Metadata = map[string]any{}
	if got, err := program.RunWithMatchContext(context.Background(), 0, nil, nil, matches, 1000); err != nil || got.Value != 1 {
		t.Fatalf("bound metadata = %d, %v", got.Value, err)
	}
	matches.Metadata = nil
	if got, err := program.RunWithMatchContext(context.Background(), 0, nil, nil, matches, 1000); err != nil || got.Value != 0 {
		t.Fatalf("missing scanner metadata = %d, %v", got.Value, err)
	}
}

func TestJSONCoercionReference(t *testing.T) {
	for _, tc := range []struct {
		value    any
		wantInt  int32
		wantBool bool
	}{
		{map[string]any{"x": 1}, 0, false},
		{[]any{1}, 0, false},
		{"12trailing", 12, true},
		{"  -8x", -8, true},
		{"unparseable", 0, true},
		{json.Number("0.5"), 0, true},
	} {
		i, known := jsonInteger(tc.value)
		if !known || i != tc.wantInt {
			t.Fatalf("int(%v) = %d, %t; want %d", tc.value, i, known, tc.wantInt)
		}
		b, known := jsonBoolean(tc.value)
		if !known || b != tc.wantBool {
			t.Fatalf("bool(%v) = %t, %t; want %t", tc.value, b, known, tc.wantBool)
		}
	}
}
