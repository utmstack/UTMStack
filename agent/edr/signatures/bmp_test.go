package signatures

import (
	"context"
	"testing"
)

func TestBMPImageHashesAgreeWithReference(t *testing.T) {
	cases, files := loadImageOracleCases(t, "testdata/bmp_oracle.tar.gz")
	if len(cases) != 60 {
		t.Fatalf("got %d BMP fixtures, want 60", len(cases))
	}
	for _, c := range cases {
		if !fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40}).ok {
			t.Errorf("%s was not decoded", c.name)
		}
	}
	checkImageCases(t, cases, 0.9)
	flat := compileTest(t, "ldb", "BMP.Flat;Engine:150-255,Target:0;0;fuzzy_img#8000000000000000#0\n")
	for _, name := range []string{"00-rgb.bmp", "00-gray.bmp", "00-palette.bmp", "00-rgba.bmp"} {
		got, err := matchObject(t, flat, string(files[name]), MatchContext{Type: TypeGraphics})
		if got != "BMP.Flat" || err != nil {
			t.Errorf("%s: match %q, %v", name, got, err)
		}
	}
	expectOutcome(t, flat, "BMabcdef", MatchContext{Type: TypeGraphics}, "", UnknownFuzzyImage)
}
