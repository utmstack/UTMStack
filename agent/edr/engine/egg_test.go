package engine

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
)

const eggOracle = "../../../docs/edr/native-engine/audits/2026-10-28-v143-egg/"
const eggCommentOracle = "../../../docs/edr/native-engine/audits/2026-10-28-v144-egg-comments/"

func TestEGGReferenceComments(t *testing.T) {
	b, err := os.ReadFile(eggCommentOracle + "reference-verdicts.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Fixture    string `json:"fixture"`
		ArchiveSHA string `json:"archive_sha256"`
		CommentMD5 string `json:"comment_md5"`
		Size       int    `json:"comment_size"`
		Detect     bool   `json:"detect"`
	}
	if err := json.Unmarshal(b, &cases); err != nil || len(cases) != 8 {
		t.Fatalf("comment receipt: %d cases, %v", len(cases), err)
	}
	for _, tc := range cases {
		t.Run(tc.Fixture, func(t *testing.T) {
			archive, err := os.ReadFile(eggCommentOracle + "oracle/" + tc.Fixture)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.ArchiveSHA {
				t.Fatalf("EGG comment oracle changed: %s", got)
			}
			rule := map[string]string{"comment.hdb": fmt.Sprintf("%s:%d:Reference.EGG.Comment\n", tc.CommentMD5, tc.Size)}
			if tc.Detect {
				expectDetected(t, pipelineState(t, rule), archive, "Reference.EGG.Comment")
			} else {
				expectScan(t, pipelineState(t, rule), archive, "", "")
			}
		})
	}
}

func TestEGGReferenceChildHashes(t *testing.T) {
	b, err := os.ReadFile(eggOracle + "reference-verdicts.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Fixture    string `json:"fixture"`
		ArchiveSHA string `json:"archive_sha256"`
		ChildMD5   string `json:"child_md5"`
		ChildSize  int    `json:"child_size"`
	}
	if err := json.Unmarshal(b, &cases); err != nil || len(cases) != 7 {
		t.Fatalf("reference receipt: %d cases, %v", len(cases), err)
	}
	for _, tc := range cases {
		t.Run(tc.Fixture, func(t *testing.T) {
			archive, err := os.ReadFile(eggOracle + "oracle/" + tc.Fixture)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.ArchiveSHA {
				t.Fatalf("EGG oracle changed: %s", got)
			}
			rule := map[string]string{"child.hdb": fmt.Sprintf("%s:%d:Reference.EGG.Child\n", tc.ChildMD5, tc.ChildSize)}
			expectDetected(t, pipelineState(t, rule), archive, "Reference.EGG.Child")
			clean := map[string]string{"other.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}
			if tc.Fixture == "encrypted-before-stored.egg" || tc.Fixture == "unsupported-before-stored.egg" {
				expectIncomplete(t, pipelineState(t, clean), archive, ErrUnsupported)
			} else {
				expectScan(t, pipelineState(t, clean), archive, "", "")
			}
		})
	}
}

func TestEGGContainerMetadataAndFilenameFlag(t *testing.T) {
	archive, err := os.ReadFile(eggOracle + "oracle/encrypted-before-stored.egg")
	if err != nil {
		t.Fatal(err)
	}
	rules := map[string]string{"egg.cdb": "EGG.Secret:CL_TYPE_EGG:*:^secret\\.txt$:*:*:0:1:*:\n"}
	expectDetected(t, pipelineState(t, rules), archive, "EGG.Secret")
	rules = map[string]string{"egg.cdb": "EGG.Child:CL_TYPE_EGG:*:^child\\.txt$:*:*:0:2:*:\n"}
	expectDetected(t, pipelineState(t, rules), archive, "EGG.Child")
	// The filename's bit 2 does not encrypt the file contents. The reference
	// scans the first member even though its metadata encryption field is 0.
	digest := md5.Sum([]byte("opaque"))
	expectDetected(t, pipelineState(t, map[string]string{"secret.hdb": fmt.Sprintf("%x:6:EGG.Opaque\n", digest)}), archive, "EGG.Opaque")
}

func TestEGGMalformedAndLimited(t *testing.T) {
	archive, err := os.ReadFile(eggOracle + "oracle/stored.egg")
	if err != nil {
		t.Fatal(err)
	}
	truncated := append([]byte(nil), archive...)
	truncated[64] = 250 // packed size exceeds archive bounds
	_, _, err = eggArchiveOf(truncated, 10, nil)
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("oversized block: %v", err)
	}
	if _, _, err = eggArchiveOf(archive, 0, nil); !errors.Is(err, ErrLimit) {
		t.Fatalf("file-count limit: %v", err)
	}
	clean := pipelineState(t, map[string]string{"other.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"})
	clean.maxFile = 50
	if sig, err := clean.egg(archive, 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("decoded child limit: sig=%q err=%v", sig, err)
	}
	var empty [22]byte
	binary.LittleEndian.PutUint32(empty[:], eggBlockMagic)
	binary.LittleEndian.PutUint32(empty[18:], eggTerminator)
	manyBlocks := append([]byte(nil), archive[:54]...)
	manyBlocks = append(manyBlocks, bytes.Repeat(empty[:], 65537)...)
	manyBlocks = append(manyBlocks, archive[len(archive)-4:]...)
	if _, _, err = eggArchiveOf(manyBlocks, 10, nil); !errors.Is(err, ErrLimit) {
		t.Fatalf("block-count limit: %v", err)
	}
}

func TestEGGIntegrityMismatchScansChildButCannotReturnClean(t *testing.T) {
	archive, err := os.ReadFile(eggOracle + "oracle/deflate.egg")
	if err != nil {
		t.Fatal(err)
	}
	fileAt := bytes.Index(archive, []byte{0xe3, 0x90, 0x85, 0x0a})
	blockAt := bytes.Index(archive, []byte{0x13, 0x0c, 0xb5, 0x02})
	if fileAt < 0 || blockAt < 0 {
		t.Fatal("pinned EGG headers missing")
	}
	child := bytes.Repeat([]byte("EGG_GO_CHILD_ONLY_VALIDATION_2026_"), 3)
	positive := map[string]string{"child.hdb": fmt.Sprintf("%x:%d:Integrity.Child\n", md5.Sum(child), len(child))}
	negative := map[string]string{"other.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}
	cases := []struct {
		name   string
		change func([]byte)
	}{
		{"member-length", func(b []byte) { binary.LittleEndian.PutUint64(b[fileAt+8:], uint64(len(child)+1)) }},
		{"block-length", func(b []byte) { binary.LittleEndian.PutUint32(b[blockAt+6:], uint32(len(child)+1)) }},
		{"block-crc", func(b []byte) { b[blockAt+14] ^= 0x80 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			broken := append([]byte(nil), archive...)
			tc.change(broken)
			expectDetected(t, pipelineState(t, positive), broken, "Integrity.Child")
			expectIncomplete(t, pipelineState(t, negative), broken, ErrMalformed)
		})
	}
}
