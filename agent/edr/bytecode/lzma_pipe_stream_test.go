package bytecode

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func lzmaStreamingProgram(window int) memoryFixture {
	m := bzip2StreamingProgram(window)
	for i, line := range m.apis {
		switch line {
		case memoryAPI(105, 70, "bzip2_init"):
			m.apis[i] = memoryAPI(102, 70, "lzma_init")
		case memoryAPI(106, 69, "bzip2_process"):
			m.apis[i] = memoryAPI(103, 69, "lzma_process")
		case memoryAPI(107, 69, "bzip2_done"):
			m.apis[i] = memoryAPI(104, 69, "lzma_done")
		}
	}
	for i := range m.functions[0].blocks {
		for j, line := range m.functions[0].blocks[i] {
			switch line {
			case memoryAPICall(2, 105, compatNum(0), compatNum(1)):
				m.functions[0].blocks[i][j] = memoryAPICall(2, 102, compatNum(0), compatNum(1))
			case memoryAPICall(3, 106, compatNum(2)):
				m.functions[0].blocks[i][j] = memoryAPICall(3, 103, compatNum(2))
			case memoryAPICall(7, 107, compatNum(2)):
				m.functions[0].blocks[i][j] = memoryAPICall(7, 104, compatNum(2))
			}
		}
	}
	return m
}

func TestAPILZMASerializedOutputWindows(t *testing.T) {
	for _, window := range []int{1, 3, 5, 13, 64} {
		for _, known := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/known-%t", window, known), func(t *testing.T) {
				p := memoryParse(t, lzmaStreamingProgram(window))
				scanned := 0
				result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, append([]byte("AB"), lzmaTestPackedConfig(t, known)...), EmptyMatchContext(), func(child []byte, id int32) (string, error) {
					scanned++
					if id != -1 || !bytes.Equal(child, []byte(lzmaTestChild)) {
						t.Fatalf("child=%q id=%d", child, id)
					}
					return "Child.LZMA", nil
				}, 100000)
				if err != nil || result.Value != 2 || result.ChildName != "Child.LZMA" || scanned != 1 {
					t.Fatalf("result=%+v err=%v scans=%d", result, err, scanned)
				}
			})
		}
	}
}

func TestAPILZMAErrorClosesAndReleases(t *testing.T) {
	packed := lzmaTestPacked(t)
	packed[13] = 1 // invalid range initializer
	var mem memory
	s := &apiState{input: packed}
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 30, 1, 64)
	apiTestCall(t, s, &mem, 102, 0, 0, 1)
	if s.lzmaDict == 0 || s.lzmaModels == 0 {
		t.Fatal("missing decoder reservations")
	}
	apiTestCall(t, s, &mem, 103, 1, 0)
	if s.pipes[0].read != 18 || s.pipes[1].write != 0 || s.lzmaDict != 0 || s.lzmaModels != 0 {
		t.Fatalf("error cursors and cleanup: in=%d out=%d dict=%d models=%d", s.pipes[0].read, s.pipes[1].write, s.lzmaDict, s.lzmaModels)
	}
	apiTestCall(t, s, &mem, 103, apiFailure, 0)
	apiTestCall(t, s, &mem, 104, apiFailure, 0)
}

func TestAPILZMASharedStorageCeilings(t *testing.T) {
	for _, models := range []bool{false, true} {
		var mem memory
		s := &apiState{input: lzmaTestPacked(t)}
		if models {
			s.lzmaModels = maxLZMADict
		} else {
			s.lzmaDict = maxLZMADict
		}
		apiTestCall(t, s, &mem, 31, 0, 0)
		apiTestCall(t, s, &mem, 30, 1, 64)
		_, err := s.invoke(context.Background(), 102, []uint64{0, 1}, &mem, func(uint64) error { return nil })
		if !errors.Is(err, ErrLimit) || len(s.lzmas) != 0 || s.pipes[0].read != 0 {
			t.Fatalf("reservation failure: %v", err)
		}
	}
}

func TestWriteLZMAStreamingOracleFixtures(t *testing.T) {
	dir := os.Getenv("EDR_LZMA_STREAM_EXPORT")
	if dir == "" {
		t.Skip("set export path for inert full-scanner controls")
	}
	for _, window := range []int{1, 5, 13, 64} {
		out := filepath.Join(dir, fmt.Sprint(window))
		if err := os.MkdirAll(out, 0700); err != nil {
			t.Fatal(err)
		}
		m := lzmaStreamingProgram(window)
		m.kind = 256
		m.logical = "UTM.LZMA.Stream;Engine:56-255,Target:0;0;4142"
		files := map[string][]byte{"program.cbc": m.encode(), "child.ndb": []byte("Child.LZMA:0:*:" + hex.EncodeToString([]byte(lzmaTestChild)) + "\n")}
		for _, known := range []bool{false, true} {
			for _, prefix := range []string{"AB", "AC"} {
				name := fmt.Sprintf("sample-%s-known-%t.bin", prefix, known)
				files[name] = append([]byte(prefix), lzmaTestPackedConfig(t, known)...)
			}
		}
		for name, data := range files {
			if err := os.WriteFile(filepath.Join(out, name), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
}
