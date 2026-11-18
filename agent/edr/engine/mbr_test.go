package engine

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func mbrTestImage() ([]byte, []byte) {
	const sector = 512
	part := make([]byte, 2*sector)
	copy(part[251:], "UTMSTACK-MBR-PARTITION-CONTENT")
	img := make([]byte, 4*sector)
	img[446+4] = 0x83
	binary.LittleEndian.PutUint32(img[446+8:], 1)
	binary.LittleEndian.PutUint32(img[446+12:], 2)
	img[510], img[511] = 0x55, 0xaa
	copy(img[sector:], part)
	return img, part
}

func mbrExtendedTestImage() ([]byte, []byte) {
	const sector = 512
	part := make([]byte, 2*sector)
	copy(part[219:], "UTMSTACK-EXTENDED-PARTITION-ABC")
	img := make([]byte, 8*sector)
	img[446+4] = 0x05
	binary.LittleEndian.PutUint32(img[446+8:], 1)
	binary.LittleEndian.PutUint32(img[446+12:], 6)
	img[510], img[511] = 0x55, 0xaa
	ebr := img[sector : 2*sector]
	ebr[446+4] = 0x83
	binary.LittleEndian.PutUint32(ebr[446+8:], 1)
	binary.LittleEndian.PutUint32(ebr[446+12:], 2)
	ebr[510], ebr[511] = 0x55, 0xaa
	copy(img[2*sector:], part)
	return img, part
}

func mbrLinkedTestImage() ([]byte, []byte) {
	const sector = 512
	img := make([]byte, 10*sector)
	img[446+4] = 0x05
	binary.LittleEndian.PutUint32(img[446+8:], 1)
	binary.LittleEndian.PutUint32(img[446+12:], 8)
	img[510], img[511] = 0x55, 0xaa
	first := img[sector : 2*sector]
	first[446+4] = 0x83
	binary.LittleEndian.PutUint32(first[446+8:], 1)
	binary.LittleEndian.PutUint32(first[446+12:], 1)
	first[462+4] = 0x05
	binary.LittleEndian.PutUint32(first[462+8:], 3)
	binary.LittleEndian.PutUint32(first[462+12:], 4)
	first[510], first[511] = 0x55, 0xaa
	second := img[4*sector : 5*sector]
	second[446+4] = 0x83
	binary.LittleEndian.PutUint32(second[446+8:], 1)
	binary.LittleEndian.PutUint32(second[446+12:], 2)
	second[510], second[511] = 0x55, 0xaa
	part := img[5*sector : 7*sector]
	copy(part[267:], "UTMSTACK-SECOND-LOGICAL-PARTITION")
	return img, part
}

func TestMBRPrimaryPartitionHash(t *testing.T) {
	img, part := mbrTestImage()
	if got := fmt.Sprintf("%x", sha256.Sum256(img)); got != "8bd28714a01b925c9af3ed2524e35041cca1993f9156d43e61787874007f5fa0" {
		t.Fatalf("oracle image changed: %s", got)
	}
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:MBR.Child\n", sum, len(part)),
	}
	expectDetected(t, pipelineState(t, files), img, "MBR.Child")
	control := append([]byte(nil), img...)
	control[512+251] ^= 0x01
	expectScan(t, pipelineState(t, files), control, "", "")
	if dir := os.Getenv("EDR_MBR_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "primary.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMBRExtendedPartitionHash(t *testing.T) {
	img, part := mbrExtendedTestImage()
	if got := fmt.Sprintf("%x", sha256.Sum256(img)); got != "8e2244120566210195ed5a14d43a437c4e9d2455573c6986642e9e375cca0183" {
		t.Fatalf("oracle image changed: %s", got)
	}
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:MBR.Logical\n", sum, len(part)),
	}
	expectDetected(t, pipelineState(t, files), img, "MBR.Logical")
	control := append([]byte(nil), img...)
	control[2*512+219] ^= 0x01
	expectScan(t, pipelineState(t, files), control, "", "")
	if dir := os.Getenv("EDR_MBR_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "extended.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMBRLinkedExtendedPartitionHash(t *testing.T) {
	img, part := mbrLinkedTestImage()
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:MBR.Linked\n", sum, len(part)),
	}
	expectDetected(t, pipelineState(t, files), img, "MBR.Linked")
	control := append([]byte(nil), img...)
	control[5*512+267] ^= 0x01
	expectScan(t, pipelineState(t, files), control, "", "")
	if dir := os.Getenv("EDR_MBR_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "linked.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMBRBootSectorHashAndBounds(t *testing.T) {
	img, _ := mbrTestImage()
	boot := md5.Sum(img[:512])
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:512:MBR.Boot\n", boot),
	}
	expectDetected(t, pipelineState(t, files), img, "MBR.Boot")
	bad := append([]byte(nil), img...)
	binary.LittleEndian.PutUint32(bad[446+12:], 8) // past the image
	if sig, err := pipelineState(t, files).visit(bad, 0); sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("out-of-range partition: sig=%q err=%v", sig, err)
	}
}

func TestMBRLaterPartitionDetectsAfterUnsupportedPartition(t *testing.T) {
	img := make([]byte, 4*512)
	img[446+4] = 0x83
	binary.LittleEndian.PutUint32(img[446+8:], 1)
	binary.LittleEndian.PutUint32(img[446+12:], 1)
	img[462+4] = 0x83
	binary.LittleEndian.PutUint32(img[462+8:], 2)
	binary.LittleEndian.PutUint32(img[462+12:], 1)
	img[510], img[511] = 0x55, 0xaa
	copy(img[512:], "H+")
	copy(img[2*512+127:], "UTMSTACK-LATER-MBR-PARTITION")
	part := img[2*512 : 3*512]
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n" +
			"4:0:482b:HFS+ partition:CL_TYPE_PART_ANY:CL_TYPE_PART_HFSPLUS\n",
		"probe.hdb": fmt.Sprintf("%x:%d:MBR.Later\n", sum, len(part)),
	}
	expectDetected(t, pipelineState(t, files), img, "MBR.Later")
	withoutLater := append([]byte(nil), img...)
	withoutLater[462+4] = 0
	if sig, err := pipelineState(t, files).visit(withoutLater, 0); sig != "" || unknownReasonOf(err) != "format-part_hfsplus" {
		t.Fatalf("specialized partition typing: sig=%q err=%v", sig, err)
	}
}

func TestMBRLogicalPartitionInSecondEBREntry(t *testing.T) {
	img, part := mbrExtendedTestImage()
	img = append([]byte(nil), img...)
	ebr := img[512:1024]
	copy(ebr[462:478], ebr[446:462])
	clear(ebr[446:462])
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:MBR.SecondEntry\n", sum, len(part)),
	}
	if dir := os.Getenv("EDR_MBR_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "second-entry.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	expectDetected(t, pipelineState(t, files), img, "MBR.SecondEntry")
}

func TestMBRTwoLogicalEntries(t *testing.T) {
	img, _ := mbrExtendedTestImage()
	img = append([]byte(nil), img...)
	ebr := img[512:1024]
	binary.LittleEndian.PutUint32(ebr[446+12:], 1)
	ebr[462+4] = 0x83
	binary.LittleEndian.PutUint32(ebr[462+8:], 3)
	binary.LittleEndian.PutUint32(ebr[462+12:], 2)
	part := img[4*512 : 6*512]
	copy(part[201:], "UTMSTACK-SECOND-EBR-PARTITION")
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:MBR.TwoEntries\n", sum, len(part)),
	}
	if dir := os.Getenv("EDR_MBR_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "two-entries.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	expectDetected(t, pipelineState(t, files), img, "MBR.TwoEntries")
}

func TestMBRLinkBeforeLogicalEntry(t *testing.T) {
	img := make([]byte, 10*512)
	img[446+4] = 0x05
	binary.LittleEndian.PutUint32(img[446+8:], 1)
	binary.LittleEndian.PutUint32(img[446+12:], 8)
	img[510], img[511] = 0x55, 0xaa
	ebr := img[512:1024]
	ebr[446+4] = 0x05
	binary.LittleEndian.PutUint32(ebr[446+8:], 3)
	binary.LittleEndian.PutUint32(ebr[446+12:], 4)
	ebr[462+4] = 0x83
	binary.LittleEndian.PutUint32(ebr[462+8:], 1)
	binary.LittleEndian.PutUint32(ebr[462+12:], 2)
	ebr[510], ebr[511] = 0x55, 0xaa
	part := img[5*512 : 7*512]
	copy(part[179:], "UTMSTACK-LINK-BEFORE-PARTITION")
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:MBR.LinkFirst\n", sum, len(part)),
	}
	if dir := os.Getenv("EDR_MBR_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "link-first.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	expectDetected(t, pipelineState(t, files), img, "MBR.LinkFirst")
}

func TestMBREmptyEntryBeforeLink(t *testing.T) {
	img, part := mbrLinkedTestImage()
	img = append([]byte(nil), img...)
	clear(img[512+446 : 512+462])
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:MBR.EmptyBeforeLink\n", sum, len(part)),
	}
	if dir := os.Getenv("EDR_MBR_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "empty-before-link.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	expectDetected(t, pipelineState(t, files), img, "MBR.EmptyBeforeLink")
}
