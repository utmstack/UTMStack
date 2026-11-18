package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestBMPRLEFuzzyRuleEndToEnd(t *testing.T) {
	db, err := signatures.Compile(map[string][]byte{
		"types.ftm": []byte("0:0:424d:BMP:CL_TYPE_ANY:CL_TYPE_GRAPHICS\n"),
		"probe.ldb": []byte("BMP.RLE4;Engine:150-255,Target:0;0;fuzzy_img#813b383b383b383f\n" +
			"BMP.RLE8;Engine:150-255,Target:0;0;fuzzy_img#84af06b2b54d8ee6\n" +
			"BMP.RLE8.Overrun;Engine:150-255,Target:0;0;fuzzy_img#94af16b2354d0ee6\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	e := New(config.EDRConfig{})
	e.current.Store(&generation{db: db})
	for _, tc := range []struct{ file, want string }{
		{"bmp_rle4.bmp", "BMP.RLE4"},
		{"bmp_rle8.bmp", "BMP.RLE8"},
		{"bmp_rle8_overrun.bmp", "BMP.RLE8.Overrun"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "signatures", "testdata", tc.file))
			if err != nil {
				t.Fatal(err)
			}
			clean, name, err := e.Scan(context.Background(), data)
			if clean || name != tc.want || err != nil {
				t.Fatalf("RLE BMP scan: clean=%v name=%q err=%v", clean, name, err)
			}
		})
	}
}
