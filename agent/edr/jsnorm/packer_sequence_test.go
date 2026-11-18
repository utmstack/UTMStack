package jsnorm

import (
	"bytes"
	"testing"
)

// The native named-call walker concatenates separately decoded payloads.
// With one shared dictionary and an encoded token boundary between payloads,
// this must equal one invocation containing their concatenation. This stresses
// replacement ranges and declaration transitions through two different paths.
func FuzzNamedPackerSequence(f *testing.F) {
	f.Add([]byte("0(1);function alpha(){return beta;} alpha();"), uint16(8))
	f.Add([]byte(""), uint16(0))
	f.Fuzz(func(t *testing.T, data []byte, split uint16) {
		if len(data) > 4096 {
			t.Skip()
		}
		const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_+()-;{}= "
		payload := make([]byte, len(data))
		for i, b := range data {
			payload[i] = alphabet[int(b)%len(alphabet)]
		}
		cut := 0
		if len(payload) > 0 {
			cut = int(split) % (len(payload) + 1)
		}
		a, b := string(payload[:cut])+";", string(payload[cut:])+";"
		call := func(text string) string { return `pack('` + text + `',36,4,'alpha|beta|safe|marker'.split('|'),0,{})` }
		const declaration = `function pack(p,a,c,k,e,r){return p} `
		one, oneUnknown := NormalizeBlocks([][][]byte{{[]byte(declaration + call(a+b) + ";")}})
		many, manyUnknown := NormalizeBlocks([][][]byte{{[]byte(declaration + call(a) + ";" + call(b) + ";")}})
		if oneUnknown || manyUnknown || !bytes.Equal(one, many) {
			t.Fatalf("single=%q unknown=%t repeated=%q unknown=%t", one, oneUnknown, many, manyUnknown)
		}
		if len(one) > maxPackerOutput+len(declaration)+512 {
			t.Fatal("unbounded normalized output")
		}
	})
}
