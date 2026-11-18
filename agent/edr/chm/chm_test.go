package chm

import (
	"bytes"
	"errors"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/chm/chmtest"
)

func TestReadsBothSections(t *testing.T) {
	entries := []chmtest.Entry{
		{Name: "/#SYSTEM", Section: 0, Data: []byte("stored system data")},
		{Name: "/page.htm", Section: 1, Data: bytes.Repeat([]byte("<p>compressed page</p>"), 300)},
		{Name: "/empty", Section: 1},
		{Name: "/images/", Section: 0},
		{Name: "/second.htm", Section: 1, Data: []byte("second compressed file")},
	}
	a, err := Open(chmtest.Build(entries), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Files) != 4 || len(a.System) != 3 {
		t.Fatalf("%d files, %d system files", len(a.Files), len(a.System))
	}
	for i, f := range a.Files {
		want := entries[[]int{0, 1, 2, 4}[i]]
		b, err := a.Read(f)
		if f.Name != want.Name || err != nil || !bytes.Equal(b, want.Data) {
			t.Errorf("%s: %d bytes, %v", f.Name, len(b), err)
		}
	}
}

func TestDamage(t *testing.T) {
	good := chmtest.Build([]chmtest.Entry{{Name: "/a.htm", Section: 1, Data: bytes.Repeat([]byte("some compressed text"), 20)}})
	for name, edit := range map[string]func(b []byte){
		"version":  func(b []byte) { b[4] = 9 },
		"dir past": func(b []byte) { b[0x48] = 0xff; b[0x49] = 0xff },
		"no ITSP":  func(b []byte) { b[0x78] = 'X' },
	} {
		b := append([]byte(nil), good...)
		edit(b)
		if _, err := Open(b, 1<<20); !errors.Is(err, ErrFormat) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if a, err := Open(good, 100); err != nil {
		t.Fatal(err)
	} else if _, err := a.Read(a.Files[0]); !errors.Is(err, ErrLimit) {
		t.Errorf("section beyond the limit: %v", err)
	}
}

func FuzzOpen(f *testing.F) {
	f.Add(chmtest.Build([]chmtest.Entry{{Name: "/a", Section: 1, Data: []byte("fuzz seed text")}, {Name: "/b", Data: []byte("x")}}))
	f.Fuzz(func(t *testing.T, b []byte) {
		a, err := Open(b, 1<<20)
		if err != nil {
			return
		}
		for _, file := range a.Files {
			a.Read(file)
		}
	})
}
