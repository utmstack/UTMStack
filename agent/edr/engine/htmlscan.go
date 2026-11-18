package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
	htmltoken "golang.org/x/net/html"
)

// maxHTMLTextView is the largest object whose text view is matched.
const maxHTMLTextView = 8 << 20

// htmlViews matches an HTML object's normalized views, in the reference's
// order, stopping at the first detection: the page without comments and the
// text outside tags as HTML (typed rule root 3 and root 0; the text view
// only for objects up to 8 MiB), the scripts as HTML and then as ASCII text
// (root 7 and root 0, not text-normalized again), then each data: URI file
// and each base64 image of a style block as an object of its own. UTF-16
// HTML is converted first. Script Encoder forms are normalized in the same
// chunked state machine; malformed forms and JavaScript transforms that
// normalizeJS cannot reproduce remain incomplete.
func (s *scanState) htmlViews(data []byte, depth int, typ signatures.FileType) (string, error) {
	src := data
	if typ == signatures.TypeHTMLUTF16 {
		src = htmlUTF16ToASCII(data)
	}
	v := normalizeHTML(src)
	js, jsUnsupported := normalizeJS(v.scripts)
	if v.unsupported != "" && v.unsupported != "script-encoder-truncated" {
		// The page views are uncertain, but an encoded script's own decoded
		// object can still contain a decisive signature.
		if sig, err := s.encodedHTMLScripts(src, depth); sig != "" || err != nil {
			return sig, err
		}
		return "", unknownRepresentation(signatures.UnknownHTML)
	}
	if depth+1 > s.maxDepth {
		return "", ErrLimit
	}
	var incomplete error
	if v.unsupported != "" {
		incomplete = unknownRepresentation(signatures.UnknownHTML)
	}
	if jsUnsupported {
		// Script uncertainty does not change the independently normalized
		// page/text or embedded files. Skip the uncertain script view only.
		incomplete = unknownRepresentation(signatures.UnknownHTML)
	}
	match := func(b []byte, t signatures.FileType) string {
		if len(b) == 0 {
			return "" // empty views are not matched
		}
		s.layers.Push(signatures.Layer{Type: t, Size: int64(len(b)), Normalized: true})
		defer s.layers.Pop()
		res, err := s.db.MatchObject(s.ctx, b, signatures.MatchContext{Type: t, Layers: &s.layers, Metadata: s.representationMetadata(), MetadataPartial: s.collectMetadata, ScanExtractedTyped: s.bytecodeChildScanner(depth + 1), BytecodeOptions: s.bytecodeEngineOptions(), PreclassMatched: s.preclassMatched})
		if err != nil && incomplete == nil {
			incomplete = err
		}
		return res.Name
	}
	if sig := match(v.noComment, signatures.TypeHTML); sig != "" {
		return sig, nil
	}
	if len(data) <= maxHTMLTextView {
		if sig := match(v.noTags, signatures.TypeHTML); sig != "" {
			return sig, nil
		}
	}
	if len(v.scripts) > 0 && !jsUnsupported {
		if sig := match(js, signatures.TypeHTML); sig != "" {
			return sig, nil
		}
		if sig := match(js, signatures.TypeTextASCII); sig != "" {
			return sig, nil
		}
	}
	for i := range v.rfc2397 {
		p := &v.rfc2397[i]
		sig, err := s.visitObject(p.compatibility, depth+1, objectOptions{htmlData: p})
		if sig != "" {
			return sig, nil
		}
		if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	if sig, err := s.styleImages(v.style, depth); sig != "" {
		return sig, nil
	} else if err != nil && incomplete == nil {
		incomplete = err
	}
	if sig, err := s.entityDataURIs(src, depth); sig != "" {
		return sig, nil
	} else if err != nil && incomplete == nil {
		incomplete = err
	}
	return "", incomplete
}

func (s *scanState) encodedHTMLScripts(data []byte, depth int) (string, error) {
	z := htmltoken.NewTokenizer(contextReader{s.ctx, bytes.NewReader(data)})
	z.SetMaxBuf(maxHTMLNormalize)
	encoded := false
	for {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		switch z.Next() {
		case htmltoken.ErrorToken:
			if err := z.Err(); err != io.EOF {
				return "", err
			}
			return "", nil
		case htmltoken.StartTagToken:
			name, more := z.TagName()
			encoded = false
			if !bytes.Equal(name, []byte("script")) {
				continue
			}
			for more {
				var key, value []byte
				key, value, more = z.TagAttr()
				if bytes.Equal(key, []byte("language")) &&
					(bytes.EqualFold(value, []byte("jscript.encode")) || bytes.EqualFold(value, []byte("vbscript.encode"))) {
					encoded = true
				}
			}
		case htmltoken.TextToken:
			if !encoded {
				continue
			}
			encoded = false
			if !bytes.Contains(z.Raw(), []byte("#@~^")) {
				continue
			}
			decoded, err := decodeScrEnc(s.ctx, z.Raw())
			if decoded == nil {
				if err == ErrLimit || s.ctx.Err() != nil {
					return "", err
				}
				continue // the page stays incomplete
			}
			wrapper := make([]byte, 0, len(decoded)+17)
			wrapper = append(wrapper, "<script>"...)
			wrapper = append(wrapper, decoded...)
			wrapper = append(wrapper, "</script>"...)
			if sig, err := s.childBytes(wrapper, depth); sig != "" || err != nil {
				return sig, err
			}
		case htmltoken.EndTagToken:
			encoded = false
		}
	}
}

// Browsers resolve character references inside attribute values before using
// a URL. The reference-compatible normalizer stops its data envelope at '&'.
// Inspect this additional, independently decoded view without changing its
// original page/envelope bytes or treating escaped quotes as markup syntax.
func (s *scanState) entityDataURIs(data []byte, depth int) (string, error) {
	if bytes.IndexByte(data, '&') < 0 {
		return "", nil
	}
	z := htmltoken.NewTokenizer(contextReader{s.ctx, bytes.NewReader(data)})
	z.SetMaxBuf(maxHTMLNormalize)
	var incomplete error
	for {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		switch z.Next() {
		case htmltoken.ErrorToken:
			if err := z.Err(); err != io.EOF {
				if err == htmltoken.ErrBufferExceeded {
					return "", ErrLimit
				}
				return "", err
			}
			return "", incomplete
		case htmltoken.StartTagToken, htmltoken.SelfClosingTagToken:
			raw := z.Raw()
			if bytes.IndexByte(raw, '&') < 0 {
				continue
			}
			// TagAttr unescapes its values in the tokenizer's buffer. Preserve
			// this tag's original spelling for the already-scanned comparison.
			raw = bytes.Clone(raw)
			_, more := z.TagName()
			for more {
				var key, value []byte
				key, value, more = z.TagAttr()
				switch string(key) {
				case "src", "href", "data", "poster", "background", "action", "formaction":
				default:
					continue
				}
				if len(value) < 5 || !bytes.EqualFold(value[:5], []byte("data:")) || bytes.Contains(raw, value) {
					continue
				}
				comma := bytes.IndexByte(value, ',')
				if comma < 5 || comma >= 5+maxDataURIHeader {
					continue
				}
				valid := true
				for _, c := range value[5:comma] {
					if !dataURIHeaderByte(c) {
						valid = false
						break
					}
				}
				if !valid {
					continue
				}
				sig, err := s.dataURIContent(value[5:comma], value[comma+1:], depth)
				if sig != "" {
					return sig, nil
				}
				if err != nil {
					incomplete = err
					if err == ErrLimit {
						return "", err
					}
				}
			}
		}
	}
}

// styleImages visits one decoded base64 image at a time. It preserves the
// reference's CSS URL grammar and strict base64 rules while bounding each
// decode by the scan budget and checking cancellation during decoding.
// Eligible URLs are url( ... ) parameters, trimmed and without double
// quotes, that start with "data:" and have, after their first ';', optional
// spaces, "base64", optional
// spaces and ','. Text that is not strict base64 is skipped.
func (s *scanState) styleImages(style []byte, depth int) (string, error) {
	var incomplete error
	for i := 0; i < len(style); {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		k := bytes.Index(style[i:], []byte("url"))
		if k < 0 {
			break
		}
		p := i + k + 3
		i = p
		for p < len(style) && isHTMLSpace(style[p]) {
			if p&4095 == 0 && s.ctx.Err() != nil {
				return "", s.ctx.Err()
			}
			p++
		}
		if p >= len(style) || style[p] != '(' {
			continue
		}
		nesting, start := 1, p+1
		end := start
		for end < len(style) && nesting > 0 {
			if end&4095 == 0 {
				if err := s.ctx.Err(); err != nil {
					return "", err
				}
			}
			switch style[end] {
			case '(':
				nesting++
			case ')':
				nesting--
			}
			end++
		}
		if nesting != 0 {
			break
		}
		param := bytes.TrimSpace(style[start : end-1])
		i = end
		r := cssImageReader{data: param, ctx: s.ctx}
		if !r.literal("data:") {
			continue
		}
		semi := bytes.IndexByte(param[r.pos:], ';')
		if semi < 0 {
			continue
		}
		r.pos += semi + 1
		r.skipSpace()
		if !r.literal("base64") {
			continue
		}
		r.skipSpace()
		if !r.literal(",") {
			continue
		}
		decoded, err := s.readChild(base64.NewDecoder(base64.StdEncoding.Strict(), &r), depth)
		if err != nil {
			if s.ctx.Err() != nil {
				return "", s.ctx.Err()
			}
			if !errors.Is(err, ErrMalformed) && incomplete == nil {
				incomplete = err
			}
			continue // malformed strict base64 remains a skipped CSS image
		}
		if len(decoded) == 0 {
			continue
		}
		if sig, err := s.visit(decoded, depth+1); sig != "" {
			return sig, nil
		} else if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	if err := s.ctx.Err(); err != nil {
		return "", err
	}
	return "", incomplete
}

// The reference removes every double quote before parsing a CSS URL. Filter
// them while reading so a quoted image cannot require an input-sized copy
// before its decoded-byte budget has even been checked.
type cssImageReader struct {
	data []byte
	pos  int
	ctx  context.Context
	err  error
}

func (r *cssImageReader) next() (byte, bool) {
	for r.pos < len(r.data) {
		if r.pos&4095 == 0 {
			if r.err = r.ctx.Err(); r.err != nil {
				return 0, false
			}
		}
		c := r.data[r.pos]
		r.pos++
		if c != '"' {
			return c, true
		}
	}
	return 0, false
}

func (r *cssImageReader) literal(s string) bool {
	for i := range len(s) {
		if c, ok := r.next(); !ok || c != s[i] {
			return false
		}
	}
	return true
}

func (r *cssImageReader) skipSpace() {
	for {
		start := r.pos
		c, ok := r.next()
		if !ok {
			return
		}
		if !isHTMLSpace(c) {
			r.pos = start
			return
		}
	}
}

func (r *cssImageReader) Read(p []byte) (int, error) {
	for i := range len(p) {
		c, ok := r.next()
		if !ok {
			if r.err != nil {
				return i, r.err
			}
			return i, io.EOF
		}
		p[i] = c
	}
	return len(p), nil
}
