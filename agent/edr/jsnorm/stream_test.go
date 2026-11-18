package jsnorm

import (
	"bytes"
	"testing"
)

func TestStreamAcrossEveryInputBoundary(t *testing.T) {
	for _, source := range [][]byte{
		[]byte("var x=1;danger(x)"),
		[]byte("function f(a){return a+1;}f(2)"),
		[]byte("/* comment */eval(\"danger()\")"),
	} {
		for _, chunk := range []int{1, 2, 3, 7, len(source)} {
			stream := NewStream()
			var segments [][]byte
			for i := 0; i < len(source); i += chunk {
				segment := source[i:min(i+chunk, len(source))]
				segments = append(segments, segment)
				stream.Feed(segment)
			}
			want, wantUnknown := NormalizeBlocks([][][]byte{segments})
			got, unknown := stream.Finish()
			if unknown != wantUnknown || !bytes.Equal(got, want) {
				t.Fatalf("source=%q chunk=%d got=%q unknown=%t want=%q unknown=%t", source, chunk, got, unknown, want, wantUnknown)
			}
		}
	}
}
