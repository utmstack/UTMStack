package signatures

import "testing"

func TestTextNormalizationFollowsRootSeven(t *testing.T) {
	d := compileTest(t, "ndb", "T7a:7:*:41424344\nT7b:7:*:4142{3}43*444546\nLong:0:*:"+
		"41424344454647484950515253545556575859\nT7c:7:EOF-4:4142(4344|454647)\n")
	chunk, fileMode := d.TextNormalization()
	// The longest root-7 part is 4142{3}43: two bytes, three wildcards, one byte.
	if chunk != 131072+6 || !fileMode {
		t.Fatalf("chunk %d file mode %v, want %d true", chunk, fileMode, 131072+6)
	}
	d = compileTest(t, "ndb", "T7:7:*:41424344\n")
	if chunk, fileMode := d.TextNormalization(); chunk != 131072+4 || fileMode {
		t.Fatalf("chunk %d file mode %v: a root without relative offsets or bytecode is matched as a buffer", chunk, fileMode)
	}
}
