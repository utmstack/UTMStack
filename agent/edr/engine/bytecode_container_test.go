package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

const bytecodeContainerOracle = "../../../docs/edr/native-engine/audits/2026-11-01-v191-container-type/oracle/"

func TestBytecodeForcedContainerTypeScansChildRoot(t *testing.T) {
	childRule, err := os.ReadFile(bytecodeContainerOracle + "child.ndb")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		mode       string
		wantName   string
		incomplete bool
	}{
		{"with", "Child.Other", false},
		{"without", "", false},
		{"invalid", "", false},
		{"gap", "", true},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			program, err := os.ReadFile(bytecodeContainerOracle + tc.mode + ".cbc")
			if err != nil {
				t.Fatal(err)
			}
			cfg := config.Default()
			cfg.SignatureDBDir = t.TempDir()
			e := New(cfg)
			if _, err := e.ActivateGeneration(map[string][]byte{
				"program.cbc": program, "child.ndb": childRule,
			}, 1, manifestHash("container-"+tc.mode)); err != nil {
				t.Fatal(err)
			}
			for _, sample := range []struct {
				name string
				want string
			}{{"sample-ab.bin", tc.wantName}, {"sample-ac.bin", ""}} {
				data, err := os.ReadFile(filepath.Join(bytecodeContainerOracle, sample.name))
				if err != nil {
					t.Fatal(err)
				}
				clean, name, _, err := e.ScanSnapshot(context.Background(), data)
				if sample.name == "sample-ab.bin" && tc.incomplete {
					if clean || name != "" || !errors.Is(err, ErrUnsupported) {
						t.Fatalf("unknown forced type: clean %v, name %q, err %v", clean, name, err)
					}
					continue
				}
				if err != nil || clean != (sample.want == "") || name != sample.want {
					t.Fatalf("%s: clean %v, name %q, err %v; want %q", sample.name, clean, name, err, sample.want)
				}
			}
		})
	}
}
