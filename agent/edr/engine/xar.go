package engine

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/ulikunitz/xz/lzma"
)

// XAR stores a zlib-compressed XML TOC followed by a heap of independently
// encoded extents. The reference scans the TOC, subdocuments, and every data
// or extended-attribute extent, regardless of checksum mismatches.
const maxXARTOC = 16 << 20

type xarExtent struct {
	offset, length, size             uint64
	encoding                         string
	haveOffset, haveLength, haveSize bool
}

func (s *scanState) xar(data []byte, depth int) (string, error) {
	if len(data) < 28 || binary.BigEndian.Uint16(data[4:]) < 28 {
		return "", ErrMalformed
	}
	header := uint64(binary.BigEndian.Uint16(data[4:]))
	compressed := binary.BigEndian.Uint64(data[8:])
	if header > uint64(len(data)) || compressed > uint64(len(data))-header {
		return "", ErrMalformed
	}
	reader, err := zlib.NewReader(bytes.NewReader(data[header : header+compressed]))
	if err != nil {
		return "", ErrMalformed
	}
	toc, err := s.readChild(io.LimitReader(reader, maxXARTOC+1), depth)
	closeErr := reader.Close()
	if err != nil {
		return "", err
	}
	if len(toc) > maxXARTOC {
		s.spend(len(toc))
		return "", ErrLimit
	}
	if closeErr != nil {
		return "", ErrMalformed
	}
	if len(toc) == 0 {
		return "", ErrMalformed
	}
	// ClamAV scans the XML as its own object before walking extents.
	var incomplete error
	if sig, err := s.visit(toc, depth+1); sig != "" {
		return sig, nil
	} else if err != nil {
		incomplete = err
	}
	extents, subdocs, parseErr := parseXARTOC(toc, s.maxFiles-s.files)
	if parseErr != nil && incomplete == nil {
		incomplete = parseErr
	}
	for _, subdoc := range subdocs {
		if sig, err := s.visit(subdoc, depth+1); sig != "" {
			return sig, nil
		} else if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	heap := data[header+compressed:]
	for _, extent := range extents {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if extent.offset > uint64(len(heap)) || extent.length > uint64(len(heap))-extent.offset {
			if incomplete == nil {
				incomplete = ErrMalformed
			}
			continue
		}
		packed := heap[extent.offset : extent.offset+extent.length]
		var sig string
		var err error
		switch extent.encoding {
		case "application/x-gzip":
			var r io.ReadCloser
			r, err = zlib.NewReader(bytes.NewReader(packed))
			if err == nil {
				sig, err = s.child(r, depth)
				r.Close()
			} else {
				err = ErrMalformed
			}
		case "application/x-lzma":
			sig, err = s.xarLZMA(packed, depth)
		default:
			// ClamAV sends stored, bzip2, XZ and unknown encodings through
			// its ordinary recursive scanner without decoding them here.
			sig, err = s.visit(packed, depth+1)
		}
		if sig != "" {
			return sig, nil
		}
		if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	return "", incomplete
}

func (s *scanState) xarLZMA(packed []byte, depth int) (string, error) {
	if len(packed) < 13 || packed[0] >= 225 {
		return "", ErrMalformed
	}
	if binary.LittleEndian.Uint32(packed[1:]) > maxCodecDictionary {
		return "", ErrLimit
	}
	r, err := (lzma.ReaderConfig{DictCap: maxCodecDictionary}).NewReader(contextReader{s.ctx, bytes.NewReader(packed)})
	if err != nil {
		var dictionary *lzma.ErrDictSize
		if errors.As(err, &dictionary) {
			return "", ErrLimit
		}
		return "", ErrMalformed
	}
	return s.child(r, depth)
}

// parseXARTOC uses a bounded token stream so deeply nested attacker input
// cannot create a recursive object tree. Parsed extents remain scannable if
// a later entry is malformed; the archive then remains incomplete.
func parseXARTOC(toc []byte, maxEntries int) ([]xarExtent, [][]byte, error) {
	dec := xml.NewDecoder(bytes.NewReader(toc))
	var extents []xarExtent
	var subdocs [][]byte
	var active *xarExtent
	var field string
	var value strings.Builder
	fieldOverflow := false
	var dataDepth, fieldDepth, subDepth, tocDepth, depth int
	var subStart int64
	root, seenTOC := false, false
	var incomplete error
	for {
		before := dec.InputOffset()
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return extents, subdocs, ErrMalformed
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if depth > 64 {
				return extents, subdocs, ErrLimit
			}
			if depth == 1 {
				root = t.Name.Local == "xar"
			} else if depth == 2 && t.Name.Local == "toc" && !seenTOC {
				seenTOC = true
				tocDepth = depth
			}
			if !root {
				continue
			}
			if t.Name.Local == "subdoc" && depth == 2 && subDepth == 0 {
				subDepth, subStart = depth, dec.InputOffset()
			}
			if tocDepth == 0 || depth <= tocDepth {
				continue
			}
			if active == nil && (t.Name.Local == "data" || t.Name.Local == "ea") {
				active, dataDepth = &xarExtent{}, depth
			} else if active != nil && depth == dataDepth+1 {
				switch t.Name.Local {
				case "offset", "length", "size":
					field, fieldDepth = t.Name.Local, depth
					value.Reset()
					fieldOverflow = false
				case "encoding":
					for _, attr := range t.Attr {
						if attr.Name.Local == "style" {
							active.encoding = attr.Value
						}
					}
				}
			}
		case xml.CharData:
			if field != "" && depth == fieldDepth {
				if value.Len()+len(t) <= 32 {
					value.Write(t)
				} else {
					fieldOverflow = true
				}
			}
		case xml.EndElement:
			if field != "" && depth == fieldDepth && t.Name.Local == field {
				n, err := strconv.ParseUint(strings.TrimSpace(value.String()), 10, 64)
				if err != nil || fieldOverflow {
					incomplete = ErrMalformed
				} else {
					switch field {
					case "offset":
						active.offset, active.haveOffset = n, true
					case "length":
						active.length, active.haveLength = n, true
					case "size":
						active.size, active.haveSize = n, true
					}
				}
				field = ""
			}
			if active != nil && depth == dataDepth && (t.Name.Local == "data" || t.Name.Local == "ea") {
				if active.haveOffset && active.haveLength && active.haveSize {
					extents = append(extents, *active)
					if len(extents) > maxEntries {
						return extents, subdocs, ErrLimit
					}
				} else {
					incomplete = ErrMalformed
				}
				active = nil
			}
			if subDepth != 0 && depth == subDepth && t.Name.Local == "subdoc" {
				if before >= subStart && before <= int64(len(toc)) {
					subdocs = append(subdocs, toc[subStart:before])
					if len(subdocs)+len(extents) > maxEntries {
						return extents, subdocs, ErrLimit
					}
				}
				subDepth = 0
			}
			if tocDepth != 0 && depth == tocDepth && t.Name.Local == "toc" {
				tocDepth = 0
			}
			depth--
		case xml.Directive:
			return extents, subdocs, ErrMalformed
		}
	}
	if !root || !seenTOC || depth != 0 || active != nil || subDepth != 0 {
		return extents, subdocs, ErrMalformed
	}
	return extents, subdocs, incomplete
}
