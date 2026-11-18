package engine

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func mailAttachmentBody(t *testing.T, boundary string) string {
	t.Helper()
	attachment := base64.StdEncoding.EncodeToString(gzipFixture(t, []byte(testMarker)))
	return "MIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n\r\n" +
		"--" + boundary + "\r\nContent-Type: text/plain\r\n\r\nplease see attached\r\n" +
		"--" + boundary + "\r\nContent-Type: application/octet-stream\r\nContent-Transfer-Encoding: base64\r\n\r\n" +
		attachment + "\r\n--" + boundary + "--\r\n"
}

// A MIME part declared message/rfc822 is an encapsulated message. Its
// attachments are decoded whichever header field the inner message starts
// with, instead of depending on the top-level mail recognition heuristic.
func TestMIMEEncapsulatedMessageIsParsed(t *testing.T) {
	e := nativeTestEngine(t)
	inner := func(firstHeader string) string {
		return firstHeader + "\r\nFrom: c@example.com\r\nSubject: invoice\r\n" + mailAttachmentBody(t, "inner")
	}
	outer := func(partHeaders, innerMsg string) []byte {
		return []byte("From: a@example.com\r\nTo: b@example.com\r\nSubject: Fwd: invoice\r\n" +
			"MIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=\"outer\"\r\n\r\n" +
			"--outer\r\nContent-Type: text/plain\r\n\r\nforwarding\r\n" +
			"--outer\r\n" + partHeaders + "\r\n\r\n" + innerMsg + "\r\n--outer--\r\n")
	}
	// Control: an inner message that happens to begin with "From:" is detected.
	if clean, sig, err := e.Scan(context.Background(), outer("Content-Type: message/rfc822", inner("From: x@example.com"))); clean || sig != "Native.Test" || err != nil {
		t.Fatalf("control not detected: clean=%v sig=%q err=%v", clean, sig, err)
	}
	for _, first := range []string{"Date: Mon, 21 Sep 2026 10:00:00 +0000", "Subject: invoice", "Message-ID: <1@example.com>", "To: b@example.com", "X-Custom: 1"} {
		t.Run(first[:strings.IndexByte(first, ':')], func(t *testing.T) {
			clean, sig, err := e.Scan(context.Background(), outer("Content-Type: message/rfc822", inner(first)))
			if clean || sig != "Native.Test" || err != nil {
				t.Fatalf("message/rfc822 part whose first header is %q: clean=%v sig=%q err=%v; want detection", first, clean, sig, err)
			}
		})
	}
	t.Run("global", func(t *testing.T) {
		clean, sig, err := e.Scan(context.Background(), outer("Content-Type: message/global", inner("Date: Mon, 21 Sep 2026 10:00:00 +0000")))
		if clean || sig != "Native.Test" || err != nil {
			t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
		}
	})
	t.Run("base64 transfer encoding", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte(inner("Date: Mon, 21 Sep 2026 10:00:00 +0000")))
		clean, sig, err := e.Scan(context.Background(), outer("Content-Type: message/rfc822\r\nContent-Transfer-Encoding: base64", encoded))
		if clean || sig != "Native.Test" || err != nil {
			t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
		}
	})
	t.Run("benign", func(t *testing.T) {
		benign := "Date: Mon, 21 Sep 2026 10:00:00 +0000\r\nSubject: hello\r\n\r\nordinary text\r\n"
		if clean, sig, err := e.Scan(context.Background(), outer("Content-Type: message/rfc822", benign)); !clean || sig != "" || err != nil {
			t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
		}
	})
	t.Run("empty header block", func(t *testing.T) {
		clean, sig, err := e.Scan(context.Background(), outer("Content-Type: message/rfc822", "\r\n"+inner("Date: Mon, 21 Sep 2026 10:00:00 +0000")))
		if clean || sig != "Native.Test" || err != nil {
			t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
		}
	})
	t.Run("malformed", func(t *testing.T) {
		clean, sig, err := e.Scan(context.Background(), outer("Content-Type: message/rfc822", " folded first line\r\nnot a header line"))
		if clean || sig != "" || !errors.Is(err, ErrMalformed) {
			t.Fatalf("clean=%v sig=%q err=%v; want malformed", clean, sig, err)
		}
	})
	t.Run("partial", func(t *testing.T) {
		// A message/partial fragment cannot be reassembled with its siblings.
		clean, sig, err := e.Scan(context.Background(), outer("Content-Type: message/partial; id=\"a@example.com\"; number=1; total=2", "Subject: part one\r\n\r\nfirst half"))
		if clean || sig != "" || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("clean=%v sig=%q err=%v; want unsupported", clean, sig, err)
		}
	})
}

// Ordinary delivered mail is recognized even when its trace headers exceed
// 4 KiB or it begins with Return-Path: or Delivered-To:. Otherwise its base64
// attachments are never decoded and the message returns clean.
func TestDeliveredMailWithLargeTraceHeadersIsRecognized(t *testing.T) {
	e := nativeTestEngine(t)
	body := "From: a@example.com\r\nTo: b@example.com\r\nSubject: invoice\r\n" + mailAttachmentBody(t, "b")
	received := func(n int) string {
		var b strings.Builder
		for i := 0; i < n; i++ {
			fmt.Fprintf(&b, "Received: from relay%02d.example.net (relay%02d.example.net [192.0.2.%d])\r\n\tby mx.example.com with ESMTPS id x%02d\r\n", i, i, i%250, i)
		}
		return b.String()
	}
	// Control: one Received line is recognized and the attachment detected.
	if clean, sig, err := e.Scan(context.Background(), []byte(received(1)+body)); clean || sig != "Native.Test" || err != nil {
		t.Fatalf("control not detected: clean=%v sig=%q err=%v", clean, sig, err)
	}
	for name, msg := range map[string]string{
		"received-over-4KiB": received(45) + body,
		"return-path-first":  "Return-Path: <a@example.com>\r\n" + received(2) + body,
		"delivered-to-first": "Delivered-To: b@example.com\r\n" + received(2) + body,
		"arc-and-dkim":       "ARC-Seal: i=1; a=rsa-sha256; t=1; cv=none;\r\n        d=example.com; s=arc; b=abc\r\nDKIM-Signature: v=1; a=rsa-sha256\r\n" + received(3) + body,
		"lf-line-endings":    strings.ReplaceAll(received(45)+body, "\r\n", "\n"),
	} {
		t.Run(name, func(t *testing.T) {
			if clean, sig, err := e.Scan(context.Background(), []byte(msg)); clean || sig != "Native.Test" || err != nil {
				t.Fatalf("delivered message (%d bytes): clean=%v sig=%q err=%v; want detection", len(msg), clean, sig, err)
			}
		})
	}
	t.Run("header block over the bound", func(t *testing.T) {
		msg := received(mailMaxHeaderLines) + body
		clean, sig, err := e.Scan(context.Background(), []byte(msg))
		if clean || sig != "" || !errors.Is(err, ErrLimit) {
			t.Fatalf("clean=%v sig=%q err=%v; want resource limit", clean, sig, err)
		}
	})
	t.Run("MIME fields after a line that is not a field", func(t *testing.T) {
		// Lenient readers may still see these MIME fields; the message cannot
		// be parsed here, so it is malformed rather than generic text.
		msg := "From: a@example.com\r\nthis line has no colon\r\n" + mailAttachmentBody(t, "b")
		clean, sig, err := e.Scan(context.Background(), []byte(msg))
		if clean || sig != "" || !errors.Is(err, ErrMalformed) {
			t.Fatalf("clean=%v sig=%q err=%v; want malformed", clean, sig, err)
		}
	})
	t.Run("ordinary text stays generic", func(t *testing.T) {
		for _, text := range []string{
			"Dear team,\r\nContent-Type: is discussed below.\r\n",
			strings.Repeat("2026-09-23T10:00:00 INFO: service started\n", 10000),
			"Package: demo\nVersion: 1.0\n\nPackage: other\n",
		} {
			if clean, sig, err := e.Scan(context.Background(), []byte(text)); !clean || sig != "" || err != nil {
				t.Fatalf("%.30q: clean=%v sig=%q err=%v", text, clean, sig, err)
			}
		}
	})
}

// RFC 5322 obsolete syntax allows white space before a field's colon, and
// receivers must accept it. The recognizer counts such a field as MIME, so the
// parser must read its value too: Go's readers keep the white space in the key.
func TestMIMEFieldsWithSpaceBeforeColonAreRead(t *testing.T) {
	e := nativeTestEngine(t)
	attachment := base64.StdEncoding.EncodeToString(gzipFixture(t, []byte(testMarker)))
	multipart := func(typeField, encodingField string) string {
		return "From: a@example.com\r\nMIME-Version: 1.0\r\n" + typeField + " multipart/mixed; boundary=\"b\"\r\n\r\n" +
			"--b\r\nContent-Type: application/octet-stream\r\n" + encodingField + " base64\r\n\r\n" + attachment + "\r\n--b--\r\n"
	}
	for name, msg := range map[string]string{
		"message type":  multipart("Content-Type :", "Content-Transfer-Encoding:"),
		"part encoding": multipart("Content-Type:", "Content-Transfer-Encoding :"),
		"single part":   "Received: from relay\r\nContent-Transfer-Encoding  : base64\r\n\r\n" + attachment + "\r\n",
	} {
		t.Run(name, func(t *testing.T) {
			if clean, sig, err := e.Scan(context.Background(), []byte(msg)); clean || sig != "Native.Test" || err != nil {
				t.Fatalf("clean=%v sig=%q err=%v; want the attachment decoded", clean, sig, err)
			}
		})
	}
	// Two spellings of one field with different values are ambiguous.
	ambiguous := "From: a@example.com\r\nContent-Type: text/plain\r\n" + strings.TrimPrefix(multipart("Content-Type :", "Content-Transfer-Encoding:"), "From: a@example.com\r\n")
	if clean, sig, err := e.Scan(context.Background(), []byte(ambiguous)); clean || sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("ambiguous content type: clean=%v sig=%q err=%v; want malformed", clean, sig, err)
	}
}
