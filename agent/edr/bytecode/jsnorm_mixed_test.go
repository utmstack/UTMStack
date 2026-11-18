package bytecode

import (
	"bytes"
	"context"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestAPIJSNormalizerMixedOrder(t *testing.T) {
	for _, script := range []string{
		`eval("pre();");dF("%48%41%52%4d%4c%45%53%53%43%48%49%4c%44%3b0");`,
		`dF("%48%41%52%4d%4c%45%53%53%43%48%49%4c%44%3b0");eval("ignored();");`,
		`eval("pre();dF('%48%41%52%4d%4c%45%53%53%43%48%49%4c%44%3b0');");`,
	} {
		p := memoryParse(t, namedJSNormalizerProgram())
		seen := 0
		result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, []byte("AB"+script), EmptyMatchContext(), func(data []byte, id int32) (string, error) {
			seen++
			if id != -1 || !bytes.Contains(data, []byte("harmlesschild")) {
				t.Fatalf("child=%q id=%d", data, id)
			}
			return "Decoded.Mixed.API", nil
		}, 100000)
		if err != nil || result.ChildName != "Decoded.Mixed.API" || seen != 1 {
			t.Fatalf("result=%+v err=%v scans=%d", result, err, seen)
		}
	}
}

func TestWriteMixedJSNormalizerOracleProgram(t *testing.T) {
	dir := os.Getenv("EDR_JS_MIXED_EXPORT")
	if dir == "" {
		t.Skip("set output path for inert native scanner controls")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	m := namedJSNormalizerProgram()
	m.kind = 256
	m.logical = "UTM.MixedJS;Engine:56-255,Target:0;0;4142"
	for name, data := range map[string][]byte{"program.cbc": m.encode(), "child.ndb": []byte("Decoded.Mixed.API:0:*:" + hex.EncodeToString([]byte("harmlesschild")) + "\n")} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
