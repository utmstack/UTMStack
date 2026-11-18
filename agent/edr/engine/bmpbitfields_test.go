package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestBMPBitfieldFuzzyRuleEndToEnd(t *testing.T) {
	db, err := signatures.Compile(map[string][]byte{
		"types.ftm": []byte("0:0:424d:BMP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"),
		"probe.ldb": []byte("BMP.RGB555;Engine:150-255,Target:0;0;fuzzy_img#d88fd8cf84f084f1\n" +
			"BMP.RGB565;Engine:150-255,Target:0;0;fuzzy_img#cc5db29bac00ecf1\n" +
			"BMP.RGB1010102;Engine:150-255,Target:0;0;fuzzy_img#8f4f010b031f9f1f\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	e := New(config.EDRConfig{})
	e.current.Store(&generation{db: db})
	for _, tc := range []struct{ file, want string }{
		{"bmp_bitfields_rgb555.bmp", "BMP.RGB555"},
		{"bmp_bitfields_rgb565.bmp", "BMP.RGB565"},
		{"bmp_bitfields_rgb1010102.bmp", "BMP.RGB1010102"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "signatures", "testdata", tc.file))
			if err != nil {
				t.Fatal(err)
			}
			clean, name, err := e.Scan(context.Background(), data)
			if clean || name != tc.want || err != nil {
				t.Fatalf("bitfield BMP scan: clean=%v name=%q err=%v", clean, name, err)
			}
		})
	}
}
