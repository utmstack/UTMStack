package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func entityURIPage(t *testing.T, payload []byte, hex bool) []byte {
	t.Helper()
	encoded := base64.StdEncoding.EncodeToString(gzipFixture(t, payload))
	var b strings.Builder
	for _, c := range []byte(encoded) {
		if hex {
			fmt.Fprintf(&b, "&#x%x;", c)
		} else {
			fmt.Fprintf(&b, "&#%d;", c)
		}
	}
	return []byte(`<html><object data="data:application/octet-stream;base64,` + b.String() + `"></object></html>`)
}

func TestHTMLNumericEntityDataURI(t *testing.T) {
	rules := map[string]string{"test.ndb": "Native.Test:0:*:" + hx(formatCoverageMarker) + "\n"}
	for _, hex := range []bool{false, true} {
		for _, positive := range []bool{false, true} {
			payload := []byte("ordinary harmless data")
			want := ""
			if positive {
				payload = []byte(formatCoverageMarker)
				want = "Native.Test"
			}
			page := entityURIPage(t, payload, hex)
			if bytes.Contains(page, payload) {
				t.Fatal("fixture exposes payload")
			}
			expectScan(t, pipelineState(t, rules), page, want, "")
		}
	}
}

func TestHTMLEntityURIOnlyScansChangedURLAttributes(t *testing.T) {
	rules := map[string]string{"test.ndb": "Native.Test:0:*:" + hx(formatCoverageMarker) + "\n"}
	page := entityURIPage(t, []byte(formatCoverageMarker), false)
	for _, wrapper := range []func([]byte) []byte{
		func(p []byte) []byte { return append(append([]byte("<!--"), p...), []byte("-->")...) },
		func(p []byte) []byte {
			return append(append([]byte("<script>var text='"), p...), []byte("';</script>")...)
		},
	} {
		s := pipelineState(t, rules)
		if sig, err := s.entityDataURIs(wrapper(page), 0); sig != "" || err != nil {
			t.Fatalf("parsed inactive markup: %q %v", sig, err)
		}
	}
	plain := []byte(`<html><img alt="&amp;" src="data:application/octet-stream;base64,` + base64.StdEncoding.EncodeToString([]byte("ordinary harmless data")) + `"></html>`)
	s := pipelineState(t, rules)
	before := s.files
	if sig, err := s.entityDataURIs(plain, 0); sig != "" || err != nil || s.files != before {
		t.Fatalf("reparsed unchanged URI: %q %v files=%d", sig, err, s.files)
	}
}

func TestHTMLEntityURIResourceAndCancellation(t *testing.T) {
	rules := map[string]string{"test.ndb": "Never:0:*:ff00ff00ff00ff00\n"}
	page := entityURIPage(t, bytes.Repeat([]byte("x"), 10000), false)
	for _, which := range []string{"file", "aggregate", "depth", "objects"} {
		s := pipelineState(t, rules)
		switch which {
		case "file":
			s.maxFile = 1000
		case "aggregate":
			s.remaining = 1000
		case "depth":
			s.maxDepth = 0
		case "objects":
			s.maxFiles = 0
		}
		if sig, err := s.entityDataURIs(page, 0); sig != "" || !errors.Is(err, ErrLimit) {
			t.Fatalf("%s: %q %v", which, sig, err)
		}
	}
	s := pipelineState(t, rules)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.ctx = ctx
	if _, err := s.entityDataURIs(page, 0); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

// Trigger cancellation at a deterministic cooperative check, so the decoder
// must stop after beginning a large attribute rather than only at entry.
type decodeCancelContext struct {
	context.Context
	checks int
}

func (c *decodeCancelContext) Err() error {
	c.checks++
	if c.checks >= 3 {
		return context.Canceled
	}
	return nil
}

func TestDataURIDecodingChecksCancellationDuringBody(t *testing.T) {
	ctx := &decodeCancelContext{Context: context.Background()}
	body := bytes.Repeat([]byte("A"), 1<<20)
	decoded, err := dataURIDecodeContext(ctx, body, true, 1<<20)
	if !errors.Is(err, context.Canceled) || len(decoded) == 0 || len(decoded) >= len(body)/2 {
		t.Fatalf("decoded %d bytes before %v", len(decoded), err)
	}
	s := pipelineState(t, map[string]string{"never.ndb": "Never:0:*:ff00ff00ff00ff00\n"})
	s.maxFiles = 0
	if sig, err := s.dataURIContent([]byte(";base64"), body, 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("object preflight: %q %v", sig, err)
	}
}
