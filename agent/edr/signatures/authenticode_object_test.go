package signatures

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestAuthenticodeObjectReference(t *testing.T) {
	file, err := os.Open("testdata/authenticode_objects.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := bufio.NewScanner(file)
	count, accepted := 0, 0
	for reader.Scan() {
		line := reader.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 6 {
			t.Fatal("invalid independent object record")
		}
		var data []byte
		if parts[0] != "-" {
			data, err = hex.DecodeString(parts[0])
			if err != nil {
				t.Fatal(err)
			}
		}
		var want [5]int
		for i := range want {
			want[i], err = strconv.Atoi(parts[i+1])
			if err != nil {
				t.Fatal(err)
			}
		}
		value, tail, err := asn1Object(data)
		if (err == nil) != (want[0] == 1) {
			t.Fatalf("row %d acceptance=%v want=%v", count, err == nil, want[0])
		}
		if err == nil {
			accepted++
			tag := value.Class<<6 | value.Tag
			if value.IsCompound {
				tag |= 0x20
			}
			if tag != want[1] || len(value.FullBytes)-len(value.Bytes) != want[2] || len(value.Bytes) != want[3] || len(value.FullBytes) != want[4] || len(data)-len(tail) != want[4] || !bytes.Equal(value.FullBytes, data[:want[4]]) {
				t.Fatalf("row %d changed framing", count)
			}
		}
		count++
	}
	if err := reader.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 71380 || accepted != 34368 {
		t.Fatalf("incomplete object oracle: %d/%d", count, accepted)
	}
}

func FuzzAuthenticodeObject(f *testing.F) {
	for _, data := range [][]byte{nil, {0x30, 0}, {0x30, 0x84, 0, 0, 0, 0}, {0xff, 0x81, 0}, {0x30, 0x80, 0, 0}, {0x30, 0x84, 0xff, 0xff, 0xff, 0xff}} {
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			return
		}
		before := bytes.Clone(data)
		value, tail, err := asn1Object(data)
		if !bytes.Equal(data, before) {
			t.Fatal("framing reader mutated input")
		}
		if err != nil {
			return
		}
		if len(value.FullBytes) < 2 || len(value.FullBytes)+len(tail) != len(data) || len(value.Bytes) > len(value.FullBytes)-2 || len(value.FullBytes)-len(value.Bytes) > 6 {
			t.Fatal("framing escaped parent or header bound")
		}
		if !bytes.Equal(value.FullBytes, data[:len(value.FullBytes)]) || !bytes.Equal(value.Bytes, value.FullBytes[len(value.FullBytes)-len(value.Bytes):]) {
			t.Fatal("framing normalized original signed bytes")
		}
	})
}
