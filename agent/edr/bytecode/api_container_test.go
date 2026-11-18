package bytecode

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAPIExtractSetContainerBoundsAndChildLifetime(t *testing.T) {
	var mem memory
	var calls []struct {
		data string
		id   int32
		typ  uint32
	}
	s := &apiState{scanning: true, input: []byte("AB"), outputID: -1,
		scanExtractedTyped: func(data []byte, id int32, typ uint32) (string, error) {
			calls = append(calls, struct {
				data string
				id   int32
				typ  uint32
			}{string(data), id, typ})
			return "", nil
		}}
	data := apiTestBytes(t, &mem, []byte("CHILD-AB"))
	for _, typ := range []uint64{0, 500, 586, 587} {
		apiTestCall(t, s, &mem, 70, 0, typ)
		if s.containerType != uint32(typ) {
			t.Fatalf("container type after %#x = %d", typ, s.containerType)
		}
	}
	for _, typ := range []uint64{588, 0xffffffff} {
		apiTestCall(t, s, &mem, 70, apiFailure, typ)
		if s.containerType != 587 {
			t.Fatalf("invalid container type mutated state to %d", s.containerType)
		}
	}
	apiTestCall(t, s, &mem, 70, 0, 586)
	apiTestCall(t, s, &mem, 3, 8, data, 8)
	apiTestCall(t, s, &mem, 22, 0, 7)
	apiTestCall(t, s, &mem, 3, 8, data, 8)
	apiTestCall(t, s, &mem, 70, 0, 504)
	apiTestCall(t, s, &mem, 22, 0, 8)
	if len(calls) != 2 || calls[0].data != "CHILD-AB" || calls[0].id != -1 || calls[0].typ != 586 ||
		calls[1].data != "CHILD-AB" || calls[1].id != 7 || calls[1].typ != 504 {
		t.Fatalf("typed extracted children = %+v", calls)
	}
}

func TestAPIExtractSetContainerRequiresTypedChildScanner(t *testing.T) {
	var mem memory
	data := apiTestBytes(t, &mem, []byte("CHILD-AB"))
	s := &apiState{scanning: true, scanExtracted: func([]byte, int32) (string, error) { return "", nil }}
	apiTestCall(t, s, &mem, 70, 0, 586)
	apiTestCall(t, s, &mem, 3, 8, data, 8)
	if _, err := s.invoke(context.Background(), 22, []uint64{7}, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrExtraction) {
		t.Fatalf("untyped callback accepted forced container: %v", err)
	}
}

func containerOracleFixture(typ uint32, label string) memoryFixture {
	m := extractionFixture(compatRet(32, compatNum(1)))
	m.kind = 256
	m.logical = label + ";Engine:56-255,Target:0;0;4142"
	m.apiMax = 70
	m.apis = append(m.apis, memoryAPI(70, 72, "extract_set_container"))
	m.functions[0].locals = []int{32, 32, 32}
	m.functions[0].blocks[0] = []string{
		memoryAPICall(2, 70, compatConstant(uint64(typ), 32)),
		memoryAPICall(0, 3, memoryGlobalRef(2), compatConstant(uint64(len("HARMLESS-CHILD")), 32)),
		memoryAPICall(1, 22, compatConstant(7, 32)),
		compatRet(32, compatNum(1)),
	}
	return m
}

func TestSerializedExtractSetContainerRoutesType(t *testing.T) {
	m := containerOracleFixture(586, "UTM.Container.With")
	var called bool
	result, err := memoryParse(t, m).RunWithMatchContextAndTypedScanner(context.Background(), 0, nil,
		[]byte("AB"), EmptyMatchContext(), func(data []byte, id int32, typ uint32) (string, error) {
			called = true
			if id != -1 || typ != 586 || !bytes.Equal(data, []byte("HARMLESS-CHILD")) {
				t.Errorf("typed child: id %d, type %d, bytes %q", id, typ, data)
			}
			return "Child.Other", nil
		}, 1000)
	if err != nil || !called || result.ChildName != "Child.Other" || result.Value != 1 {
		t.Fatalf("serialized forced-type child = %+v, %v, called %v", result, err, called)
	}
	if _, err := memoryParse(t, m).RunWithMatchContextAndScanner(context.Background(), 0, nil,
		[]byte("AB"), EmptyMatchContext(), func([]byte, int32) (string, error) { return "", nil }, 1000); !errors.Is(err, ErrExtraction) {
		t.Fatalf("untyped VM accepted forced type: %v", err)
	}
}

func TestWriteContainerOracleFixture(t *testing.T) {
	dir := os.Getenv("EDR_CONTAINER_ORACLE_DIR")
	if dir == "" {
		t.Skip("set EDR_CONTAINER_ORACLE_DIR for the native ClamAV differential")
	}
	with := containerOracleFixture(586, "UTM.Container.With")
	invalid := containerOracleFixture(588, "UTM.Container.Invalid")
	gap := containerOracleFixture(499, "UTM.Container.Gap")
	without := extractionFixture(compatRet(32, compatNum(1)))
	without.kind = 256
	without.logical = "UTM.Container.Without;Engine:56-255,Target:0;0;4142"
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"with.cbc":      with.encode(),
		"invalid.cbc":   invalid.encode(),
		"gap.cbc":       gap.encode(),
		"without.cbc":   without.encode(),
		"child.ndb":     []byte("Child.Other:14:*:4841524d4c4553532d4348494c44\n"),
		"sample-ab.bin": append([]byte("AB"), bytes.Repeat([]byte{'Q'}, 1024)...),
		"sample-ac.bin": append([]byte("AC"), bytes.Repeat([]byte{'Q'}, 1024)...),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
