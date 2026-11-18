package engine

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestOtherTargetPassForBinaryPythonAndAIModel(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	const marker = "HARMLESS-TARGET14-MARKER"
	rules := map[string][]byte{
		"other.ndb": []byte("Other.Target.Marker:14:*:" + hex.EncodeToString([]byte(marker)) + "\n"),
		"model.ftm": []byte("0:0:47475546:GGUF:CL_TYPE_ANY:CL_TYPE_AI_MODEL\n" +
			"0:0:43505921:Compiled Python:CL_TYPE_ANY:CL_TYPE_PYTHON_COMPILED\n"),
	}
	if _, err := e.ActivateGeneration(rules, 1, manifestHash("other-target")); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		data []byte
		typ  signatures.FileType
	}{
		{"binary", append([]byte{0x7f, 0x80, 0, 1, 0xff, 0xfe}, marker...), signatures.TypeBinaryData},
		{"model", []byte("GGUF-" + marker), signatures.TypeAIModel},
		{"python", []byte("CPY!-" + marker), signatures.TypePythonComp},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, _ := e.current.Load().db.DetermineType(tc.data); got != tc.typ {
				t.Fatalf("type=%v want %v", got, tc.typ)
			}
			clean, name, err := e.Scan(context.Background(), tc.data)
			if err != nil || clean || name != "Other.Target.Marker" {
				t.Fatalf("target 14: clean=%v name=%q err=%v", clean, name, err)
			}
		})
	}
}

func TestOtherTargetPassUsesBuiltInPythonType(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	const marker = "HARMLESS-BUILTIN-PYTHON-TARGET14"
	rules := map[string][]byte{
		"other.ndb": []byte("Other.Target.Python:14:*:" + hex.EncodeToString([]byte(marker)) + "\n"),
	}
	if _, err := e.ActivateGeneration(rules, 1, manifestHash("other-target-builtin-python")); err != nil {
		t.Fatal(err)
	}
	data := append([]byte{0x2d, 0x0d, 0x0d, 0x0a, 0x80}, marker...)
	if got, complete := e.current.Load().db.DetermineType(data); !complete || got != signatures.TypePythonComp {
		t.Fatalf("type=%s complete=%v, want PYTHON_COMPILED", got, complete)
	}
	clean, name, err := e.Scan(context.Background(), data)
	if err != nil || clean || name != "Other.Target.Python" {
		t.Fatalf("target 14: clean=%v name=%q err=%v", clean, name, err)
	}
}
