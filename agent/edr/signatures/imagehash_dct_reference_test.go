package signatures

import (
	"context"
	"math/bits"
	"strconv"
	"testing"
)

func TestDCTJPEGNativeRules(t *testing.T) {
	for _, path := range []string{"testdata/jpeg_boundary_oracle.tar.gz", "testdata/jpeg_cmyk_oracle.tar.gz", "testdata/jpeg_restart_oracle.tar.gz"} {
		cases, _ := loadImageOracleCases(t, path)
		for _, c := range cases {
			t.Run(path+"/"+c.name, func(t *testing.T) {
				want, err := strconv.ParseUint(c.hash, 16, 64)
				if err != nil {
					t.Fatal(err)
				}
				got := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
				partial := false
				if !got.ok || (got.uncertain != 0) != partial || got.hash != want {
					t.Fatalf("JPEG transform: %+v reference=%016x", got, want)
				}
				if partial {
					t.Logf("JPEG mixed-axis partial: hash=%016x uncertain=%016x", got.hash, got.uncertain)
				}
				wrong := want ^ (uint64(1) << bits.TrailingZeros64(^got.uncertain))
				referenceImageRules(t, c.data, want, wrong, partial)
			})
		}
	}
}
