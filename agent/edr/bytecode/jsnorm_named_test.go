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

func namedJSNormalizerProgram() memoryFixture {
	jump := func(block uint64) string { return "T" + compatFixed(18, 2) + compatNum(block) }
	branch := func(condition, yes, no uint64) string {
		return "T" + compatFixed(17, 2) + compatNum(condition) + compatNum(yes) + compatNum(no)
	}
	return memoryFixture{types: []string{memoryType(1, 2, 32, 32)}, apiMax: 45,
		apis: []string{memoryAPI(31, 69, "buffer_pipe_new_fromfile"), memoryAPI(32, 69, "buffer_pipe_read_avail"), memoryAPI(43, 69, "jsnorm_init"), memoryAPI(44, 69, "jsnorm_process"), memoryAPI(45, 69, "jsnorm_done")},
		functions: []compatFunction{{locals: []int{32, 32, 32, 32, 32, 1}, result: 32, blocks: [][]string{
			{memoryAPICall(0, 31, compatConstant(2, 32)), memoryAPICall(1, 43, compatNum(0)), jump(1)},
			{memoryAPICall(4, 32, compatNum(0)), compatInst(1, 5, 21, compatNum(32), compatNum(4), compatConstant(0, 32)), branch(5, 3, 2)},
			{memoryAPICall(2, 44, compatNum(1)), jump(1)},
			{memoryAPICall(3, 45, compatNum(1)), compatRet(32, compatNum(3))},
		}}}}
}

func TestAPIJSNormalizerRepeatedNamedCalls(t *testing.T) {
	const script = `function pack(p,a,c,k,e,r){return p} pack('HARMLESS',2,2,''.split('|'),0,{});pack('CHILD;',2,2,''.split('|'),0,{});`
	for _, pad := range []int{0, 997, 1010, 1023, 2048} {
		t.Run(fmt.Sprint(pad), func(t *testing.T) {
			input := append([]byte("AB"), bytes.Repeat([]byte(" "), pad)...)
			input = append(input, script...)
			p := memoryParse(t, namedJSNormalizerProgram())
			seen := 0
			result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, input, EmptyMatchContext(), func(child []byte, id int32) (string, error) {
				seen++
				if id != -1 || !bytes.Contains(child, []byte("harmlesschild")) {
					t.Fatalf("normalized child=%q id=%d", child, id)
				}
				return "Decoded.Named.API", nil
			}, 100000)
			if err != nil || result.ChildName != "Decoded.Named.API" || seen != 1 {
				t.Fatalf("result=%+v err=%v scans=%d", result, err, seen)
			}
		})
	}
}

func TestWriteNamedJSNormalizerOracleFixtures(t *testing.T) {
	dir := os.Getenv("EDR_JS_NAMED_EXPORT")
	if dir == "" {
		t.Skip("set output path for harmless complete-scanner controls")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m := namedJSNormalizerProgram()
	m.kind = 256
	m.logical = "UTM.NamedJS;Engine:56-255,Target:0;0;4142"
	files := map[string][]byte{"program.cbc": m.encode(), "child.ndb": []byte("Decoded.Named.API:0:*:" + hex.EncodeToString([]byte("harmlesschild")) + "\n")}
	for _, pad := range []int{0, 997, 1010, 1023, 2048} {
		for _, tc := range []struct{ prefix, tail, label string }{{"AB", "CHILD;", "positive"}, {"AB", "SAFE;", "negative"}, {"AC", "CHILD;", "disabled"}} {
			script := `function pack(p,a,c,k,e,r){return p} pack('HARMLESS',2,2,''.split('|'),0,{});pack('` + tc.tail + `',2,2,''.split('|'),0,{});`
			data := append([]byte(tc.prefix), bytes.Repeat([]byte(" "), pad)...)
			data = append(data, script...)
			files[fmt.Sprintf("%04d-%s.bin", pad, tc.label)] = data
		}
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
