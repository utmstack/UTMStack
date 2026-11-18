package engine

import (
	"bytes"
	"crypto/md5"
	"crypto/rc4"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/crypto/tea"
)

// The official unpacker recognizes an eight-byte stage prefix. Everything
// after that prefix is repeated harmless text. These bytes are scanned only;
// neither the parent PE nor the child is ever executed.
func officialGandCrabControl(t *testing.T, cipher, keySection string) (parent, child []byte) {
	t.Helper()
	const size = 0x10108 // satisfies the official resource signature's size pattern
	marker := []byte("OPENEDR-INERT-GANDCRAB-RESOURCE-")
	child = append([]byte{0xeb, 3, 0xc2, 0x0c, 0, 0x55, 0x8b, 0xec}, bytes.Repeat(marker, size/len(marker)+1)...)
	child = child[:size]
	key := bytes.Repeat([]byte{0x81, 0x42, 0x43}, 10) // 30 nonzero bytes, including non-ASCII
	encoded := bytes.Clone(child)
	le := binary.LittleEndian
	switch cipher {
	case "rc4-reverse":
		r, err := rc4.NewCipher(key)
		if err != nil {
			t.Fatal(err)
		}
		slices.Reverse(encoded)
		r.XORKeyStream(encoded, encoded)
		slices.Reverse(encoded)
	case "lcg":
		state := le.Uint32(key)
		for i := range encoded {
			state = state*214013 + 2531011
			encoded[i] ^= byte(state >> 16)
		}
	case "tea":
		// x/crypto's independent TEA implementation uses big-endian words;
		// this resource format stores the same key and block words little-endian.
		be := binary.BigEndian
		var bigKey [16]byte
		for i := 0; i < len(bigKey); i += 4 {
			be.PutUint32(bigKey[i:], le.Uint32(key[i:]))
		}
		block, err := tea.NewCipher(bigKey[:])
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < len(encoded); i += 8 {
			var big [8]byte
			be.PutUint32(big[:], le.Uint32(child[i:]))
			be.PutUint32(big[4:], le.Uint32(child[i+4:]))
			block.Encrypt(big[:], big[:])
			le.PutUint32(encoded[i:], be.Uint32(big[:]))
			le.PutUint32(encoded[i+4:], be.Uint32(big[4:]))
		}
	default:
		t.Fatalf("unknown test cipher %q", cipher)
	}
	data := make([]byte, 512)
	if keySection == ".data" {
		copy(data[16:], key)
	}
	rsrc := make([]byte, (0x200+size+511)&^511)
	le.PutUint16(rsrc[12:], 1) // one named root entry
	le.PutUint32(rsrc[0x10:], 0x80000100)
	le.PutUint32(rsrc[0x14:], 0x80000030)
	le.PutUint16(rsrc[0x30+14:], 1)
	le.PutUint32(rsrc[0x40:], 0x8c)
	le.PutUint32(rsrc[0x44:], 0x80000060)
	le.PutUint16(rsrc[0x60+14:], 1)
	le.PutUint32(rsrc[0x70:], 0)
	le.PutUint32(rsrc[0x74:], 0x90)
	le.PutUint32(rsrc[0x90:], 0x3200)
	le.PutUint32(rsrc[0x94:], size)
	le.PutUint16(rsrc[0x100:], 4)
	copy(rsrc[0x102:], []byte{'T', 0, 'E', 0, 'S', 0, 'T', 0})
	copy(rsrc[0x200:], encoded)
	f := simplePE(make([]byte, 512))
	f.sections = append(f.sections,
		peFixtureSection{name: ".data", va: 0x2000, vsize: 0x1000, raw: data},
		peFixtureSection{name: ".rsrc", va: 0x3000, vsize: 0x11000, raw: rsrc})
	if keySection == ".rdata" {
		f.headers = 0x400 // four section records do not fit a 512-byte header
		rdata := make([]byte, 512)
		copy(rdata[16:], key)
		f.sections = append(f.sections, peFixtureSection{name: ".rdata", va: 0x14000, vsize: 0x1000, raw: rdata})
	}
	parent = f.build()
	le.PutUint32(parent[264:], 0x3000)
	le.PutUint32(parent[268:], uint32(len(rsrc)))
	if bytes.Contains(parent, marker) {
		t.Fatal("decoded marker occurs in parent")
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(child)); got != "48108a18368fd260d67fe98cd052fb81c332e2e8c2c46e1682460c632e69196a" {
		t.Fatalf("inert GandCrab child SHA-256 = %s", got)
	}
	parentHashes := map[string]string{
		"rc4-reverse.data":  "5518702e290c8c56228fccb2be452bb35f27d004f85fbaf5d6295c4214210d06",
		"rc4-reverse.rdata": "d99bbdf689272731d27c245c11204b0752f2cfefa338dde9f98288227ccb1b71",
		"lcg.data":          "005b6a3027b430478b04fe65a0cb472df476d2af07f1ecc793f9592087468855",
		"lcg.rdata":         "9c546b342b4d7345cac1d414eadaca8506f40648e359df7cc25353ecec675007",
		"tea.data":          "52801a70875651006346b17d2573b4dc47ec64901f20aadd633c59837e2e6dc4",
		"tea.rdata":         "ac73278d8dd5498bffb513dadd76c8c6e3f5e984309daca83cb7bd228c87eac5",
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(parent)); got != parentHashes[cipher+keySection] {
		t.Fatalf("inert GandCrab parent SHA-256 = %s", got)
	}
	return parent, child
}

func TestOfficialGandCrabExtractedChild(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_BYTECODE_DIR")
	if dir == "" {
		t.Skip("set EDR_OFFICIAL_BYTECODE_DIR to expanded authenticated bytecode.cvd members")
	}
	program, err := os.ReadFile(filepath.Join(dir, "bytecode-6539706.cbc"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(program)); got != "f93354db8e201d386a1cd05b92ffe37116885fd7da9cd09f27f845754ba10db3" {
		t.Fatalf("official CBC SHA-256 = %s", got)
	}
	for _, cipher := range []string{"rc4-reverse", "lcg", "tea"} {
		for _, keySection := range []string{".data", ".rdata"} {
			t.Run(cipher+keySection, func(t *testing.T) {
				parent, child := officialGandCrabControl(t, cipher, keySection)
				rule := fmt.Sprintf("%x:%d:Child.Official.GandCrab\n", md5.Sum(child), len(child))
				s := pipelineState(t, map[string]string{"hook.cbc": string(program), "child.hdb": rule})
				if sig, err := s.visit(parent, 0); sig != "Child.Official.GandCrab" || err != nil {
					t.Fatalf("official CBC child: signature=%q error=%v", sig, err)
				}
				without := pipelineState(t, map[string]string{"child.hdb": rule})
				if sig, err := without.visit(parent, 0); sig != "" || err != nil {
					t.Fatalf("parent without CBC: signature=%q error=%v", sig, err)
				}
				broken := bytes.Clone(parent)
				at := int(binary.LittleEndian.Uint32(parent[0x178+2*40+20:])) + 0x200
				broken[at] ^= 1 // corrupt the encrypted prefix while retaining the trigger
				if sig, err := s.visit(broken, 0); sig != "" || err != nil {
					t.Fatalf("corrupt resource: signature=%q error=%v", sig, err)
				}
				t.Logf("parent bytes=%d SHA256=%x; child bytes=%d SHA256=%x", len(parent), sha256.Sum256(parent), len(child), sha256.Sum256(child))
				referenceGandCrabOutput(t, program, parent, child, rule)
			})
		}
	}
}

func referenceGandCrabOutput(t *testing.T, program, parent, child []byte, rule string) {
	t.Helper()
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	clamscan := filepath.Join(ref, "bin", "clamscan")
	tool, err := os.ReadFile(clamscan)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(tool)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("unpinned ClamAV 1.5.4 scanner: %v", err)
	}
	dir := t.TempDir()
	for name, data := range map[string][]byte{"hook.cbc": program, "child.hdb": []byte(rule), "parent.exe": parent} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	broken := bytes.Clone(parent)
	at := int(binary.LittleEndian.Uint32(parent[0x178+2*40+20:])) + 0x200
	broken[at] ^= 1
	if err := os.WriteFile(filepath.Join(dir, "broken.exe"), broken, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(clamscan, "--database="+dir, "--bytecode-unsigned=yes", "--bytecode=yes", "--debug", "--no-summary", filepath.Join(dir, "parent.exe"))
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	status, ok := err.(*exec.ExitError)
	if !ok || status.ExitCode() != 1 || !strings.Contains(string(out), "parent.exe: Child.Official.GandCrab.UNOFFICIAL FOUND") ||
		!strings.Contains(string(out), fmt.Sprintf("previous tempfile had %d bytes", len(child))) ||
		!strings.Contains(string(out), fmt.Sprintf("fmap_get_hash: calculated sha2-256 hash: %x", sha256.Sum256(child))) {
		t.Fatalf("reference GandCrab output: %v: %s", err, out)
	}
	for _, tc := range []struct{ name, bytecode string }{{"broken.exe", "yes"}, {"parent.exe", "no"}} {
		cmd := exec.Command(clamscan, "--database="+dir, "--bytecode-unsigned=yes", "--bytecode="+tc.bytecode, "--no-summary", filepath.Join(dir, tc.name))
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), tc.name+": OK") {
			t.Fatalf("reference control %s bytecode=%s: %v: %s", tc.name, tc.bytecode, err, out)
		}
	}
	t.Log("ClamAV detects the exact extracted child; corrupt-resource and disabled-hook controls are clean")
}
