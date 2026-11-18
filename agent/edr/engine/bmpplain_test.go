package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestBMPPlainFuzzyRuleEndToEnd(t *testing.T) {
	db, err := signatures.Compile(map[string][]byte{
		"types.ftm": []byte("0:0:424d:BMP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"),
		"probe.ldb": []byte("BMP.Core1;Engine:150-255,Target:0;0;fuzzy_img#ed83a7d05aa98c17\n" +
			"BMP.Info2Gap;Engine:150-255,Target:0;0;fuzzy_img#d5aa12aac0ef90eb\n" +
			"BMP.ShortPalette;Engine:150-255,Target:0;0;fuzzy_img#9f474ca59d78c059\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	e := New(config.EDRConfig{})
	e.current.Store(&generation{db: db})
	for _, tc := range []struct{ file, want string }{
		{"bmp_plain_core1.bmp", "BMP.Core1"},
		{"bmp_plain_info2_gap.bmp", "BMP.Info2Gap"},
		{"bmp_plain_info8_shortpal.bmp", "BMP.ShortPalette"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "signatures", "testdata", tc.file))
			if err != nil {
				t.Fatal(err)
			}
			clean, name, err := e.Scan(context.Background(), data)
			if clean || name != tc.want || err != nil {
				t.Fatalf("plain BMP scan: clean=%v name=%q err=%v", clean, name, err)
			}
		})
	}
}
