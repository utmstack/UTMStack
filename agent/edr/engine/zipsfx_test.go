package engine

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"slices"
	"strings"
	"testing"
)

const sfxTypes = "0:0:89504e47:PNG:CL_TYPE_ANY:CL_TYPE_PNG\n0:0:504b0304:ZIP:CL_TYPE_ANY:CL_TYPE_ZIP\n" +
	"1:*:504b0304:zip self-extractor:CL_TYPE_ANY:CL_TYPE_ZIPSFX\n"

var pngPrefix = []byte("\x89PNG\r\n\x1a\n image bytes then ")

// localHeader is one ZIP local file header followed by its name and data.
func localHeader(flags, method uint16, csize, usize uint32, name string, payload []byte) []byte {
	h := make([]byte, 30)
	copy(h, "PK\x03\x04")
	binary.LittleEndian.PutUint16(h[4:], 20)
	binary.LittleEndian.PutUint16(h[6:], flags)
	binary.LittleEndian.PutUint16(h[8:], method)
	binary.LittleEndian.PutUint32(h[18:], csize)
	binary.LittleEndian.PutUint32(h[22:], usize)
	binary.LittleEndian.PutUint16(h[26:], uint16(len(name)))
	return append(append(h, name...), payload...)
}

func sfxState(t *testing.T) *scanState {
	return pipelineState(t, map[string]string{"a.ftm": sfxTypes, "a.ndb": "Embedded.Mark:0:*:" + hx("EMBEDDEDMARKER") + "\n"})
}

func TestEmbeddedZipHeaderThatDoesNotHoldIsIgnored(t *testing.T) {
	s := sfxState(t)
	for name, header := range map[string][]byte{
		"masked":              localHeader(0x2000, 8, 10, 10, "a", []byte("0123456789")),
		"stream past the end": localHeader(0, 8, 1<<30, 1<<30, "a", []byte("0123456789")),
		"name past the end":   append([]byte("PK\x03\x04"), bytes.Repeat([]byte{0xff}, 26)...),
		"empty member":        localHeader(0, 0, 0, 0, "a", nil),
		"streamed garbage":    localHeader(8, 8, 0, 0, "a", []byte("\xff\xff\xff not deflate")),
		"encrypted, no data":  localHeader(1, 8, 1<<30, 1<<30, "a", []byte("0123456789")),
	} {
		data := append(append([]byte(nil), pngPrefix...), header...)
		data = append(data, " trailing bytes of the image"...)
		if sig, err := s.visit(data, 0); sig != "" || err != nil {
			t.Errorf("%s: %q, %v; want clean", name, sig, err)
		}
	}
}

func TestEmbeddedZipMemberIsExtractedFromItsLocalHeader(t *testing.T) {
	s := sfxState(t)
	// A compressed member without any central directory: the local header
	// alone describes it, as the reference reads it.
	member := bytes.Repeat([]byte("a compressed EMBEDDEDMARKER member "), 20)
	var packed bytes.Buffer
	fw, _ := flate.NewWriter(&packed, flate.BestCompression)
	fw.Write(member)
	fw.Close()
	data := append(append([]byte(nil), pngPrefix...), localHeader(0, 8, uint32(packed.Len()), uint32(len(member)), "m.txt", packed.Bytes())...)
	if bytes.Contains(data, []byte("EMBEDDEDMARKER")) {
		t.Fatal("the marker must only appear once the member is decompressed")
	}
	expectScan(t, s, data, "Embedded.Mark", "")

	encrypted := append(append([]byte(nil), pngPrefix...), localHeader(1, 0, 10, 10, "m.txt", []byte("0123456789"))...)
	if sig, err := s.visit(encrypted, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("encrypted member: %q, %v; want incomplete", sig, err)
	}
	// Imploded data is decoded too; data that does not decode yields
	// nothing, as in the reference.
	imploded := implodeLiterals(member)
	implode := append(append([]byte(nil), pngPrefix...), localHeader(0, 6, uint32(len(imploded)), uint32(len(member)), "m.txt", imploded)...)
	expectScan(t, s, implode, "Embedded.Mark", "")
	broken := append(append([]byte(nil), pngPrefix...), localHeader(0, 6, 10, 20, "m.txt", []byte("0123456789"))...)
	if sig, err := s.visit(broken, 0); sig != "" || err != nil {
		t.Fatalf("broken imploded member: %q, %v; want nothing", sig, err)
	}
}

// implodeLiterals implodes data as literals only (flags 0: no literal tree,
// 4 KiB window): two complete trees of 64 six-bit codes, then a set bit and
// the byte's eight bits per byte, least significant bit first.
func implodeLiterals(data []byte) []byte {
	var out []byte
	var acc uint64
	var n uint
	put := func(v uint64, bits uint) {
		acc |= v << n
		for n += bits; n >= 8; n -= 8 {
			out = append(out, byte(acc))
			acc >>= 8
		}
	}
	for _, b := range []byte{3, 0xf5, 0xf5, 0xf5, 0xf5, 3, 0xf5, 0xf5, 0xf5, 0xf5} {
		put(uint64(b), 8)
	}
	for _, b := range data {
		put(1, 1)
		put(uint64(b), 8)
	}
	if n > 0 {
		out = append(out, byte(acc))
	}
	return out
}

func TestZipDeflate64AndImplodeMembers(t *testing.T) {
	member := bytes.Repeat([]byte("a member hiding the "+payloadMarker+" text "), 8)
	var huffman bytes.Buffer
	fw, _ := flate.NewWriter(&huffman, flate.HuffmanOnly) // literals only: valid Deflate64 too
	fw.Write(member)
	fw.Close()
	for name, c := range map[string]struct {
		method uint16
		flags  uint16
		packed []byte
	}{
		"deflate64": {9, 0, huffman.Bytes()},
		"implode":   {6, 0, implodeLiterals(member)},
	} {
		t.Run(name, func(t *testing.T) {
			archive := func(packed []byte) []byte {
				var b bytes.Buffer
				zw := zip.NewWriter(&b)
				w, err := zw.CreateRaw(&zip.FileHeader{Name: "m.bin", Method: c.method, Flags: c.flags,
					CRC32: crc32.ChecksumIEEE(member), CompressedSize64: uint64(len(packed)), UncompressedSize64: uint64(len(member))})
				if err != nil {
					t.Fatal(err)
				}
				w.Write(packed)
				zw.Close()
				return b.Bytes()
			}
			whole := archive(c.packed)
			requireHidden(t, whole)
			expectDetected(t, payloadState(t), whole, "Payload.Test")
			// Cut short, Deflate64 keeps what decoded and implode decodes
			// until its data ends: the first copies of the marker remain.
			expectDetected(t, payloadState(t), archive(c.packed[:len(c.packed)*3/4]), "Payload.Test")
		})
	}
}

func TestSelfExtractorMembersAreEachAnalyzedOnce(t *testing.T) {
	archive := func(last []byte) []byte {
		var b bytes.Buffer
		zw := zip.NewWriter(&b)
		for i := 0; i < 20; i++ {
			content := bytes.Repeat([]byte("clean member content "), 50)
			method := zip.Store
			if i == 19 && last != nil {
				content, method = last, zip.Deflate
			}
			packed := content
			if method == zip.Deflate {
				var p bytes.Buffer
				fw, _ := flate.NewWriter(&p, flate.BestCompression)
				fw.Write(content)
				fw.Close()
				packed = p.Bytes()
			}
			// Sizes in the local header, as archivers that can seek write them.
			f, _ := zw.CreateRaw(&zip.FileHeader{Name: fmt.Sprintf("m%d.txt", i), Method: method, CRC32: crc32.ChecksumIEEE(content),
				CompressedSize64: uint64(len(packed)), UncompressedSize64: uint64(len(content))})
			f.Write(packed)
		}
		zw.Close()
		return append(append([]byte(nil), pngPrefix...), b.Bytes()...)
	}
	s := sfxState(t)
	data := archive(nil)
	// Room for the object and its members a few times over, not for the whole
	// archive once per member.
	s.remaining = int64(4 * len(data))
	if sig, err := s.visit(data, 0); sig != "" || err != nil {
		t.Fatalf("%q, %v; want clean within the budget", sig, err)
	}
	data = archive(bytes.Repeat([]byte("the last member holds EMBEDDEDMARKER "), 20))
	if bytes.Contains(data, []byte("EMBEDDEDMARKER")) {
		t.Fatal("the marker must only appear once the member is decompressed")
	}
	s = sfxState(t)
	s.remaining = int64(4 * len(data))
	expectScan(t, s, data, "Embedded.Mark", "")
}

// A damaged deflate member found inside another object is read as the
// reference reads it (docs/2026-10-14-cfb-oracle/ziplab.py): data that ends
// inside the stream yields what decoded up to there; a stream that fails to
// decode yields nothing, even when its first blocks decoded.
func TestEmbeddedZipDamagedDeflateAsReference(t *testing.T) {
	s := sfxState(t)
	member := append([]byte("EMBEDDEDMARKER "), bytes.Repeat([]byte("filler text of the member "), 4000)...)
	var packed bytes.Buffer
	fw, _ := flate.NewWriter(&packed, flate.DefaultCompression)
	var flushes []int
	for i := 0; i < len(member); i += 8192 {
		fw.Write(member[i:min(len(member), i+8192)])
		fw.Flush()
		flushes = append(flushes, packed.Len()) // a block starts byte-aligned here
	}
	fw.Close()
	full := packed.Bytes()
	embed := func(csize int, data []byte) []byte {
		out := append(append([]byte(nil), pngPrefix...), localHeader(0, 8, uint32(csize), uint32(len(member)), "m.txt", data)...)
		return append(out, " trailing bytes of the image"...)
	}
	// The declared size ends inside the stream: the decoded start is scanned.
	expectScan(t, s, embed(len(full)/2, full[:len(full)/2]), "Embedded.Mark", "")
	// A reserved block type after the marker's block: nothing is scanned.
	bad := append([]byte(nil), full...)
	bad[flushes[1]] = 0x07
	if sig, err := s.visit(embed(len(bad), bad), 0); sig != "" || err != nil {
		t.Fatalf("undecodable member: %q, %v; want clean", sig, err)
	}
}

// testdata/zip/oracle.zip holds damaged archives and objects holding them
// (docs/2026-10-14-cfb-oracle/ziplab.py, ziplab2.py) with the markers the
// reference found in each: ZIPLAB-MARK-ONE near the start of the member's
// text and ZIPLAB-MARK-TWO at its end, so a member read only in part shows
// one. Each marker is checked on its own, as this engine reports the first
// detection only.
func TestZIPReadsAsReference(t *testing.T) {
	z, err := zip.OpenReader("testdata/zip/oracle.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	files := map[string][]byte{}
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		files[f.Name], err = io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	var expected map[string][]string
	if err := json.Unmarshal(files["expected.json"], &expected); err != nil {
		t.Fatal(err)
	}
	if len(expected) != 125 {
		t.Fatalf("%d cases", len(expected))
	}
	types := "0:0:d0cf11e0a1b11ae1:OLE2 container:CL_TYPE_ANY:CL_TYPE_MSOLE2\n" +
		"0:0:504b0304:ZIP:CL_TYPE_ANY:CL_TYPE_ZIP\n1:*:504b0304:zip self-extractor:CL_TYPE_ANY:CL_TYPE_ZIPSFX\n"
	marks := map[string]string{"Mark1": "ZIPLAB-MARK-ONE-7c1e", "Mark2": "ZIPLAB-MARK-TWO-93fa"}
	// Where this engine differs on purpose: encrypted members and unknown
	// compression methods are incomplete (the reference skips them), and a
	// member whose sizes follow its data, found inside another object, is
	// decoded (the reference does not read it).
	incomplete := map[string]bool{"zl3/a_encrypted_flag.zip": true, "zl3/b_enc_local.zip": true, "zl3/a_method_unknown.zip": true}
	for name, want := range expected {
		for mark, text := range marks {
			s := pipelineState(t, map[string]string{"a.ftm": types, "a.ndb": "ZipLab." + mark + ":0:*:" + hx(text) + "\n"})
			sig, err := s.visit(files[name], 0)
			wantFound := slices.Contains(want, mark)
			switch {
			case incomplete[name]:
				if sig != "" || !errors.Is(err, ErrUnsupported) {
					t.Errorf("%s %s: sig=%q err=%v, want incomplete", name, mark, sig, err)
				}
			case strings.Contains(name, "_z_desc_"):
				if sig == "" && mark == "Mark1" {
					t.Errorf("%s: the streamed member is no longer decoded", name)
				}
			case name == "zl3/d_zip64.zip":
				if sig != "ZipLab."+mark || err != nil {
					t.Errorf("%s %s: sig=%q err=%v, want ZIP64 extraction", name, mark, sig, err)
				}
			case (sig != "") != wantFound || sig == "" && err != nil:
				t.Errorf("%s %s: sig=%q err=%v, the reference found it: %v", name, mark, sig, err, wantFound)
			}
		}
	}
}
