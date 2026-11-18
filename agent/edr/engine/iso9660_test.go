package engine

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
)

func putISO16(b []byte, v uint16) {
	binary.LittleEndian.PutUint16(b, v)
	binary.BigEndian.PutUint16(b[2:], v)
}

func putISO32(b []byte, v uint32) {
	binary.LittleEndian.PutUint32(b, v)
	binary.BigEndian.PutUint32(b[4:], v)
}

func isoTestRecord(name []byte, block, size uint32, flags byte) []byte {
	n := 33 + len(name)
	if n&1 != 0 {
		n++
	}
	b := make([]byte, n)
	b[0], b[25], b[32] = byte(n), flags, byte(len(name))
	putISO32(b[2:], block)
	putISO32(b[10:], size)
	putISO16(b[28:], 1)
	copy(b[33:], name)
	return b
}

func isoTestImage(content []byte) []byte {
	b := make([]byte, 32*isoSectorSize)
	vol := b[16*isoSectorSize : 17*isoSectorSize]
	vol[0], vol[6] = 1, 1
	copy(vol[1:], "CD001")
	putISO32(vol[80:], 32)
	putISO16(vol[120:], 1)
	putISO16(vol[124:], 1)
	putISO16(vol[128:], isoSectorSize)
	copy(vol[156:], isoTestRecord([]byte{0}, 20, isoSectorSize, 2))
	end := b[17*isoSectorSize : 18*isoSectorSize]
	end[0], end[6] = 255, 1
	copy(end[1:], "CD001")
	dir := b[20*isoSectorSize : 21*isoSectorSize]
	off := 0
	for _, record := range [][]byte{
		isoTestRecord([]byte{0}, 20, isoSectorSize, 2),
		isoTestRecord([]byte{1}, 20, isoSectorSize, 2),
		isoTestRecord([]byte("INNER.TXT;1"), 21, uint32(len(content)), 0),
	} {
		copy(dir[off:], record)
		off += len(record)
	}
	copy(b[21*isoSectorSize:], content)
	return b
}

func isoTestRaw(cooked []byte, sectorSize, prefix int) []byte {
	out := make([]byte, len(cooked)/isoSectorSize*sectorSize)
	for i := 0; i < len(cooked)/isoSectorSize; i++ {
		copy(out[i*sectorSize+prefix:], cooked[i*isoSectorSize:(i+1)*isoSectorSize])
	}
	return out
}

func isoHashRules(content []byte) map[string]string {
	sum := md5.Sum(content)
	return map[string]string{"iso.hdb": fmt.Sprintf("%x:%d:ISO.Child\n", sum, len(content))}
}

func TestISO9660ScansContainedFileByHash(t *testing.T) {
	child := []byte("This file is only found inside the ISO image.")
	image := isoTestImage(child)
	if got := fmt.Sprintf("%x", sha256.Sum256(image)); got != "7196cf17f1ff44f18d537be47c839098f6d9e96a32fb43759e5ec6ae297a4f78" {
		t.Fatalf("oracle input changed: %s", got)
	}
	if !isoHeader(image) {
		t.Fatal("valid primary volume not recognized")
	}
	expectDetected(t, pipelineState(t, isoHashRules(child)), image, "ISO.Child")
	// The image's own bytes are not the same whole-file hash as its member.
	other := isoTestImage([]byte("A harmless alternate file in the same image."))
	expectScan(t, pipelineState(t, isoHashRules(child)), other, "", "")
	metaRule := fmt.Sprintf("ISO.Name:CL_TYPE_ISO9660:%d:^INNER\\.TXT$:%d:%d:0:0:*:*\n", len(image), len(child), len(child))
	expectScan(t, pipelineState(t, map[string]string{"iso.cdb": metaRule}), image, "ISO.Name", "")
}

func TestISO9660NestedDirectoryAndJolietVolume(t *testing.T) {
	child := []byte("This file is only found inside the ISO image.")
	t.Run("nested", func(t *testing.T) {
		b := isoTestImage(child)
		root := b[20*isoSectorSize : 21*isoSectorSize]
		off := len(isoTestRecord([]byte{0}, 20, isoSectorSize, 2)) + len(isoTestRecord([]byte{1}, 20, isoSectorSize, 2))
		copy(root[off:], isoTestRecord([]byte("SUBDIR"), 22, isoSectorSize, 2))
		sub := b[22*isoSectorSize : 23*isoSectorSize]
		pos := 0
		for _, record := range [][]byte{
			isoTestRecord([]byte{0}, 22, isoSectorSize, 2),
			isoTestRecord([]byte{1}, 20, isoSectorSize, 2),
			isoTestRecord([]byte("INNER.TXT;1"), 21, uint32(len(child)), 0),
		} {
			copy(sub[pos:], record)
			pos += len(record)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != "a9c67db95b6a57fc8dc0a069141faec44996e03688b1bdc6b382d49cc40fd6df" {
			t.Fatalf("oracle input changed: %s", got)
		}
		expectDetected(t, pipelineState(t, isoHashRules(child)), b, "ISO.Child")
	})
	t.Run("joliet only", func(t *testing.T) {
		b := isoTestImage(child)
		primary := b[20*isoSectorSize : 21*isoSectorSize]
		for i := len(isoTestRecord([]byte{0}, 20, isoSectorSize, 2)) + len(isoTestRecord([]byte{1}, 20, isoSectorSize, 2)); i < len(primary); i++ {
			primary[i] = 0
		}
		secondary := b[17*isoSectorSize : 18*isoSectorSize]
		copy(secondary, b[16*isoSectorSize:17*isoSectorSize])
		secondary[0] = 2
		copy(secondary[88:], "%/E")
		copy(secondary[156:], isoTestRecord([]byte{0}, 22, isoSectorSize, 2))
		end := b[18*isoSectorSize : 19*isoSectorSize]
		end[0], end[6] = 255, 1
		copy(end[1:], "CD001")
		jolietName := []byte{0, 'U', 0, 'N', 0, 'I', 0, 'C', 0, 'O', 0, 'D', 0, 'E', 0, '.', 0, 'T', 0, 'X', 0, 'T', 0, ';', 0, '1'}
		jdir := b[22*isoSectorSize : 23*isoSectorSize]
		pos := 0
		for _, record := range [][]byte{
			isoTestRecord([]byte{0}, 22, isoSectorSize, 2),
			isoTestRecord([]byte{1}, 22, isoSectorSize, 2),
			isoTestRecord(jolietName, 21, uint32(len(child)), 0),
		} {
			copy(jdir[pos:], record)
			pos += len(record)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != "05bb8a5e56b140cf4a29a264d87dd026b8a814a6db53982a141ef2f6c6f3ede4" {
			t.Fatalf("oracle input changed: %s", got)
		}
		expectDetected(t, pipelineState(t, isoHashRules(child)), b, "ISO.Child")
	})
}

func TestISO9660DistinctDirectoryLengthsAtOneExtent(t *testing.T) {
	child := []byte("Harmless content behind a second directory extent.")
	b := isoTestImage(child)
	root := b[20*isoSectorSize : 21*isoSectorSize]
	off := len(isoTestRecord([]byte{0}, 20, isoSectorSize, 2)) + len(isoTestRecord([]byte{1}, 20, isoSectorSize, 2))
	short := isoTestRecord([]byte("SHORT"), 22, uint32(off), 2)
	full := isoTestRecord([]byte("FULL"), 22, isoSectorSize, 2)
	copy(root[off:], short)
	copy(root[off+len(short):], full)
	sub := b[22*isoSectorSize : 23*isoSectorSize]
	pos := 0
	for _, record := range [][]byte{
		isoTestRecord([]byte{0}, 22, isoSectorSize, 2),
		isoTestRecord([]byte{1}, 20, isoSectorSize, 2),
		isoTestRecord([]byte("INNER.TXT;1"), 21, uint32(len(child)), 0),
	} {
		copy(sub[pos:], record)
		pos += len(record)
	}
	expectDetected(t, pipelineState(t, isoHashRules(child)), b, "ISO.Child")
}

func TestISO9660RawSectorLayouts(t *testing.T) {
	child := []byte("This file is only found inside the ISO image.")
	cooked := isoTestImage(child)
	for _, tc := range []struct {
		name        string
		sector, pre int
		sha256      string
	}{
		{"raw2336-p0", 2336, 0, "0baf06eafaa03a3ec85c90077deefc6097680e0aa4450ba17670de37d00d2d59"},
		{"raw2336", 2336, 8, "4c36b7076e6977fec3963f4143844d6d36e5515d9de798e32275e71e4bdd63da"},
		{"raw2336-p16", 2336, 16, "da7fe1966bf1a6243fca0a2e7dec5e77729a4637faa03a569546032be6751e45"},
		{"raw2336-p24", 2336, 24, "02ea36bb1e50bfa800f822bf52c72af023c39cad8bce9d4b91fc220f82db56ed"},
		{"raw2352-p0", 2352, 0, "cfeb938b6bc74978ca597f26ac7c0b518a350815536eafbe0fc929cbb022b7a2"},
		{"raw2352-p8", 2352, 8, "504a4a827ed0fb64fe5021b6d49f493612749826dd1e37daefde094d98742704"},
		{"raw2352", 2352, 16, "c73f6a60d55831254bbfc58c9991396ac90585c309ba7255f08ae27f79b7ac54"},
		{"raw2352-p24", 2352, 24, "7061faee43dbb75d5b7c4c28bc5ea5546c0ca8a6fd6c231e53afabcc203e87ee"},
		{"raw2448-p0", 2448, 0, "68067b8d0fb473f1f8038eef7e4409c6a8acd7c9a257eb3ed84985ae6c3a9b45"},
		{"raw2448-p8", 2448, 8, "631d2220cb08adfd106182102968dc854b099087664f622a080ef7e15d349bf9"},
		{"raw2448", 2448, 16, "6ccd3b33d6c8549729241be30685ff0c1ea344fba8dea59e9ee95929988485e0"},
		{"raw2448-p24", 2448, 24, "91ba3f6ba2b3b12b7bde33e8476e366b610777f9b13cdf8223f0733325f8202a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			image := isoTestRaw(cooked, tc.sector, tc.pre)
			if !isoHeader(image) {
				t.Fatal("raw primary volume not recognized")
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(image)); got != tc.sha256 {
				t.Fatalf("oracle input changed: %s", got)
			}
			expectDetected(t, pipelineState(t, isoHashRules(child)), image, "ISO.Child")
			truncated := image[:21*tc.sector+tc.pre+3]
			expectIncomplete(t, pipelineState(t, isoHashRules(child)), truncated, ErrMalformed)
		})
	}
}

func TestISO9660SmallerLogicalBlocks(t *testing.T) {
	child := []byte("This file is only found inside the ISO image.")
	for _, blockSize := range []int{512, 1024} {
		t.Run(fmt.Sprint(blockSize), func(t *testing.T) {
			b := isoTestImage(child)
			blocksPerSector := isoSectorSize / blockSize
			putISO16(b[16*isoSectorSize+128:], uint16(blockSize))
			putISO32(b[16*isoSectorSize+156+2:], uint32(20*blocksPerSector))
			dir := b[20*isoSectorSize : 21*isoSectorSize]
			pos := 0
			for _, extent := range []int{20, 20, 21} {
				putISO32(dir[pos+2:], uint32(extent*blocksPerSector))
				if dir[pos] == 0 {
					t.Fatal("missing test directory entry")
				}
				pos += int(dir[pos])
			}
			want := "4816573d4c433cbf65a75f28418a93902d94893196ec4cff481a26a95f3c8268"
			if blockSize == 1024 {
				want = "c3acefdd0f57a721b4f9e974e3ac29067983e1ec52164aa3a9877b4c864cb7b8"
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != want {
				t.Fatalf("oracle input changed: %s", got)
			}
			expectDetected(t, pipelineState(t, isoHashRules(child)), b, "ISO.Child")
			raw := isoTestRaw(b, 2352, 16)
			if blockSize == 512 {
				if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "04aea92e74773ccb48f3b8e5547bf88acb9618e7cf5ec3f9d7e8399b6fb3301d" {
					t.Fatalf("oracle input changed: %s", got)
				}
			}
			expectDetected(t, pipelineState(t, isoHashRules(child)), raw, "ISO.Child")
		})
	}
}

func TestISO9660MalformedAndUnsupportedMembersRemainIncomplete(t *testing.T) {
	child := []byte("A harmless file of adequate length.")
	base := isoTestImage(child)
	t.Run("truncated member", func(t *testing.T) {
		b := append([]byte(nil), base...)
		b = b[:21*isoSectorSize+3]
		expectIncomplete(t, pipelineState(t, isoHashRules(child)), b, ErrMalformed)
	})
	t.Run("interleaved member", func(t *testing.T) {
		b := append([]byte(nil), base...)
		dir := b[20*isoSectorSize : 21*isoSectorSize]
		off := len(isoTestRecord([]byte{0}, 20, isoSectorSize, 2)) + len(isoTestRecord([]byte{1}, 20, isoSectorSize, 2))
		dir[off+26] = 1
		expectIncomplete(t, pipelineState(t, isoHashRules(child)), b, ErrUnsupported)
	})
	t.Run("multi extent", func(t *testing.T) {
		b := append([]byte(nil), base...)
		dir := b[20*isoSectorSize : 21*isoSectorSize]
		off := len(isoTestRecord([]byte{0}, 20, isoSectorSize, 2)) + len(isoTestRecord([]byte{1}, 20, isoSectorSize, 2))
		dir[off+25] |= 0x80
		// The reference scans this record's section despite the flag.
		expectDetected(t, pipelineState(t, isoHashRules(child)), b, "ISO.Child")
		expectIncomplete(t, pipelineState(t, isoHashRules([]byte("An unrelated harmless section."))), b, ErrUnsupported)
	})
	t.Run("bad endian pair", func(t *testing.T) {
		b := append([]byte(nil), base...)
		b[16*isoSectorSize+128+2] ^= 1
		expectIncomplete(t, pipelineState(t, isoHashRules(child)), b, ErrUnsupported)
	})
}

func TestISO9660ScansEachMultiExtentSection(t *testing.T) {
	first := []byte("Harmless first extent with its own digest.")
	second := []byte("Harmless second extent with a different digest.")
	one := isoTestImage(first)
	oneDir := one[20*isoSectorSize : 21*isoSectorSize]
	off := len(isoTestRecord([]byte{0}, 20, isoSectorSize, 2)) + len(isoTestRecord([]byte{1}, 20, isoSectorSize, 2))
	oneDir[off+25] = 0x80
	if got := fmt.Sprintf("%x", sha256.Sum256(one)); got != "b87d1b293c3935fc516006f920dcc5fbac70427d455fdc80df946c46cc549759" {
		t.Fatalf("single-section oracle image changed: %s", got)
	}
	expectDetected(t, pipelineState(t, isoHashRules(first)), one, "ISO.Child")
	expectIncomplete(t, pipelineState(t, isoHashRules(second)), one, ErrUnsupported)
	b := isoTestImage(first)
	dir := b[20*isoSectorSize : 21*isoSectorSize]
	firstRecord := isoTestRecord([]byte("INNER.TXT;1"), 21, uint32(len(first)), 0x80)
	secondRecord := isoTestRecord([]byte("INNER.TXT;1"), 22, uint32(len(second)), 0)
	copy(dir[off:], firstRecord)
	copy(dir[off+len(firstRecord):], secondRecord)
	copy(b[22*isoSectorSize:], second)
	if got := fmt.Sprintf("%x", sha256.Sum256(b)); got != "0978068503b299d5979f8356d87b82a8a434acd4a28723043905eba76adbafe7" {
		t.Fatalf("two-section oracle image changed: %s", got)
	}
	for _, section := range [][]byte{first, second} {
		expectDetected(t, pipelineState(t, isoHashRules(section)), b, "ISO.Child")
	}
	expectIncomplete(t, pipelineState(t, isoHashRules([]byte("Absent harmless extent digest."))), b, ErrUnsupported)
	expectIncomplete(t, pipelineState(t, isoHashRules(append(append([]byte(nil), first...), second...))), b, ErrUnsupported)
}

func TestISO9660RespectsScanBudget(t *testing.T) {
	child := []byte("A harmless file of adequate length.")
	s := pipelineState(t, isoHashRules(child))
	s.remaining = int64(len(isoTestImage(child)))
	sig, err := s.visit(isoTestImage(child), 0)
	if sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("sig=%q err=%v, want resource limit", sig, err)
	}
}
