package engine

import "testing"

func TestBytecodeMetadataOptionMatchesCollectionMode(t *testing.T) {
	const fileProperties = uint32(0x10000000)
	off := (&scanState{}).bytecodeEngineOptions()
	on := (&scanState{collectMetadata: true}).bytecodeEngineOptions()
	if off.Scan&fileProperties != 0 || on.Scan&fileProperties == 0 {
		t.Fatalf("file-properties option: off=%#x on=%#x", off.Scan, on.Scan)
	}
	if off.Scan|fileProperties != on.Scan || off.Database != on.Database {
		t.Fatalf("metadata option changed other bytecode flags: off=%#v on=%#v", off, on)
	}
	if (&scanState{}).representationMetadata() != nil || (&scanState{collectMetadata: true}).representationMetadata()["Magic"] != "CLAMJSONv0" {
		t.Fatal("representation metadata activation does not follow scan option")
	}
}
