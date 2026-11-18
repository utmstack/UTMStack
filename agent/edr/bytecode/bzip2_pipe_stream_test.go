package bytecode

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func bzip2StreamingProgram(window int) memoryFixture {
	m := bzip2SerializedProgram()
	m.apis = append(m.apis, memoryAPI(34, 70, "buffer_pipe_read_stopped"))
	jump := func(block uint64) string { return "T" + compatFixed(18, 2) + compatNum(block) }
	branch := func(condition, yes, no uint64) string {
		return "T" + compatFixed(17, 2) + compatNum(condition) + compatNum(yes) + compatNum(no)
	}
	m.functions[0].blocks = [][]string{
		{memoryAPICall(0, 31, compatConstant(2, 32)), memoryAPICall(1, 30, compatConstant(uint64(window), 32)), memoryAPICall(2, 105, compatNum(0), compatNum(1)), jump(1)},
		{memoryAPICall(3, 106, compatNum(2)), memoryAPICall(4, 32, compatNum(1)), compatInst(1, 8, 21, compatNum(32), compatNum(4), compatConstant(0, 32)), branch(8, 3, 2)},
		{compatInst(65, 5, 33, compatFixed(2, 1), compatNum(33), compatNum(1), compatNum(4)), memoryAPICall(6, 3, compatNum(5), compatNum(4)), memoryAPICall(9, 34, compatNum(1), compatNum(4)), jump(3)},
		{compatInst(1, 10, 21, compatNum(32), compatNum(3), compatConstant(0, 32)), branch(10, 1, 4)},
		{memoryAPICall(7, 107, compatNum(2)), compatRet(32, compatNum(3))},
	}
	return m
}

func TestAPIBzip2SerializedOutputWindows(t *testing.T) {
	for _, window := range []int{1, 3, 5, 13, 64} {
		t.Run(fmt.Sprint(window), func(t *testing.T) {
			p := memoryParse(t, bzip2StreamingProgram(window))
			scanned := 0
			result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, append([]byte("AB"), bzip2TestPacked(t)...), EmptyMatchContext(), func(child []byte, id int32) (string, error) {
				scanned++
				if id != -1 || !bytes.Equal(child, []byte(bzip2TestChild)) {
					t.Fatalf("child=%q id=%d", child, id)
				}
				return "Child.Bzip2", nil
			}, 100000)
			if err != nil || result.Value != 4 || result.ChildName != "Child.Bzip2" || scanned != 1 {
				t.Fatalf("result=%+v err=%v scans=%d", result, err, scanned)
			}
		})
	}
}

func TestWriteBzip2StreamingOracleFixtures(t *testing.T) {
	dir := os.Getenv("EDR_BZIP2_STREAM_EXPORT")
	if dir == "" {
		t.Skip("set EDR_BZIP2_STREAM_EXPORT to export inert streaming ClamBC controls")
	}
	for _, window := range []int{1, 5, 13, 64} {
		out := filepath.Join(dir, fmt.Sprint(window))
		if err := os.MkdirAll(out, 0700); err != nil {
			t.Fatal(err)
		}
		m := bzip2StreamingProgram(window)
		m.kind = 256
		m.logical = "UTM.Bzip2.Stream;Engine:56-255,Target:0;0;4142"
		for name, data := range map[string][]byte{
			"program.cbc":   m.encode(),
			"child.ndb":     []byte("Child.Bzip2:0:*:" + hex.EncodeToString([]byte(bzip2TestChild)) + "\n"),
			"sample-ab.bin": append([]byte("AB"), bzip2TestPacked(t)...),
			"sample-ac.bin": append([]byte("AC"), bzip2TestPacked(t)...),
		} {
			if err := os.WriteFile(filepath.Join(out, name), data, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
}
