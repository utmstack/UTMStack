package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"io"
	"strings"
)

// The reference's XDP reader examines base64 chunk elements and only scans
// their decoded content when a PDF header occurs in the first 1,028 bytes.
// A decoded non-PDF chunk ends its traversal, including later chunks.
func (s *scanState) xdp(data []byte, depth int) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var chunk strings.Builder
	var chunkDepth, nesting, depthXML int
	var nested bool
	for {
		token, err := dec.Token()
		if err == io.EOF {
			return "", nil
		}
		if err != nil {
			// The reference treats an invalid XML view as a non-XDP object.
			return "", nil
		}
		switch t := token.(type) {
		case xml.StartElement:
			depthXML++
			if depthXML > 64 {
				return "", ErrLimit
			}
			if chunkDepth != 0 {
				nesting++
				nested = true
			} else if t.Name.Local == "chunk" {
				chunkDepth = depthXML
				chunk.Reset()
				nested = false
			}
		case xml.CharData:
			if chunkDepth != 0 && nesting == 0 {
				chunk.Write(t)
			}
		case xml.EndElement:
			if chunkDepth != 0 {
				if depthXML == chunkDepth {
					if !nested {
						max, err := s.childBudget(depth)
						if err != nil {
							return "", err
						}
						decoded, err := decodeXDPChunk(chunk.String(), max)
						if err != nil {
							return "", err
						}
						if !xdpPDF(decoded) {
							return "", nil
						}
						if sig, err := s.childBytes(decoded, depth); sig != "" || err != nil {
							return sig, err
						}
					}
					chunkDepth = 0
					chunk.Reset()
				} else if nesting > 0 {
					nesting--
				}
			}
			depthXML--
		}
	}
}

func decodeXDPChunk(text string, max int64) ([]byte, error) {
	// XDP chunks commonly wrap base64 over many XML lines. Strip XML
	// whitespace before decoding; malformed base64 is not an extracted file.
	packed := make([]byte, 0, len(text))
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case ' ', '\t', '\n', '\r':
		default:
			packed = append(packed, text[i])
			if int64(len(packed)) > max/3*4+8 {
				return nil, ErrLimit
			}
		}
	}
	if int64(base64.StdEncoding.DecodedLen(len(packed))) > max+2 {
		return nil, ErrLimit
	}
	decoded, err := base64.StdEncoding.DecodeString(string(packed))
	if err != nil {
		// ClamAV's decoder can return a non-PDF byte string for malformed
		// input, which ends the chunk walk. Ending it here preserves that
		// observed behavior for damaged chunks.
		return nil, nil
	}
	return decoded, nil
}

func xdpPDF(decoded []byte) bool {
	if len(decoded) <= 5 {
		return false
	}
	for i := 0; i < min(1028, len(decoded)-5); i++ {
		if len(decoded)-i >= 5 && bytes.EqualFold(decoded[i:i+5], []byte("%PDF-")) {
			return true
		}
	}
	return false
}
