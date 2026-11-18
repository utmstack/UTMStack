package signatures

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"testing"
)

func TestBMPRLEFuzzyImageReference(t *testing.T) {
	for _, tc := range []struct {
		file string
		hash uint64
	}{
		{"bmp_rle4.bmp", 0x813b383b383b383f},
		{"bmp_rle8.bmp", 0x84af06b2b54d8ee6},
		{"bmp_rle4_absolute.bmp", 0xdf55155515551555},
		{"bmp_rle8_absolute.bmp", 0xa2b9b84c3c13d5ce},
		{"bmp_rle4_delta.bmp", 0x853a723a723a723e},
		{"bmp_rle8_delta.bmp", 0x856cace3a56c9cc6},
		{"bmp_rle4_odd_literal.bmp", 0x8081ff81ff81bf81},
		{"bmp_rle8_odd_literal.bmp", 0x80817f877f83807f},
		{"bmp_rle8_overrun.bmp", 0x94af16b2354d0ee6},
	} {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			ih := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !ih.ok || ih.hash != tc.hash || ih.uncertain != 0 {
				t.Fatalf("RLE BMP hash %+v, reference %016x", ih, tc.hash)
			}
			rule := compileTest(t, "ldb", fmt.Sprintf("BMP.RLE;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash))
			expectOutcome(t, rule, string(data), MatchContext{Type: TypeGraphics}, "BMP.RLE", "")
			negative := compileTest(t, "ldb", fmt.Sprintf("BMP.Other;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash^1<<40))
			expectOutcome(t, negative, string(data), MatchContext{Type: TypeGraphics}, "", "")
			for _, cut := range []int{1, 2} {
				truncated := bytes.Clone(data[:len(data)-cut])
				expectOutcome(t, rule, string(truncated), MatchContext{Type: TypeGraphics}, "BMP.RLE", "")
			}
			truncated := bytes.Clone(data[:len(data)-3])
			expectOutcome(t, rule, string(truncated), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
		})
	}
}

func TestBMPRLE4EncodedRunOverrunIsIncomplete(t *testing.T) {
	data, err := os.ReadFile("testdata/bmp_rle4.bmp")
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.Clone(data)
	offset := int(binary.LittleEndian.Uint32(data[10:14]))
	data[offset] = 70 // A 64-pixel row; unlike RLE8, the reference rejects this.
	if ih := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); ih.ok {
		t.Fatalf("RLE4 overrun unexpectedly hashed: %+v", ih)
	}
	rule := compileTest(t, "ldb", "BMP.RLE4;Engine:150-255,Target:0;0;fuzzy_img#813b383b383b383f\n")
	expectOutcome(t, rule, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
}
