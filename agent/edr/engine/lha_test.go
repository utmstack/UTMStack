package engine

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func lhaLArcTestMember(method string, child, packed []byte) []byte {
	h := make([]byte, 33)
	h[0] = 31
	copy(h[2:7], method)
	binary.LittleEndian.PutUint32(h[7:11], uint32(len(packed)))
	binary.LittleEndian.PutUint32(h[11:15], uint32(len(child)))
	h[19], h[21] = 32, 9
	copy(h[22:31], "child.txt")
	binary.LittleEndian.PutUint16(h[31:33], lhaCRC16(child))
	for _, b := range h[2:] {
		h[1] += b
	}
	return append(append(h, packed...), 0)
}

func lhaLevel3TestMember(method string, child, packed []byte, extCount int) []byte {
	records := [][]byte{[]byte("\x01child.txt"), []byte("\x02DIR\xff")}
	ext := make([]byte, 0, 32)
	for _, record := range records[:extCount] {
		ext = binary.LittleEndian.AppendUint32(ext, uint32(len(record)+4))
		ext = append(ext, record...)
	}
	ext = binary.LittleEndian.AppendUint32(ext, 0)
	h := make([]byte, 28+len(ext))
	binary.LittleEndian.PutUint16(h[:2], 4)
	copy(h[2:7], method)
	binary.LittleEndian.PutUint32(h[7:11], uint32(len(packed)))
	binary.LittleEndian.PutUint32(h[11:15], uint32(len(child)))
	h[19], h[20], h[23] = 32, 3, 'U'
	binary.LittleEndian.PutUint16(h[21:23], lhaCRC16(child))
	binary.LittleEndian.PutUint32(h[24:28], uint32(len(h)))
	copy(h[28:], ext)
	return append(append(h, packed...), 0)
}

func TestLHALevelThreeChildHashes(t *testing.T) {
	for _, tc := range []struct {
		name, method, child, packed, sha string
		extCount                         int
	}{
		{"stored-ext0", "-lh0-", "harmless level three child", "6861726d6c657373206c6576656c207468726565206368696c64", "250a3174c9a21d4082e0723882bf2b4fcc5e54b4ed9d13eb73caab2c191c110e", 0},
		{"stored-ext1", "-lh0-", "harmless level three child", "6861726d6c657373206c6576656c207468726565206368696c64", "f860a5bb5dec93736c2617fc0be4471d20ffb13a7b2138e343202c3e1a4651c5", 1},
		{"stored-ext2", "-lh0-", "harmless level three child", "6861726d6c657373206c6576656c207468726565206368696c64", "f681183587283304d97c4cdbbefee217adc46bcb799eef3c5a4efbf1371136d6", 2},
		{"lh5-ext1", "-lh5-", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "00400000041000", "644ec789545c8d284796d583b8f9296ac0cd56acce2a183c8b6a557296ad7f9b", 1},
		{"lz5-ext2", "-lz5-", "ABCDEFGHIJKLMNOPQRSTABC", "ff4142434445464748ff494a4b4c4d4e4f500f51525354eef0", "4395e72912acc56081ab15686ef587b1b5f885a2b550928e12156a891992d63e", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := hex.DecodeString(tc.packed)
			if err != nil {
				t.Fatal(err)
			}
			archive := lhaLevel3TestMember(tc.method, []byte(tc.child), packed, tc.extCount)
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.sha {
				t.Fatalf("reference fixture changed: %s", got)
			}
			sum := md5.Sum([]byte(tc.child))
			rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:Reference.LHA.Level3\n", sum, len(tc.child))}
			expectDetected(t, pipelineState(t, rules), archive, "Reference.LHA.Level3")
			expectScan(t, pipelineState(t, map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}), archive, "", "")
		})
	}
}

func TestLH1HistoricalChildHash(t *testing.T) {
	for _, tc := range []struct{ file, sha string }{
		{"lh1-lharc113.lzh", "594af58cb27d6a80e9bebd108cc82a8ab881a4b5549e4fc10379fbf7a0d83069"},
		{"lh1-os9-h2.lzh", "e6aaa8a5e7520d8cda49c29befe8edf6570f2aac425154e23483582b086550af"},
		{"lh1-morphos-h0.lzh", "36503045e565c50a1f4cef66dd8c544fcb2ce4ac18f9cd4d55f4f0051b3d5b6e"},
		{"lh1-morphos-h1.lzh", "921399b4f348489d9ca1e5c0b93620164bbe6493e7eae48cee2ebaf71558ade0"},
		{"lh1-morphos-h2.lzh", "fba263ff370eee30485d4be73d13fbb9ceeb0edf5691253288e83884ab30dd21"},
		{"lh1-amiga.lzh", "4981a27f0e051b33d7813a630f970d820703f61c14f896dc5ed2a84f8240a836"},
		{"lh1-osk-h2.lzh", "82020122cf7ad688b4182fe26435c808c5f1b729eb16859dfdf42a173da36a81"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			archive, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.sha {
				t.Fatalf("reference archive changed: %s", got)
			}
			m, err := lhaMemberOf(archive)
			if err != nil || string(m.method[:]) != "-lh1-" || m.plain != 18092 {
				t.Fatalf("member: %+v %v", m, err)
			}
			body := archive[int(m.start):int(m.start+m.packed)]
			child, err := lh1Expand(context.Background(), body, int(m.plain))
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(child)); got != "8177f97513213526df2cf6184d8ff986c675afb514d4e68a404010521b880643" {
				t.Fatalf("decoded child changed: %s", got)
			}
			if lhaCRC16(child) != m.crc {
				t.Fatal("member CRC does not match")
			}
			sum := md5.Sum(child)
			rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:Reference.LH1.Child\n", sum, len(child))}
			expectDetected(t, pipelineState(t, rules), archive, "Reference.LH1.Child")
			expectScan(t, pipelineState(t, map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}), archive, "", "")
		})
	}
}

func TestLH1HistoricalTreeRebuild(t *testing.T) {
	archive, err := os.ReadFile("testdata/lh1-lharc113-long.lzh")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != "fe1d64ba8b8e67a32d3d2bfa910fc13a70d220d4bfe4f68987a6d9c3a77ebd03" {
		t.Fatalf("reference archive changed: %s", got)
	}
	m, err := lhaMemberOf(archive)
	if err != nil || string(m.method[:]) != "-lh1-" || m.plain != 1241658 {
		t.Fatalf("member: %+v %v", m, err)
	}
	child, err := lh1Expand(context.Background(), archive[int(m.start):int(m.start+m.packed)], int(m.plain))
	if err != nil {
		t.Fatal(err)
	}
	if lhaCRC16(child) != m.crc {
		t.Fatalf("decoded child CRC mismatch: got %04x want %04x", lhaCRC16(child), m.crc)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(child)); got != "1211b353951c19b6e69a28c1f7ed5bdf123015e6e22b5d3109135c76f8488188" {
		t.Fatalf("decoded child changed: %s", got)
	}
	sum := md5.Sum(child)
	rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:Reference.LH1.Long\n", sum, len(child))}
	expectDetected(t, pipelineState(t, rules), archive, "Reference.LH1.Long")
	expectScan(t, pipelineState(t, map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}), archive, "", "")
}

func TestLH1TruncatedDamagedLimitedAndCanceled(t *testing.T) {
	archive, err := os.ReadFile("testdata/lh1-lharc113.lzh")
	if err != nil {
		t.Fatal(err)
	}
	m, err := lhaMemberOf(archive)
	if err != nil {
		t.Fatal(err)
	}
	body := archive[int(m.start):int(m.start+m.packed)]
	if _, err := lh1Expand(context.Background(), body[:10], int(m.plain)); !errors.Is(err, ErrMalformed) {
		t.Fatalf("truncated LH1: %v", err)
	}
	clean := map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}
	badCRC := append([]byte(nil), archive...)
	badCRC[22+int(badCRC[21])]++
	expectIncomplete(t, pipelineState(t, clean), badCRC, ErrMalformed)
	limited := pipelineState(t, clean)
	limited.maxFile = 18091
	expectIncomplete(t, limited, archive, ErrLimit)
	osk, err := os.ReadFile("testdata/lh1-osk-h2.lzh")
	if err != nil {
		t.Fatal(err)
	}
	badTerm := append([]byte(nil), osk...)
	badTerm[44] = 1 // OSK's two-byte extension terminator lies past its stated header length.
	expectIncomplete(t, pipelineState(t, clean), badTerm, ErrMalformed)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := lh1Expand(ctx, body, int(m.plain)); !errors.Is(err, context.Canceled) {
		t.Fatalf("LH1 cancellation: %v", err)
	}
}

func TestLH1RandomBitstreamsStayBounded(t *testing.T) {
	rng := rand.New(rand.NewSource(141))
	for i := 0; i < 200; i++ {
		in := make([]byte, 64)
		_, _ = rng.Read(in)
		out, err := lh1Expand(context.Background(), in, 1024)
		if len(out) > 1024 || err != nil && !errors.Is(err, ErrMalformed) {
			t.Fatalf("case %d: output %d err %v", i, len(out), err)
		}
	}
}

func TestLArcMethodsMatchReferenceChildHashes(t *testing.T) {
	for _, tc := range []struct {
		name, method, packed, child, sha string
	}{
		{"lz4-stored", "-lz4-", "4142434445464748", "ABCDEFGH", "4b7c2a0dd04fb443dc0c5f739f9e439a979bb28e210257903e41e6ff22710235"},
		{"lzs-literal", "-lzs-", "a0d0a8744a2d1a8f48", "ABCDEFGH", "a5ca932291427b25e2cb01c9b6eea3dcbc24bacf769d1f218ba804117fef5a42"},
		{"lz5-literal", "-lz5-", "ff4142434445464748", "ABCDEFGH", "5b5aeafacdf73c184e824ce8680943a5e0f0ddb6a62019b555a06ce3b924e063"},
		{"lzs-backref", "-lzs-", "a0bf7a00", "AAAAAAA", "85e908fd0731603dbbbc562822fdc3c481b8899528026cc98d1a44ccf0053859"},
		{"lz5-backref", "-lz5-", "0141eef3", "AAAAAAA", "9dd07e69241e4c73b087bba60b6f0f74ddbce685487d41a3c8e7134356499b5a"},
		{"lzs-preset", "-lzs-", "0006", "        ", "8ecab8e06881cf29795f2d48822166460cdd0a801b92af6a8da766a859e19646"},
		{"lz5-preset", "-lz5-", "004d35", "AAAAAAAA", "14c2c956ba91035fb405792bdc973c8530d542909c0129e5430396f14fe8a607"},
		{"lzs-wrap", "-lzs-", "a0d0a8744a2d1a8f48a4d2a974ca6d3a9f50a8d4aa7547ef10", "ABCDEFGHIJKLMNOPQRSTABC", "c0a0084e98d9b041b69fa4f7fc04d4f435d2d549f2ad4316cfe57e6003767928"},
		{"lz5-wrap-groups", "-lz5-", "ff4142434445464748ff494a4b4c4d4e4f500f51525354eef0", "ABCDEFGHIJKLMNOPQRSTABC", "57732afea78e7383509cfbdade7feb89ef21c6c05d1d90e82192c0e4f2ec28f7"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := hex.DecodeString(tc.packed)
			if err != nil {
				t.Fatal(err)
			}
			archive := lhaLArcTestMember(tc.method, []byte(tc.child), packed)
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.sha {
				t.Fatalf("reference fixture changed: %s", got)
			}
			sum := md5.Sum([]byte(tc.child))
			rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:Reference.LArc.Child\n", sum, len(tc.child))}
			expectDetected(t, pipelineState(t, rules), archive, "Reference.LArc.Child")
			expectScan(t, pipelineState(t, map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}), archive, "", "")
		})
	}
}

func TestLArcTruncationCRCLimitAndCancellation(t *testing.T) {
	clean := map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}
	for _, tc := range []struct {
		method string
		packed []byte
	}{
		{"-lzs-", []byte{0xa0, 0xbf, 0x7a, 0x00}},
		{"-lz5-", []byte{0x01, 'A', 0xee, 0xf3}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			archive := lhaLArcTestMember(tc.method, []byte("AAAAAAA"), tc.packed)
			truncated := lhaLArcTestMember(tc.method, []byte("AAAAAAA"), tc.packed[:1])
			expectIncomplete(t, pipelineState(t, clean), truncated, ErrMalformed)
			badCRC := append([]byte(nil), archive...)
			badCRC[31]++
			expectIncomplete(t, pipelineState(t, clean), badCRC, ErrMalformed)
			limited := pipelineState(t, clean)
			limited.maxFile = 6
			expectIncomplete(t, limited, archive, ErrLimit)
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := lzsExpand(ctx, []byte{0, 0, 0}, 8192); !errors.Is(err, context.Canceled) {
		t.Fatalf("LZS cancellation: %v", err)
	}
	if _, err := lz5Expand(ctx, []byte{0, 0, 0}, 8192); !errors.Is(err, context.Canceled) {
		t.Fatalf("LZ5 cancellation: %v", err)
	}
}

func lhaStoredMember(name string, child []byte) []byte {
	h := make([]byte, 24+len(name))
	h[0] = byte(len(h) - 2)
	copy(h[2:7], "-lh0-")
	binary.LittleEndian.PutUint32(h[7:11], uint32(len(child)))
	binary.LittleEndian.PutUint32(h[11:15], uint32(len(child)))
	h[19] = 0x20
	h[21] = byte(len(name))
	copy(h[22:], name)
	binary.LittleEndian.PutUint16(h[22+len(name):], lhaCRC16(child))
	for _, b := range h[2:] {
		h[1] += b
	}
	return append(h, child...)
}

func lhaCompressedTestMember(method byte, child, packed []byte) []byte {
	h := make([]byte, 24+len("child.txt"))
	h[0] = byte(len(h) - 2)
	copy(h[2:7], []byte{'-', 'l', 'h', method, '-'})
	binary.LittleEndian.PutUint32(h[7:11], uint32(len(packed)))
	binary.LittleEndian.PutUint32(h[11:15], uint32(len(child)))
	h[19] = 0x20
	h[21] = byte(len("child.txt"))
	copy(h[22:], "child.txt")
	binary.LittleEndian.PutUint16(h[22+len("child.txt"):], lhaCRC16(child))
	for _, b := range h[2:] {
		h[1] += b
	}
	return append(append(h, packed...), 0)
}

func lhaLevel1TestMember(child []byte, extended bool) []byte {
	name := "child.txt"
	if extended {
		name = "base.txt"
	}
	var ext []byte
	if extended {
		ext = append(append([]byte{1}, []byte("child.txt")...), 0, 0)
	}
	h := make([]byte, 27+len(name))
	h[0] = byte(len(h) - 2)
	copy(h[2:7], "-lh0-")
	binary.LittleEndian.PutUint32(h[7:11], uint32(len(child)+len(ext)))
	binary.LittleEndian.PutUint32(h[11:15], uint32(len(child)))
	h[19], h[20], h[21] = 0x20, 1, byte(len(name))
	copy(h[22:], name)
	binary.LittleEndian.PutUint16(h[22+len(name):], lhaCRC16(child))
	h[24+len(name)] = 'U'
	binary.LittleEndian.PutUint16(h[len(h)-2:], uint16(len(ext)))
	for _, b := range h[2:] {
		h[1] += b
	}
	return append(append(append([]byte(nil), h...), ext...), append(child, 0)...)
}

func lhaLevel2TestMember(child []byte, extended bool) []byte {
	var ext []byte
	if extended {
		ext = append(append([]byte{1}, []byte("child.txt")...), 0, 0)
	}
	h := make([]byte, 26+len(ext))
	binary.LittleEndian.PutUint16(h, uint16(len(h)))
	copy(h[2:7], "-lh0-")
	binary.LittleEndian.PutUint32(h[7:11], uint32(len(child)))
	binary.LittleEndian.PutUint32(h[11:15], uint32(len(child)))
	h[19], h[20] = 0x20, 2
	binary.LittleEndian.PutUint16(h[21:23], lhaCRC16(child))
	h[23] = 'U'
	binary.LittleEndian.PutUint16(h[24:26], uint16(len(ext)))
	copy(h[26:], ext)
	return append(append([]byte(nil), h...), append(child, 0)...)
}

func lhaChainedTestMember(child []byte, level byte) []byte {
	second := []byte{2, 'D', 'I', 'R', 0xff, 0, 0}
	if level == 1 {
		b := lhaLevel1TestMember(child, true)
		baseEnd := 27 + len("base.txt")
		firstEnd := baseEnd + 1 + len("child.txt") + 2
		binary.LittleEndian.PutUint16(b[firstEnd-2:firstEnd], uint16(len(second)))
		b = append(append(append([]byte(nil), b[:firstEnd]...), second...), b[firstEnd:]...)
		binary.LittleEndian.PutUint32(b[7:11], uint32(len(child)+firstEnd-baseEnd+len(second)))
		b[1] = 0
		for _, c := range b[2:baseEnd] {
			b[1] += c
		}
		return b
	}
	b := lhaLevel2TestMember(child, true)
	firstEnd := 26 + 1 + len("child.txt") + 2
	binary.LittleEndian.PutUint16(b[firstEnd-2:firstEnd], uint16(len(second)))
	b = append(append(append([]byte(nil), b[:firstEnd]...), second...), b[firstEnd:]...)
	binary.LittleEndian.PutUint16(b[:2], uint16(firstEnd+len(second)))
	return b
}

func TestLHALevelOneTwoChildHash(t *testing.T) {
	for _, tc := range []struct {
		name     string
		child    []byte
		make     func([]byte, bool) []byte
		extended bool
		sha      string
	}{
		{"level1", []byte("harmless level one two LHA child\n"), lhaLevel1TestMember, false, "461c33d0e3b9c14a467a38fd4f031d156df6bd58c0633f8c808e77e868989a0f"},
		{"level2", []byte("harmless level one two LHA child\n"), lhaLevel2TestMember, false, "c93a7a3d32b331d5a8fbd5a3a4823297f92ffbcf82333c206d00bb5ca9144fac"},
		{"level1-extended", []byte("harmless extended LHA child text\n"), lhaLevel1TestMember, true, "5e2b3866bcabf8224f2afca7d0cf8f1f1cfd4cf7ccffe246c1cd08e1e16a97a5"},
		{"level2-extended", []byte("harmless extended LHA child text\n"), lhaLevel2TestMember, true, "f27778438b02b9c5a76018166e64cf1dcd42bf3344a10aa673c7b34ed9cfba3b"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			archive := tc.make(tc.child, tc.extended)
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.sha {
				t.Fatalf("reference fixture changed: %s", got)
			}
			sum := md5.Sum(tc.child)
			rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:LHA.Level\n", sum, len(tc.child))}
			expectDetected(t, pipelineState(t, rules), archive, "LHA.Level")
		})
	}
	child := []byte("harmless chained LHA child\n")
	for _, tc := range []struct {
		level byte
		sha   string
	}{
		{1, "7efceb1c1fefd872e3d3f15204296b79c894d956b3529ad17f854f699bbcf13d"},
		{2, "3abef38b93e5f8538f66b18aa32e07b6c12d31824fb04435b281426dbc298982"},
	} {
		t.Run(fmt.Sprintf("level%d-chain", tc.level), func(t *testing.T) {
			archive := lhaChainedTestMember(child, tc.level)
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.sha {
				t.Fatalf("reference fixture changed: %s", got)
			}
			sum := md5.Sum(child)
			rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:LHA.Chain\n", sum, len(child))}
			expectDetected(t, pipelineState(t, rules), archive, "LHA.Chain")
		})
	}
}

func TestLHALevelOneTwoDamagedExtendedHeaders(t *testing.T) {
	child := []byte("harmless extended LHA child text\n")
	cleanRules := map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}
	v1 := lhaLevel1TestMember(child, true)
	baseEnd := 27 + len("base.txt")
	badSkip := append([]byte(nil), v1...)
	binary.LittleEndian.PutUint32(badSkip[7:11], 1)
	expectIncomplete(t, pipelineState(t, cleanRules), badSkip, ErrMalformed)
	badExt := append([]byte(nil), v1...)
	binary.LittleEndian.PutUint16(badExt[baseEnd-2:baseEnd], 2)
	expectIncomplete(t, pipelineState(t, cleanRules), badExt, ErrMalformed)
	v2 := lhaLevel2TestMember(child, true)
	badTotal := append([]byte(nil), v2...)
	binary.LittleEndian.PutUint16(badTotal[:2], uint16(len(v2)+10))
	expectIncomplete(t, pipelineState(t, cleanRules), badTotal, ErrMalformed)
	badExt = append([]byte(nil), v2...)
	binary.LittleEndian.PutUint16(badExt[24:26], uint16(len(v2)))
	expectIncomplete(t, pipelineState(t, cleanRules), badExt, ErrMalformed)
}

func TestLHALevelThreeDamagedHeaderRemainsIncomplete(t *testing.T) {
	b := make([]byte, 32)
	binary.LittleEndian.PutUint16(b[:2], 4)
	copy(b[2:7], "-lh0-")
	b[20] = 3
	cleanRules := map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}
	expectIncomplete(t, pipelineState(t, cleanRules), b, ErrMalformed)
	child := []byte("harmless level three child")
	good := lhaLevel3TestMember("-lh0-", child, child, 1)
	badLength := append([]byte(nil), good...)
	binary.LittleEndian.PutUint32(badLength[24:28], uint32(len(good)+8))
	expectIncomplete(t, pipelineState(t, cleanRules), badLength, ErrMalformed)
	badRecord := append([]byte(nil), good...)
	binary.LittleEndian.PutUint32(badRecord[28:32], 4)
	expectIncomplete(t, pipelineState(t, cleanRules), badRecord, ErrMalformed)
	badRecord = append([]byte(nil), good...)
	binary.LittleEndian.PutUint32(badRecord[28:32], uint32(len(good)))
	expectIncomplete(t, pipelineState(t, cleanRules), badRecord, ErrMalformed)
}

func TestLHALevelOneTwoCompressedExtendedChild(t *testing.T) {
	child := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	packed := []byte{0x00, 0x40, 0x00, 0x00, 0x04, 0x10, 0x00}
	sum := md5.Sum(child)
	rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:LHA.LevelCompressed\n", sum, len(child))}
	for _, tc := range []struct {
		level byte
		sha   string
	}{
		{1, "e3b50fde7aa7bebd4d6beec20fa0d860dcd9864f2d6fbfd412f249dbcb20e1dc"},
		{2, "f74f90577af25e92441bdf5726157dbb0b367e46a28a12eddc0a2af90cb5b905"},
	} {
		t.Run(fmt.Sprintf("level%d", tc.level), func(t *testing.T) {
			var archive []byte
			if tc.level == 1 {
				archive = lhaLevel1TestMember(child, true)
				start := 27 + len("base.txt") + 1 + len("child.txt") + 2
				archive = append(append(append([]byte(nil), archive[:start]...), packed...), 0)
				binary.LittleEndian.PutUint32(archive[7:11], uint32(len(packed)+start-(27+len("base.txt"))))
				archive[1] = 0
				for _, c := range archive[2 : 27+len("base.txt")] {
					archive[1] += c
				}
			} else {
				archive = lhaLevel2TestMember(child, true)
				start := 26 + 1 + len("child.txt") + 2
				archive = append(append(append([]byte(nil), archive[:start]...), packed...), 0)
				binary.LittleEndian.PutUint32(archive[7:11], uint32(len(packed)))
			}
			copy(archive[2:7], "-lh5-")
			if tc.level == 1 {
				archive[1] = 0
				for _, c := range archive[2 : 27+len("base.txt")] {
					archive[1] += c
				}
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.sha {
				t.Fatalf("reference fixture changed: %s", got)
			}
			expectDetected(t, pipelineState(t, rules), archive, "LHA.LevelCompressed")
		})
	}
}

func TestLHAStaticHuffmanChildHashes(t *testing.T) {
	child := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	// One 64-symbol block with single-symbol C, PT and P tables. No child
	// bytes occur in the archive. The independent ClamAV 1.5.4 oracle checks
	// all four fixtures by whole-child HDB rules.
	packed := []byte{0x00, 0x40, 0x00, 0x00, 0x04, 0x10, 0x00}
	sum := md5.Sum(child)
	rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:LHA.Child\n", sum, len(child))}
	for _, tc := range []struct {
		method byte
		sha    string
	}{
		{'4', "4979026968a8e15ceb4c82fae7019b885d76fbc8ddad8e32b5b63599a6eef715"},
		{'5', "e959006d643afb93559a708f6cbf99f9c3674e0b18c442d91f127e9dac54b1b5"},
		{'6', "b8936f14e157f722962ff03473442f65603762cb0d6ea970a7da2d770809f7bc"},
		{'7', "23dded614631d120b30b90a12b6719064c31b2a307b2221a0f5fc40e80fb5b66"},
	} {
		t.Run(string(tc.method), func(t *testing.T) {
			archive := lhaCompressedTestMember(tc.method, child, packed)
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.sha {
				t.Fatalf("reference LHA fixture changed: %s", got)
			}
			expectDetected(t, pipelineState(t, rules), archive, "LHA.Child")
		})
	}
}

func TestLH5BackReferenceChildHash(t *testing.T) {
	// Block 1 emits one literal A; block 2 emits two length-three matches at
	// distance zero. The resulting seven bytes are not present in the input.
	child := []byte("AAAAAAA")
	packed := []byte{0x00, 0x01, 0x00, 0x00, 0x04, 0x10, 0x00, 0x00, 0x20, 0x00, 0x01, 0x00, 0x00}
	archive := lhaCompressedTestMember('5', child, packed)
	if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != "bd5808a874c64458d1b6e2904c227ceb9b9ecbd7e1a3295f909ce471f9d4ab09" {
		t.Fatalf("reference LH5 fixture changed: %s", got)
	}
	sum := md5.Sum(child)
	rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:LHA.Match\n", sum, len(child))}
	expectDetected(t, pipelineState(t, rules), archive, "LHA.Match")
}

func TestLH5TwoSymbolHuffmanChildHash(t *testing.T) {
	// Independent C table with A and B as one-bit symbols, exercised by the
	// alternating body. The archive does not contain the eight-byte child.
	child := []byte("ABABABAB")
	packed := []byte{0x00, 0x08, 0x21, 0x00, 0x24, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0xaa}
	archive := lhaCompressedTestMember('5', child, packed)
	if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != "e1e8dd86a646e97261306045461d2b8eb2a163fdf8508a393c507f1df3edd209" {
		t.Fatalf("reference LH5 two-symbol fixture changed: %s", got)
	}
	sum := md5.Sum(child)
	rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:LHA.Mixed\n", sum, len(child))}
	expectDetected(t, pipelineState(t, rules), archive, "LHA.Mixed")
}

func TestLH5ExpansionHonorsCancellation(t *testing.T) {
	// A single-symbol block can expand thousands of bytes from seven input
	// bytes; cancellation must interrupt that loop before its declared end.
	packed := []byte{0x20, 0x00, 0x00, 0x00, 0x04, 0x10, 0x00}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d := &arjDecoder{arjBits: arjBits{in: packed}, ctx: ctx, dictSize: 1 << 13, np: 14, pBit: 4}
	out, err := d.decode(8192)
	if !errors.Is(err, context.Canceled) || len(out) >= 8192 {
		t.Fatalf("decoded=%d err=%v", len(out), err)
	}
}

func TestLH5DamagedOrOverBudgetRemainsIncomplete(t *testing.T) {
	child := []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	packed := []byte{0x00, 0x40, 0x00, 0x00, 0x04, 0x10, 0x00}
	archive := lhaCompressedTestMember('5', child, packed)
	cleanRules := map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}
	expectIncomplete(t, pipelineState(t, cleanRules), archive[:len(archive)-2], ErrMalformed)
	badCRC := append([]byte(nil), archive...)
	badCRC[22+len("child.txt")]++
	expectIncomplete(t, pipelineState(t, cleanRules), badCRC, ErrMalformed)
	limited := pipelineState(t, cleanRules)
	limited.maxFile = 32
	expectIncomplete(t, limited, archive, ErrLimit)
	// The decoded child is still scanned before a late CRC error is reported.
	sum := md5.Sum(child)
	rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:LHA.Child\n", sum, len(child))}
	expectDetected(t, pipelineState(t, rules), badCRC, "LHA.Child")
}

func TestLHAStoredChildHashAndLaterMember(t *testing.T) {
	child := []byte("harmless lha child payload 0123456789\n")
	archive := append(lhaStoredMember("first.txt", []byte("harmless first member data")), lhaStoredMember("child.txt", child)...)
	archive = append(archive, 0)
	if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != "b5fa3da7bbaa9cd3c90cdb44876081e533d085b3e2839b3e2f1a15c1cd0e6cf0" {
		t.Fatalf("reference LHA fixture changed: %s", got)
	}
	if !lhaHeader(archive) {
		t.Fatal("LHA header unrecognized")
	}
	sum := md5.Sum(child)
	rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:LHA.Child\n", sum, len(child))}
	expectDetected(t, pipelineState(t, rules), archive, "LHA.Child")
	expectDetected(t, pipelineState(t, rules), lhaStoredMember("child.txt", child), "LHA.Child")
	expectScan(t, pipelineState(t, rules), lhaStoredMember("other.txt", []byte("different harmless child data")), "", "")
}

func TestLHAStoredMalformedAndUnsupported(t *testing.T) {
	child := []byte("harmless lha child payload 0123456789\n")
	base := lhaStoredMember("child.txt", child)
	cleanRules := map[string]string{"test.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}
	t.Run("truncated", func(t *testing.T) {
		expectIncomplete(t, pipelineState(t, cleanRules), base[:len(base)-2], ErrMalformed)
	})
	t.Run("checksum", func(t *testing.T) {
		b := append([]byte(nil), base...)
		b[1]++
		expectIncomplete(t, pipelineState(t, cleanRules), b, ErrMalformed)
	})
	t.Run("crc", func(t *testing.T) {
		b := append([]byte(nil), base...)
		b[22+len("child.txt")]++
		expectIncomplete(t, pipelineState(t, cleanRules), b, ErrMalformed)
	})
	t.Run("compressed method before readable child", func(t *testing.T) {
		b := append([]byte(nil), base...)
		copy(b[2:7], "-lh2-")
		b[1] = 0
		for _, c := range b[2 : 24+len("child.txt")] {
			b[1] += c
		}
		b = append(b, base...)
		if sig, err := pipelineState(t, cleanRules).visit(b, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("sig=%q err=%v", sig, err)
		}
		sum := md5.Sum(child)
		rules := map[string]string{"test.hdb": fmt.Sprintf("%x:%d:LHA.Child\n", sum, len(child))}
		if sig, err := pipelineState(t, rules).visit(b, 0); sig != "LHA.Child" || err != nil {
			t.Fatalf("later child: sig=%q err=%v", sig, err)
		}
	})
}
