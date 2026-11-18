package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestMIMEBase64WhitespaceNestedContent(t *testing.T) {
	rules := map[string]string{"test.ndb": "Native.Test:0:*:" + hx(formatCoverageMarker) + "\n"}
	for _, positive := range []bool{false, true} {
		payload := []byte("ordinary harmless attachment content")
		if positive {
			payload = []byte(formatCoverageMarker)
		}
		encoded := base64.StdEncoding.EncodeToString(gzipFixture(t, payload))
		var folded strings.Builder
		for i, b := range []byte(encoded) {
			folded.WriteByte(b)
			if i%3 == 0 {
				folded.WriteString(" \t\r\n")
			}
		}
		mail := []byte("MIME-Version: 1.0\r\nContent-Type: application/octet-stream\r\nContent-Transfer-Encoding: base64\r\n\r\n" + folded.String() + " \t\r\n")
		if bytes.Contains(mail, payload) {
			t.Fatal("test exposes decoded payload")
		}
		want := ""
		if positive {
			want = "Native.Test"
		}
		expectScan(t, pipelineState(t, rules), mail, want, "")
	}
}

func TestMIMEBase64WhitespaceReaderBoundaries(t *testing.T) {
	r := mimeBase64WhitespaceReader{context.Background(), strings.NewReader(" S\tG\r\nVs bG8= \t\r\n")}
	data, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, r))
	if err != nil || string(data) != "Hello" {
		t.Fatalf("decoded %q: %v", data, err)
	}
	for _, corrupt := range []string{"SG!VsbG8=", "SGVsbG8", "SGVsbG8=YQ=="} {
		_, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, mimeBase64WhitespaceReader{context.Background(), strings.NewReader(corrupt)}))
		if err == nil {
			t.Errorf("accepted corrupt base64 %q", corrupt)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = io.ReadAll(mimeBase64WhitespaceReader{ctx, strings.NewReader(strings.Repeat(" \t\r\n", 100))})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ignored cancel: %v", err)
	}
}

func TestMIMEBase64WhitespaceExpansionLimit(t *testing.T) {
	s := pipelineState(t, map[string]string{"test.ndb": "Never:0:*:ff00ff00ff00ff00\n"})
	payload := gzipFixture(t, bytes.Repeat([]byte("x"), 10000))
	encoded := base64.StdEncoding.EncodeToString(payload)
	mail := []byte("Content-Type: application/octet-stream\r\nContent-Transfer-Encoding: base64\r\n\r\n" + encoded[:4] + " \t" + encoded[4:])
	s.maxFile = 1000
	if sig, err := s.visit(mail, 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("sig=%q err=%v", sig, err)
	}
}
