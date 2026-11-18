package engine

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"testing"
)

func gptTestImage() ([]byte, []byte) { return gptTestImageSized(512) }

func gptTestImageSized(sector int) ([]byte, []byte) {
	const sectors = 64
	img := make([]byte, sector*sectors)
	bootRecord := sector - 66
	img[bootRecord+4] = 0xee
	binary.LittleEndian.PutUint32(img[bootRecord+8:], 1)
	binary.LittleEndian.PutUint32(img[bootRecord+12:], sectors-1)
	img[sector-2], img[sector-1] = 0x55, 0xaa
	if sector != 512 {
		// The reference's type recognizer reads the first KiB and its MBR
		// check assumes 512-byte sectors. Duplicate the protective record
		// there so the test exercises its larger-sector GPT reader.
		copy(img[446:512], img[bootRecord:sector])
	}
	ptable := make([]byte, 4*128)
	ptable[0] = 0xa2 // nonempty partition-type GUID
	ptable[16] = 0x5b
	binary.LittleEndian.PutUint64(ptable[32:], 4)
	binary.LittleEndian.PutUint64(ptable[40:], 5)
	copy(img[2*sector:], ptable)
	copy(img[62*sector:], ptable)
	part := img[4*sector : 6*sector]
	copy(part[233:], "UTMSTACK-GPT-PARTITION-CONTENT")
	writeHeader := func(at, current, backup, table uint64) {
		h := img[int(at)*sector : int(at)*sector+92]
		copy(h, "EFI PART")
		binary.LittleEndian.PutUint32(h[8:], 0x10000)
		binary.LittleEndian.PutUint32(h[12:], 92)
		binary.LittleEndian.PutUint64(h[24:], current)
		binary.LittleEndian.PutUint64(h[32:], backup)
		binary.LittleEndian.PutUint64(h[40:], 3)
		binary.LittleEndian.PutUint64(h[48:], 61)
		h[56] = 0x97
		binary.LittleEndian.PutUint64(h[72:], table)
		binary.LittleEndian.PutUint32(h[80:], 4)
		binary.LittleEndian.PutUint32(h[84:], 128)
		binary.LittleEndian.PutUint32(h[88:], crc32.ChecksumIEEE(ptable))
		binary.LittleEndian.PutUint32(h[16:], crc32.ChecksumIEEE(h))
	}
	writeHeader(1, 1, 63, 2)
	writeHeader(63, 63, 1, 62)
	return img, part
}

func TestGPTLargeSectors(t *testing.T) {
	for _, sector := range []int{1024, 2048, 4096} {
		t.Run(fmt.Sprint(sector), func(t *testing.T) {
			img, part := gptTestImageSized(sector)
			sum := md5.Sum(part)
			files := map[string]string{
				"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
				"probe.hdb": fmt.Sprintf("%x:%d:GPT.LargeSector\n", sum, len(part)),
			}
			if sector == 4096 {
				if dir := os.Getenv("EDR_GPT_FIXTURE_DIR"); dir != "" {
					if err := os.WriteFile(filepath.Join(dir, "sector4096.img"), img, 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			expectDetected(t, pipelineState(t, files), img, "GPT.LargeSector")
		})
	}
}

func TestGPTPartitionHash(t *testing.T) {
	img, part := gptTestImage()
	t.Logf("image SHA256=%x", sha256.Sum256(img))
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:GPT.Child\n", sum, len(part)),
	}
	if dir := os.Getenv("EDR_GPT_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "primary.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	expectDetected(t, pipelineState(t, files), img, "GPT.Child")
	control := append([]byte(nil), img...)
	control[4*512+233] ^= 1
	expectScan(t, pipelineState(t, files), control, "", "")
}

func TestGPTSecondaryFallback(t *testing.T) {
	img, part := gptTestImage()
	img = append([]byte(nil), img...)
	img[512+16] ^= 0x80 // header checksum no longer agrees
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:GPT.Secondary\n", sum, len(part)),
	}
	if dir := os.Getenv("EDR_GPT_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "secondary.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	expectDetected(t, pipelineState(t, files), img, "GPT.Secondary")
	img[63*512+16] ^= 0x80
	if sig, err := pipelineState(t, files).visit(img, 0); sig != "" || err != ErrMalformed {
		t.Fatalf("both headers invalid: sig=%q err=%v", sig, err)
	}
}

func TestGPTPrimaryTableCRCFallback(t *testing.T) {
	img, part := gptTestImage()
	img = append([]byte(nil), img...)
	img[2*512+128] ^= 0x01 // header remains valid, primary table CRC does not
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:GPT.TableFallback\n", sum, len(part)),
	}
	if dir := os.Getenv("EDR_GPT_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "table-fallback.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	expectDetected(t, pipelineState(t, files), img, "GPT.TableFallback")
}

func TestGPTDisagreeingTablesScanBoth(t *testing.T) {
	img, _ := gptTestImage()
	img = append([]byte(nil), img...)
	part := img[6*512 : 8*512]
	copy(part[173:], "UTMSTACK-SECONDARY-GPT-PARTITION")
	secondaryTable := img[62*512 : 63*512]
	binary.LittleEndian.PutUint64(secondaryTable[32:], 6)
	binary.LittleEndian.PutUint64(secondaryTable[40:], 7)
	header := img[63*512 : 63*512+92]
	binary.LittleEndian.PutUint32(header[88:], crc32.ChecksumIEEE(secondaryTable))
	binary.LittleEndian.PutUint32(header[16:], 0)
	binary.LittleEndian.PutUint32(header[16:], crc32.ChecksumIEEE(header))
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:GPT.SecondTable\n", sum, len(part)),
	}
	if dir := os.Getenv("EDR_GPT_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "both.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	expectDetected(t, pipelineState(t, files), img, "GPT.SecondTable")
}

func TestGPTTypedDirectly(t *testing.T) {
	img, part := gptTestImage()
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:512:4546492050415254:GUID Partition Table:CL_TYPE_ANY:CL_TYPE_GPT:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:GPT.Typed\n", sum, len(part)),
	}
	expectDetected(t, pipelineState(t, files), img, "GPT.Typed")
}

func TestGPTProtectiveMBRRangeDoesNotBlockDispatch(t *testing.T) {
	img, part := gptTestImage()
	img = append([]byte(nil), img...)
	binary.LittleEndian.PutUint32(img[446+12:], ^uint32(0))
	sum := md5.Sum(part)
	files := map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.hdb": fmt.Sprintf("%x:%d:GPT.Protective\n", sum, len(part)),
	}
	if dir := os.Getenv("EDR_GPT_FIXTURE_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "protective-range.img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	expectDetected(t, pipelineState(t, files), img, "GPT.Protective")
}
