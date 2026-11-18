package bytecode

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFileReliabilityHostAPI(t *testing.T) {
	apiTestCall(t, &apiState{}, nil, 91, 3)
	for _, value := range []uint32{0, 1, 2} {
		v := value
		apiTestCall(t, &apiState{scanning: true, fileReliability: &v}, nil, 91, uint64(v))
	}
	if _, err := (&apiState{scanning: true}).invoke(context.Background(), 91, nil, nil, func(uint64) error { return nil }); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("unproved scan reliability = %v; want incomplete", err)
	}
}

func TestWriteReliabilityOracleFixture(t *testing.T) {
	dir := os.Getenv("EDR_RELIABILITY_ORACLE_DIR")
	if dir == "" {
		t.Skip("set EDR_RELIABILITY_ORACLE_DIR for a harmless native ClamAV differential")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		value uint64
	}{
		{"normal", 0}, {"no-context", 3},
	} {
		label := "UTM.Rel0"
		if tc.value == 3 {
			label = "UTM.Rel3"
		}
		m := engineLevelProgram(91, "get_file_reliability", tc.value, label)
		m.kind = 256
		m.logical = label + ";Engine:56-255,Target:0;0;4142"
		if _, err := Parse(m.encode()); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, tc.name+".cbc"), m.encode(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for name, prefix := range map[string]string{"sample-ab.bin": "AB", "sample-ac.bin": "AC"} {
		data := append([]byte(prefix), bytes.Repeat([]byte{'Q'}, 1024)...)
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFileReliabilitySerialized(t *testing.T) {
	m := memoryFixture{
		types:  []string{memoryType(1, 1, 32)},
		apiMax: 91,
		apis:   []string{memoryAPI(91, 69, "get_file_reliability")},
		functions: []compatFunction{{locals: []int{32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 91), compatRet(32, compatNum(0)),
		}}}},
	}
	p := memoryParse(t, m)
	if got, err := p.Run(context.Background(), 0, nil, 1000); err != nil || got != 3 {
		t.Fatalf("without scan context = %d, %v", got, err)
	}
	for _, value := range []uint32{0, 1, 2} {
		v := value
		mc := EmptyMatchContext()
		mc.FileReliability = &v
		if got, err := p.RunWithMatchContext(context.Background(), 0, nil, nil, mc, 1000); err != nil || got.Value != uint64(v) {
			t.Fatalf("reliability %d = %+v, %v", v, got, err)
		}
	}
	if _, err := p.RunWithMatchContext(context.Background(), 0, nil, nil, EmptyMatchContext(), 1000); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("unproved scan reliability = %v; want incomplete", err)
	}
}
