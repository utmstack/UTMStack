package signatures

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestWindowsResizeSineNativeTable(t *testing.T) {
	if len(resizeWindowsCorrections) != 427352 || fmt.Sprintf("%x", sha256.Sum256([]byte(resizeWindowsCorrections))) != "db386f6ce85a5c76cdfc32feabd60012319e73f5ecda7a496dad2aa70bd1963f" {
		t.Fatal("qualified native Windows table changed")
	}
	var previous uint32
	for at := 0; at < len(resizeWindowsCorrections); at += 8 {
		key := binary.LittleEndian.Uint32([]byte(resizeWindowsCorrections[at : at+4]))
		value := binary.LittleEndian.Uint32([]byte(resizeWindowsCorrections[at+4 : at+8]))
		if key <= previous || key >= math.Float32bits(3*float32(math.Pi)) {
			t.Fatal("invalid finite-domain native argument")
		}
		previous = key
		for _, sign := range []uint32{0, 1 << 31} {
			if got := math.Float32bits(resizeWindowsSine32(math.Float32frombits(key | sign))); got != value^sign {
				t.Fatalf("native correction %08x: %08x != %08x", key|sign, got, value^sign)
			}
		}
	}
}

func TestWindowsNativeHashDisagreementRemainsUndecided(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "webp-lossy-tiny-10x11.webp"))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "341d2bd65f9f4a2ee941bb60ac19ddc3187703d21262e39e62979a19221bff9a" {
		t.Fatal("independent native Windows pilot changed")
	}
	got := fuzzyImageHash(raw, &matchWork{ctx: context.Background(), limit: 1 << 40})
	const macLinux, windows = uint64(0x875717525346f8a9), uint64(0x875717565342f8a9)
	if !got.ok || got.hash != macLinux || got.uncertain != (macLinux^windows) {
		t.Fatalf("known native profiles escaped uncertainty: %+v", got)
	}
	for _, hash := range []uint64{macLinux, windows} {
		db := compileTest(t, "ldb", fmt.Sprintf("Native.Profile;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", hash))
		expectOutcome(t, db, string(raw), MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
	}
	negative := compileTest(t, "ldb", fmt.Sprintf("Native.Negative;Engine:150-255,Target:0;0;fuzzy_img#%016x\n", windows^1))
	expectOutcome(t, negative, string(raw), MatchContext{Type: TypeGraphics}, "", "")

}
