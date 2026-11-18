package signatures

import (
	"bufio"
	"encoding/asn1"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestAuthenticodeTimeMappedWindows(t *testing.T) {
	for _, tc := range []struct {
		wire string
		tag  int
		want time.Time
	}{
		{"200101000000Z", asn1.TagUTCTime, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"400101000000Z", asn1.TagUTCTime, time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2001010000Z", asn1.TagUTCTime, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"20400101000000Z", asn1.TagGeneralizedTime, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
	} {
		encoded, err := asn1.Marshal(asn1.RawValue{Tag: tc.tag, Bytes: []byte(tc.wire)})
		if err != nil {
			t.Fatal(err)
		}
		data := append(append([]byte(nil), encoded...), 0, 0)
		got, n, err := referenceAuthenticodeTime(data, len(encoded))
		if err != nil || n != len(encoded) || !got.Equal(tc.want) {
			t.Fatalf("%s: got=%v n=%d err=%v want=%v", tc.wire, got, n, err, tc.want)
		}
	}
}

func TestAuthenticodeTimeReference(t *testing.T) {
	file, err := os.Open(filepath.Join("testdata", "authenticode_times.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) != 5 {
			t.Fatal("invalid reference row", count)
		}
		data, err := hex.DecodeString(fields[0])
		if err != nil {
			t.Fatal(err)
		}
		limit, err := strconv.Atoi(fields[1])
		if err != nil {
			t.Fatal(err)
		}
		consumed, err := strconv.Atoi(fields[3])
		if err != nil {
			t.Fatal(err)
		}
		epoch, err := strconv.ParseInt(fields[4], 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		at, n, err := referenceAuthenticodeTime(data, limit)
		accepted := fields[2] == "1"
		if (err == nil) != accepted || accepted && (n != consumed || at.Unix() != epoch) {
			t.Fatalf("reference row %d (%s): at=%v n=%d err=%v; accepted=%v n=%d unix=%d", count, fields[0], at, n, err, accepted, consumed, epoch)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 49342 {
		t.Fatal("incomplete time reference", count)
	}
}

func FuzzAuthenticodeTimeWindow(f *testing.F) {
	for _, body := range []string{"240101000000Z", "2401010000Z", "240231000000Zignored"} {
		data := append([]byte{23, byte(len(body))}, []byte(body)...)
		f.Add(append(data, 0, 0), uint32(len(data)))
	}
	f.Add([]byte{24, 15, '2', '0', 'x', '?', '0', '1', '0', '1', '0', '0', '0', '0', '0', '0', 'Z'}, uint32(17))
	f.Fuzz(func(t *testing.T, data []byte, limit uint32) {
		if len(data) > 4096 {
			t.Skip()
		}
		at, n, err := referenceAuthenticodeTime(data, int(limit))
		if err == nil && (n < 2 || n > len(data) || uint32(n) > limit || at.Location() != time.UTC || at.Nanosecond() != 0 || at.Year() < 0 || at.Year() > 9999) {
			t.Fatal("time read escaped its physical/parent/UTC bounds")
		}
	})
}
