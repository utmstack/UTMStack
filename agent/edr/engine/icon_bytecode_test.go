package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestLogicalBytecodeIconCondition(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "icon_bytecode.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Image    []byte
		SHA256   string
		IconHash string `json:"icon_hash"`
	}
	if err = json.Unmarshal(body, &fixture); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(fixture.Image)) != fixture.SHA256 {
		t.Fatal("bytecode icon input differs from reference")
	}
	for _, tc := range []struct {
		name           string
		want           uint64
		hash           string
		absent, detect bool
	}{
		{"matching-icon-callback-detects", 0, fixture.IconHash, false, true},
		{"matching-icon-callback-declines", 1, fixture.IconHash, false, false},
		{"nonmatching-icon-blocks-callback", 0, "10" + strings.Repeat("0", 122), false, false},
		{"unmatched-expression", 0, fixture.IconHash, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			program := matchIconOracleProgram(256, "A", "B", tc.want)
			program = bytes.Replace(program, []byte("Target:1;0;"), []byte("Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;"), 1)
			if tc.absent {
				program = bytes.Replace(program, []byte(hex.EncodeToString([]byte("UTMSTACK-MATCHICON-HARMLESS-TRIGGER"))), []byte(hex.EncodeToString([]byte("ABSENT-ICON-TRIGGER"))), 1)
			}
			input := append(append([]byte(nil), fixture.Image...), []byte("UTMSTACK-MATCHICON-HARMLESS-TRIGGER")...)
			cfg := config.Default()
			cfg.SignatureDBDir = t.TempDir()
			for name, data := range map[string][]byte{"probe.idb": []byte("Safe.Icon:A:B:" + tc.hash + "\n"), "probe.cbc": program} {
				if err = os.WriteFile(filepath.Join(cfg.SignatureDBDir, name), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			e := New(cfg)
			if err = e.EnsureRunning(); err != nil {
				t.Fatal(err)
			}
			clean, name, err := e.Scan(context.Background(), input)
			if err != nil || clean == tc.detect || tc.detect && name != "Bytecode.Icon.Detected" || !tc.detect && name != "" {
				t.Fatalf("icon/callback sequence: clean=%t name=%q err=%v", clean, name, err)
			}
			if dir := os.Getenv("OPENEDR_ICON_CBC_LAB"); dir != "" {
				root := filepath.Join(dir, tc.name)
				if err = os.MkdirAll(root, 0700); err != nil {
					t.Fatal(err)
				}
				for file, data := range map[string][]byte{"probe.idb": []byte("Safe.Icon:A:B:" + tc.hash + "\n"), "probe.cbc": program, "object.exe": input} {
					if err = os.WriteFile(filepath.Join(root, file), data, 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
		})
	}
}
