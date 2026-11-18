package engine

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Read the old serialized representation independently: these assertions
// protect compatibility bytes and decoder inputs while production uses fields.
func legacyHTMLPart(data []byte) (contentType, encoding string, body []byte, ok bool) {
	const prefix = "From html-normalise\nContent-type: "
	if !bytes.HasPrefix(data, []byte(prefix)) {
		return "", "", nil, false
	}
	header, body, ok := bytes.Cut(data[len(prefix):], []byte("\n\n"))
	if !ok {
		return "", "", nil, false
	}
	contentType = string(header)
	if value, rest, found := strings.Cut(contentType, "\n"); found {
		if rest != "Content-transfer-encoding: base64" {
			return "", "", nil, false
		}
		contentType, encoding = value, "base64"
	}
	if strings.HasSuffix(contentType, ";base64") {
		contentType = strings.TrimSuffix(contentType, ";base64")
	} else {
		encoding = ""
	}
	if strings.HasPrefix(contentType, ";") {
		contentType = "text/plain" + contentType
	}
	return contentType, encoding, body, true
}

func checkHTMLParts(t testing.TB, page []byte) {
	t.Helper()
	for _, p := range normalizeHTML(page).rfc2397 {
		ct, enc, body, ok := legacyHTMLPart(p.compatibility)
		if !ok || ct != p.contentType || enc != p.encoding || !bytes.Equal(body, p.body) {
			t.Fatalf("metadata differs: type=%q/%q encoding=%q/%q body=%q/%q envelope=%q", p.contentType, ct, p.encoding, enc, p.body, body, p.compatibility)
		}
		if len(body) > 0 && &body[0] != &p.body[0] {
			t.Fatal("body copied instead of sharing the compatibility representation")
		}
	}
}

func TestHTMLDataPartsPreserveCompatibility(t *testing.T) {
	for _, media := range []string{"", ";charset=UTF-8", ";base64", "text/plain", "TEXT/PLAIN;BASE64", "text/plain;base64=ordinary", "text/plain;name=x;base64", "text/plain;base64;charset=x", "text/plain; BASE64", strings.Repeat("x", 1100)} {
		for _, ending := range []string{`"></html>`, "", `&amp;ignored` + `"></html>`} {
			page := []byte(`<html><img src="data:` + media + `,SGVsbG8=%41` + ending)
			checkHTMLParts(t, page)
		}
	}
	// Multiple records must keep their own body storage and decoder metadata.
	page := []byte(`<html><img src="data:;base64,SGVsbG8="><img src="data:text/plain;base64=ordinary,other%20text"></html>`)
	parts := normalizeHTML(page).rfc2397
	if len(parts) != 2 || parts[0].encoding != "base64" || parts[1].encoding != "" || string(parts[1].body) != "other text" {
		t.Fatalf("independent records: %+v", parts)
	}
	checkHTMLParts(t, page)
}

func TestHTMLDataPartAllowlistKeepsCompatibilityIdentity(t *testing.T) {
	page := []byte(`<html><img src="data:application/octet-stream;base64,` + base64.StdEncoding.EncodeToString([]byte(formatCoverageMarker)) + `"></html>`)
	part := normalizeHTML(page).rfc2397[0]
	rules := map[string]string{
		"child.ndb": "Child.Marker:0:*:" + hx(formatCoverageMarker) + "\n",
		"child.sfp": fmt.Sprintf("%x:%d:Allowed.HTML.Child\n", md5.Sum(part.compatibility), len(part.compatibility)),
	}
	// Isolate this private representation; the page's independent generic
	// data-URI view intentionally has its own decoded-object identity.
	if sig, err := pipelineState(t, rules).htmlViews(page, 0, signatures.TypeHTML); sig != "" || err != nil {
		t.Fatalf("allowed private object: %q %v", sig, err)
	}
	delete(rules, "child.sfp")
	if sig, err := pipelineState(t, rules).htmlViews(page, 0, signatures.TypeHTML); sig != "Child.Marker" || err != nil {
		t.Fatalf("private object without allowlist: %q %v", sig, err)
	}
}

func FuzzHTMLDataParts(f *testing.F) {
	for _, page := range []string{`<img src="data:;base64,SGVsbG8=">`, `<html><a href=data:,a%41b%zc&amp;d>tail`, `<img src="data:text/plain;base64=ordinary,abc">`, `<img src="data:text/plain,unfinished`} {
		f.Add([]byte(page))
	}
	f.Fuzz(func(t *testing.T, page []byte) {
		if len(page) > 8192 {
			t.Skip()
		}
		checkHTMLParts(t, page)
	})
}

func TestStoredChildBytesMatchesReader(t *testing.T) {
	for _, payload := range [][]byte{nil, []byte("short"), []byte("six!!!"), []byte(formatCoverageMarker), bytes.Repeat([]byte{0}, 4096)} {
		for _, policy := range []string{"normal", "exact", "file", "aggregate", "zero", "negative", "depth", "count", "cancel", "cancel-and-depth"} {
			t.Run(fmt.Sprintf("%d/%s", len(payload), policy), func(t *testing.T) {
				rules := map[string]string{"child.ndb": "Child.Marker:0:*:" + hx(formatCoverageMarker) + "\n"}
				old, next := pipelineState(t, rules), pipelineState(t, rules)
				for _, s := range []*scanState{old, next} {
					s.layers.Push(signatures.Layer{Type: signatures.TypeZIP, Size: 8192})
					switch policy {
					case "exact":
						s.maxFile, s.remaining = int64(len(payload)), int64(len(payload))
					case "file":
						s.maxFile = 3
					case "aggregate":
						s.remaining = 3
					case "zero":
						s.remaining = 0
					case "negative":
						s.remaining = -1
					case "depth", "cancel-and-depth":
						s.maxDepth = 0
					case "count":
						s.maxFiles = 0
					}
					if strings.HasPrefix(policy, "cancel") {
						ctx, cancel := context.WithCancel(context.Background())
						cancel()
						s.ctx = ctx
					}
				}
				before := bytes.Clone(payload)
				wantSig, wantErr := old.child(bytes.NewReader(payload), 0)
				sig, err := next.childBytes(payload, 0)
				if sig != wantSig || fmt.Sprint(err) != fmt.Sprint(wantErr) || old.files != next.files || old.remaining != next.remaining {
					t.Fatalf("reader=%q,%v files=%d remaining=%d; slice=%q,%v files=%d remaining=%d", wantSig, wantErr, old.files, old.remaining, sig, err, next.files, next.remaining)
				}
				if !bytes.Equal(payload, before) || next.layers.TypeAt(-1) != signatures.TypeZIP {
					t.Fatal("child changed parent bytes or ancestry")
				}
			})
		}
	}
}

func TestStoredZIPChildHashAncestryAndSizeQuirks(t *testing.T) {
	payload := []byte(formatCoverageMarker)
	archive := formatRawZipFixture(t, zip0Stored, payload, []byte("tiny"))
	for _, rules := range []map[string]string{
		{"child.hdb": fmt.Sprintf("%x:%d:Child.Hash\n", md5.Sum(payload), len(payload))},
		{"child.ldb": "Child.Ancestry;Engine:81-255,Container:CL_TYPE_ZIP,Target:0;0;" + hx(formatCoverageMarker) + "\n"},
	} {
		s := pipelineState(t, rules)
		s.layers.Push(signatures.Layer{Type: signatures.TypeZIP, Size: int64(len(archive))})
		if sig, err := s.zip(archive, 0); !strings.HasPrefix(sig, "Child.") || err != nil || s.files != 1 || s.remaining != (256<<20)-int64(len(payload)) {
			t.Fatalf("stored extent: %q %v files=%d remaining=%d", sig, err, s.files, s.remaining)
		}
	}
	// A larger declared size keeps the reference's inflate fallback, rather
	// than interpreting a misleading stored method as ordinary plain bytes.
	packed := deflated(t, payload)
	archive = formatRawZipFixture(t, zip0Stored, packed, bytes.Repeat([]byte("x"), len(packed)+100))
	s := pipelineState(t, map[string]string{"child.ndb": "Child.Marker:0:*:" + hx(formatCoverageMarker) + "\n"})
	if sig, err := s.zip(archive, 0); sig != "Child.Marker" || err != nil {
		t.Fatalf("stored/inflate fallback: %q %v", sig, err)
	}
	archive = formatRawZipFixture(t, zip0Stored, payload, payload)
	rules := map[string]string{
		"child.ndb": "Child.Marker:0:*:" + hx(formatCoverageMarker) + "\n",
		"child.sfp": fmt.Sprintf("%x:%d:Allowed.ZIP.Child\n", md5.Sum(payload), len(payload)),
	}
	s = pipelineState(t, rules)
	if sig, err := s.zip(archive, 0); sig != "" || err != nil || s.files != 1 {
		t.Fatalf("stored child allowlist: %q %v files=%d", sig, err, s.files)
	}
}

func TestCSSImagesStrictGrammarAndOrdering(t *testing.T) {
	payload := []byte(formatCoverageMarker)
	encoded := base64.StdEncoding.EncodeToString(payload)
	rules := map[string]string{"child.hdb": fmt.Sprintf("%x:%d:Child.Hash\n", md5.Sum(payload), len(payload))}
	for _, tc := range []struct {
		name, css string
		want      bool
	}{
		{"ordinary", "url(data:image/gif;base64," + encoded + ")", true},
		{"double-quotes", `url("data:image/gif;base64,` + encoded + `")`, true},
		{"embedded-double-quotes", `url(da"ta:image/gif;ba"se64,` + encoded[:8] + `"` + encoded[8:] + ")", true},
		{"spaces", "url \t( \rdata:image/gif; \tbase64 \v," + encoded + " )", true},
		{"CRLF", "url(data:image/gif;base64," + encoded[:8] + "\r\n" + encoded[8:] + ")", true},
		{"single-quotes", "url('data:image/gif;base64," + encoded + "')", false},
		{"case", "url(DATA:image/gif;base64," + encoded + ")", false},
		{"space-after-quote", `url(" data:image/gif;base64,` + encoded + `")`, false},
		{"first-semicolon", "url(data:image/gif;ordinary=x;base64," + encoded + ")", false},
		{"malformed-before-later", "url(data:image/gif;base64,%%%)url(data:image/gif;base64," + encoded + ")", true},
		{"trailing-garbage", "url(data:image/gif;base64," + encoded + "!invalid)", false},
		{"missing-padding", "url(data:image/gif;base64," + strings.TrimRight(encoded, "=") + ")", false},
		{"unbalanced", "url(data:image/gif;base64," + encoded, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sig, err := pipelineState(t, rules).styleImages([]byte(tc.css), 0)
			if (sig == "Child.Hash") != tc.want || err != nil {
				t.Fatalf("scan: %q %v want detected=%v", sig, err, tc.want)
			}
		})
	}
	// Detection in the first image must not decode a later image at all.
	s := pipelineState(t, rules)
	s.remaining = int64(len(payload))
	css := "url(data:image/gif;base64," + encoded + ")url(data:image/gif;base64," + strings.Repeat("A", 1<<20) + ")"
	if sig, err := s.styleImages([]byte(css), 0); sig != "Child.Hash" || err != nil || s.remaining != 0 || s.files != 1 {
		t.Fatalf("ordered early detection: %q %v files=%d remaining=%d", sig, err, s.files, s.remaining)
	}
}

func TestCSSImagesResourceBudgetAndCancellation(t *testing.T) {
	css := []byte(`url("data:image/gif;base64,` + strings.Repeat("A", 1<<20) + `")`)
	rules := map[string]string{"child.ndb": "Never:0:*:ff00ff00ff00ff00\n"}
	for _, policy := range []string{"file", "aggregate", "depth", "count"} {
		s := pipelineState(t, rules)
		switch policy {
		case "file":
			s.maxFile = 32
		case "aggregate":
			s.remaining = 32
		case "depth":
			s.maxDepth = 0
		case "count":
			s.maxFiles = 0
		}
		before := s.remaining
		if sig, err := s.styleImages(css, 0); sig != "" || !errors.Is(err, ErrLimit) || s.files != 0 {
			t.Fatalf("%s: %q %v files=%d", policy, sig, err, s.files)
		}
		want := before
		if policy == "file" {
			want -= 33
		} else if policy == "aggregate" {
			want = 0
		}
		if s.remaining != want {
			t.Fatalf("%s charged %d; want %d", policy, before-s.remaining, before-want)
		}
	}
	// Cancellation must also interrupt quote removal while a decoder is
	// reading, not only the earlier search for the end of url(...).
	ctx := &decodeCancelContext{Context: context.Background()}
	r := cssImageReader{data: bytes.Repeat([]byte(`A"`), 1<<19), ctx: ctx}
	decoded, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding.Strict(), &r))
	if !errors.Is(err, context.Canceled) || len(decoded) == 0 || len(decoded) >= 1<<18 {
		t.Fatalf("cancelled decode: bytes=%d err=%v", len(decoded), err)
	}
	s := pipelineState(t, rules)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	s.ctx = cancelled
	if _, err := s.styleImages(css, 0); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	// Cancellation in the final URL's header must survive a skipped URL,
	// even when there is no later loop iteration to notice it.
	s = pipelineState(t, rules)
	s.ctx = &decodeCancelContext{Context: context.Background(), checks: 1}
	if _, err := s.styleImages([]byte("url(data:x;base64,SGVsbG8=)"), 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("last header cancellation: %v", err)
	}
	// Invalid base64 never exposes a partial decoded marker as a detection;
	// its produced bytes still count as expansion work before the next URL.
	payload := []byte(formatCoverageMarker)
	s = pipelineState(t, map[string]string{"child.ndb": "Child.Marker:0:*:" + hx(formatCoverageMarker) + "\n"})
	before := s.remaining
	if sig, err := s.styleImages([]byte("url(data:x;base64,"+base64.StdEncoding.EncodeToString(payload)+"!)"), 0); sig != "" || err != nil || s.files != 0 || s.remaining >= before {
		t.Fatalf("invalid partial decoding: %q %v files=%d remaining=%d", sig, err, s.files, s.remaining)
	}
}

func FuzzCSSImageReader(f *testing.F) {
	for _, data := range [][]byte{nil, []byte(`"SGVsbG8="`), []byte("SGVs\r\nbG8="), []byte("AB=="), []byte("AA==!"), []byte(`S"GV"sbG8=`), bytes.Repeat([]byte("A"), 4096)} {
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 16384 {
			t.Skip()
		}
		filtered := bytes.ReplaceAll(data, []byte{'"'}, nil)
		r := cssImageReader{data: data, ctx: context.Background()}
		got, err := io.ReadAll(&r)
		if err != nil || !bytes.Equal(got, filtered) {
			t.Fatalf("quote filtering: %x %v; want %x", got, err, filtered)
		}
		r.pos = 0
		got, err = io.ReadAll(base64.NewDecoder(base64.StdEncoding.Strict(), &r))
		want, wantErr := base64.StdEncoding.Strict().DecodeString(string(filtered))
		if (err == nil) != (wantErr == nil) || err == nil && !bytes.Equal(got, want) {
			t.Fatalf("strict stream=%x,%v; original=%x,%v", got, err, want, wantErr)
		}
	})
}
