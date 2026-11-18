package bytecode

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAPIJSNormalizerCarriedState(t *testing.T) {
	for _, script := range []string{`eval("EHARMLESS");bridge();eval("CHILD;");`, `eval("1.e")eval("HARMLESSCHILD;");`} {
		p := memoryParse(t, namedJSNormalizerProgram())
		seen := 0
		result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, []byte("AB"+script), EmptyMatchContext(), func(data []byte, id int32) (string, error) {
			seen++
			if id != -1 || !bytes.Contains(data, []byte("eharmlesschild")) {
				t.Fatalf("child=%q id=%d", data, id)
			}
			return "Decoded.Carry.Join", nil
		}, 100000)
		if err != nil || seen != 1 || result.ChildName != "Decoded.Carry.Join" {
			t.Fatalf("result=%+v err=%v scans=%d", result, err, seen)
		}
	}
}
func TestWriteCarriedJSNormalizerOracleProgram(t *testing.T) {
	dir := os.Getenv("EDR_JS_CARRY_EXPORT")
	if dir == "" {
		t.Skip("set output path for inert native scanner controls")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m := namedJSNormalizerProgram()
	m.kind = 256
	m.logical = "UTM.CarriedJS;Engine:56-255,Target:0;0;4142"
	if err := os.WriteFile(filepath.Join(dir, "program.cbc"), m.encode(), 0600); err != nil {
		t.Fatal(err)
	}
}
