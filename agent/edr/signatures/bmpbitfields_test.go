package signatures

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"testing"
)

func TestBMPBitfieldFuzzyImageReference(t *testing.T) {
	for _, tc := range []struct {
		name string
		hash uint64
	}{
		{"bgr888", 0xcda363abbe24aa88},
		{"rgb1010102", 0x8f4f010b031f9f1f},
		{"rgb555", 0xd88fd8cf84f084f1},
		{"rgb565", 0xcc5db29bac00ecf1},
		{"rgba4444_top", 0xd5fc2ad5a0d580d5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile("testdata/bmp_bitfields_" + tc.name + ".bmp")
			if err != nil {
				t.Fatal(err)
			}
			ih := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !ih.ok || ih.hash != tc.hash || ih.uncertain != 0 {
				t.Fatalf("bitfield BMP hash %+v, reference %016x", ih, tc.hash)
			}
			rule := compileTest(t, "ldb", fmt.Sprintf("BMP.Bitfields;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash))
			expectOutcome(t, rule, string(data), MatchContext{Type: TypeGraphics}, "BMP.Bitfields", "")
			negative := compileTest(t, "ldb", fmt.Sprintf("BMP.Other;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", tc.hash^1<<40))
			expectOutcome(t, negative, string(data), MatchContext{Type: TypeGraphics}, "", "")
		})
	}
}

func TestBMPBitfieldMalformedRemainsIncomplete(t *testing.T) {
	original, err := os.ReadFile("testdata/bmp_bitfields_rgb565.bmp")
	if err != nil {
		t.Fatal(err)
	}
	rule := compileTest(t, "ldb", "BMP.RGB565;Engine:150-255,Target:0;0;fuzzy_img#cc5db29bac00ecf1\n")
	for _, tc := range []struct {
		name string
		edit func([]byte) []byte
	}{
		{"truncated-pixels", func(b []byte) []byte { return b[:len(b)-1] }},
		{"noncontiguous-mask", func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[54:58], 0x0000a000)
			return b
		}},
		{"oversize-width", func(b []byte) []byte {
			binary.LittleEndian.PutUint32(b[18:22], 1<<20)
			return b
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := tc.edit(bytes.Clone(original))
			if ih := fuzzyImageHash(data, &matchWork{ctx: context.Background(), limit: 1 << 40}); ih.ok {
				t.Fatalf("malformed BMP unexpectedly hashed: %+v", ih)
			}
			expectOutcome(t, rule, string(data), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
		})
	}
}
