package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"
)

const typedMboxPostmark = "From fixture@example.invalid Thu Jan 1 00:00:00 2026\n"

func typedMboxRules() map[string]string {
	return map[string]string{
		"mbox.ftm": "0:0:" + hx("From ") + ":mailbox:CL_TYPE_ANY:CL_TYPE_MAIL\n",
		"mbox.ndb": "Mbox.Coverage:0:*:" + hx(formatCoverageMarker) + "\n",
	}
}

func typedMboxFixture(t *testing.T, payload []byte) []byte {
	t.Helper()
	attachment := formatZipFixture(t, zip.Deflate, gzipFixture(t, payload))
	return []byte(typedMboxPostmark + "From: fixture@example.invalid\nMIME-Version: 1.0\nContent-Type: application/zip\nContent-Transfer-Encoding: base64\n\n" + base64.StdEncoding.EncodeToString(attachment) + "\n")
}

func TestTypedMboxPositiveAndBenign(t *testing.T) {
	for _, positive := range []bool{false, true} {
		payload, want := []byte("ordinary harmless typed mailbox control"), ""
		if positive {
			payload, want = []byte(formatCoverageMarker), "Mbox.Coverage"
		}
		data := typedMboxFixture(t, payload)
		if bytes.Contains(data, []byte(formatCoverageMarker)) {
			t.Fatal("fixture exposes marker")
		}
		expectScan(t, pipelineState(t, typedMboxRules()), data, want, "")
		// A prior complete MIME message must not contaminate the later
		// message or make a valid mailbox incomplete.
		first := []byte(typedMboxPostmark + "MIME-Version: 1.0\nContent-Type: text/plain\n\nordinary harmless first message\n\n")
		expectScan(t, pipelineState(t, typedMboxRules()), append(first, data...), want, "")
	}
}

func TestTypedMboxKeepsMalformedAndLimits(t *testing.T) {
	rules := typedMboxRules()
	for name, message := range map[string]string{
		"missing-boundary": "MIME-Version: 1.0\nContent-Type: multipart/mixed\n\nordinary text",
		"invalid-base64":   "MIME-Version: 1.0\nContent-Type: application/octet-stream\nContent-Transfer-Encoding: base64\n\n@@@",
		"ambiguous-type":   "MIME-Version: 1.0\nContent-Type: text/plain\nContent-Type : multipart/mixed; boundary=b\n\nordinary text",
	} {
		t.Run(name, func(t *testing.T) {
			if sig, err := pipelineState(t, rules).visit([]byte(typedMboxPostmark+message), 0); sig != "" || !errors.Is(err, ErrMalformed) {
				t.Fatalf("scan: %q %v; want malformed", sig, err)
			}
		})
	}
	data := typedMboxFixture(t, []byte("ordinary harmless typed mailbox control"))
	for _, limit := range []string{"depth", "count", "aggregate", "cancel"} {
		t.Run(limit, func(t *testing.T) {
			s := pipelineState(t, rules)
			want := ErrLimit
			switch limit {
			case "depth":
				s.maxDepth = 0
			case "count":
				s.maxFiles = 1
			case "aggregate":
				s.remaining = int64(len(data))
			case "cancel":
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				s.ctx = ctx
				want = context.Canceled
			}
			if sig, err := s.visit(data, 0); sig != "" || !errors.Is(err, want) {
				t.Fatalf("scan: %q %v; want %v", sig, err, want)
			}
		})
	}
	// The narrow mailbox dispatch must not relax the validation of an
	// explicitly declared message/rfc822 part with invalid leading syntax.
	s := pipelineState(t, rules)
	if sig, err := s.visitMessage(data, 0, 0); sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("declared message: %q %v; want malformed", sig, err)
	}
}
