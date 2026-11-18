package engine

import (
	"bytes"
	"compress/flate"
	"encoding/xml"
	"io"
	"strings"
)

// hwpMLScanElement marks the XML elements whose text is a separately scanned
// object. The document body and its formatting tables are intentionally
// omitted: the reference's HWPML reader skips those subtrees.
func hwpMLScanElement(local string) bool {
	switch strings.ToLower(local) {
	case "bindata", "scriptheader", "scriptsource":
		return true
	}
	return false
}

func hwpMLIgnoredElement(local string) bool {
	switch strings.ToLower(local) {
	case "body", "facenamelist", "borderfilllist", "charshapelist", "tabdeflist",
		"numberinglist", "parashapelist", "stylelist", "compatibledocument":
		return true
	}
	return false
}

// hwpML scans HWPML BinData and script text nodes. Native ClamAV 1.5.4
// scans unencoded text successfully, but its Base64 path writes a decoded
// temporary file without rewinding it, so that path scans zero bytes. Decode
// and scan the actual child here: missing malware in that known upstream edge
// is not a feature the replacement should preserve.
func (s *scanState) hwpML(data []byte, depth int) (string, error) {
	repaired, err := repairMSXMLNumericEntities(s.ctx, data, s.maxFile)
	if err != nil {
		return "", err
	}
	repaired, err = hwpMLMaskCDATA(repaired)
	if err != nil {
		return "", err
	}
	dec := xml.NewDecoder(bytes.NewReader(repaired))
	type frame struct {
		scan bool
		attr []xml.Attr
	}
	var stack []frame
	var ignored, tokens int
	var incomplete error
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			if len(stack) != 0 && incomplete == nil {
				incomplete = ErrMalformed
			}
			return "", incomplete
		}
		if err != nil {
			if incomplete == nil {
				incomplete = ErrMalformed
			}
			return "", incomplete
		}
		tokens++
		if tokens&1023 == 0 {
			if err := s.ctx.Err(); err != nil {
				return "", err
			}
		}
		switch t := tok.(type) {
		case xml.StartElement:
			// The shared ClamAV XML reader starts at level zero and skips an
			// element (including its descendants) at level 20. A child-only
			// signature must not match a representation the reference never
			// exposes. The bounded skip does not grow our stack.
			if len(stack) >= 20 {
				if err := skipMSXMLSubtree(s.ctx, dec); err != nil {
					return "", err
				}
				continue
			}
			f := frame{}
			if ignored == 0 && hwpMLIgnoredElement(t.Name.Local) {
				ignored = len(stack) + 1
			}
			if ignored == 0 && hwpMLScanElement(t.Name.Local) {
				f.scan, f.attr = true, t.Attr
			}
			stack = append(stack, f)
		case xml.CharData:
			if ignored != 0 || len(stack) == 0 || !stack[len(stack)-1].scan ||
				len(bytes.TrimSpace(t)) == 0 {
				continue
			}
			if sig, err := s.hwpMLChild([]byte(t), stack[len(stack)-1].attr, depth); sig != "" {
				return sig, nil
			} else if err != nil && incomplete == nil {
				incomplete = err
			}
		case xml.EndElement:
			if len(stack) == 0 {
				return "", ErrMalformed
			}
			if ignored == len(stack) {
				ignored = 0
			}
			stack = stack[:len(stack)-1]
		}
	}
}

// Go's XML decoder presents CDATA as CharData, whereas the reference's XML
// reader calls HWPML's scan callback only for ordinary text nodes. Mask the
// CDATA body in the parser view so its contents cannot spuriously become a
// decoded child; the original object remains available to raw signatures.
func hwpMLMaskCDATA(data []byte) ([]byte, error) {
	at := bytes.Index(data, []byte("<![CDATA["))
	if at < 0 {
		return data, nil
	}
	out := bytes.Clone(data)
	for at >= 0 {
		start := at + len("<![CDATA[")
		end := bytes.Index(out[start:], []byte("]]>"))
		if end < 0 {
			return nil, ErrMalformed
		}
		end += start
		for i := start; i < end; i++ {
			if out[i] != '\r' && out[i] != '\n' {
				out[i] = ' '
			}
		}
		at = bytes.Index(out[end+3:], []byte("<![CDATA["))
		if at >= 0 {
			at += end + 3
		}
	}
	return out, nil
}

func (s *scanState) hwpMLChild(text []byte, attrs []xml.Attr, depth int) (string, error) {
	encoded, compressed := false, false
	for i, attr := range attrs {
		if i >= 20 { // the reference's callback attribute array is fixed
			break
		}
		switch attr.Name.Local {
		case "Encoding":
			if attr.Value != "Base64" {
				return s.childBytes(text, depth)
			}
			encoded = true
		case "Compress":
			compressed = attr.Value != "false"
		}
	}
	child := text
	if encoded {
		decoded, err := s.msxmlDecode(string(text), depth)
		if err != nil {
			return "", err
		}
		if decoded == nil {
			return s.childBytes(text, depth)
		}
		child = decoded
	}
	if !compressed {
		return s.childBytes(child, depth)
	}
	max, err := s.childBudget(depth)
	if err != nil {
		return "", err
	}
	r := flate.NewReader(bytes.NewReader(child))
	decoded, readErr := io.ReadAll(io.LimitReader(contextReader{s.ctx, r}, max+1))
	r.Close()
	if int64(len(decoded)) > max {
		return "", ErrLimit
	}
	if sig, err := s.childBytes(decoded, depth); sig != "" {
		return sig, nil
	} else if err != nil {
		return "", err
	}
	if readErr != nil {
		return "", ErrMalformed
	}
	return "", nil
}
