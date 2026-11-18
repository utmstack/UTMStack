package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"io"
	"strconv"
	"strings"
)

var msxmlPropertyNames = map[string]string{
	"author": "Author", "lastauthor": "LastAuthor", "revision": "Revision",
	"totaltime": "TotalTime", "created": "Created", "lastsaved": "LastSaved",
	"pages": "Pages", "words": "Words", "characters": "Characters",
	"lines": "Lines", "paragraph": "Paragraph", "characterswithspaces": "CharactersWithSpaces",
	"version": "Version", "allowpng": "AllowPNG",
}

// msxmlProperties builds an Office-specific subset of ClamAV's property tree.
// This is available for explicitly collected metadata; ordinary EDR scans do
// not expose it to bytecode because ClamAV's default scan disables metadata
// collection. Malformed or over-budget input supplies no tree.
func msxmlProperties(ctx context.Context, input []byte, maxBytes int64, fileType string) (map[string]any, error) {
	data, err := repairMSXMLNumericEntities(ctx, input, maxBytes)
	if err != nil {
		return nil, err
	}
	const maxProperties = 1 << 16
	const maxPropertyText = 1 << 20
	root := map[string]any{
		"Magic": "CLAMJSONv0", "FileSize": int64(len(input)),
		"FileType": fileType, "RootFileType": fileType,
	}
	type frame struct {
		current map[string]any
		value   map[string]any
	}
	var stack []frame
	dec := xml.NewDecoder(bytes.NewReader(data))
	var depth, ignoredDepth, tokens, properties, textBytes int
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			if depth == 0 {
				return root, nil
			}
			return nil, ErrMalformed
		}
		if err != nil {
			return nil, ErrMalformed
		}
		tokens++
		if tokens&1023 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		switch t := tok.(type) {
		case xml.StartElement:
			// The reference's shared XML reader skips an entire element at
			// recursion level 20, including its metadata properties.
			if depth >= 20 {
				if err := skipMSXMLSubtree(ctx, dec); err != nil {
					return nil, err
				}
				continue
			}
			depth++
			parent := root
			if len(stack) > 0 {
				parent = stack[len(stack)-1].current
			}
			entry := frame{current: parent}
			key := strings.ToLower(t.Name.Local)
			if ignoredDepth != 0 || key == "fonts" || key == "styles" {
				if ignoredDepth == 0 {
					ignoredDepth = depth
				}
				stack = append(stack, entry)
				continue
			}
			switch key {
			case "worddocument", "workbook", "documentproperties", "bindata":
				name := map[string]string{"worddocument": "WordDocument", "workbook": "Workbook", "documentproperties": "DocumentProperties", "bindata": "BinaryData"}[key]
				object, _ := root[name].(map[string]any)
				if object == nil {
					object = make(map[string]any)
					root[name] = object
				}
				entry.current = object
				properties++
				if key == "bindata" {
					count, _ := object["Count"].(int32)
					object["Count"] = count + 1
				} else if key == "worddocument" || key == "workbook" {
					if len(t.Attr) != 0 {
						attrs, _ := object["Attributes"].(map[string]any)
						if attrs == nil {
							attrs = make(map[string]any)
							object["Attributes"] = attrs
						}
						for _, attr := range t.Attr {
							attrs[attr.Name.Local] = attr.Value
							properties++
							textBytes += len(attr.Value)
						}
					}
				}
			default:
				if name := msxmlPropertyNames[key]; name != "" {
					object, _ := parent[name].(map[string]any)
					if object == nil {
						object = make(map[string]any)
						parent[name] = object
					}
					entry.current, entry.value = object, object
					properties++
				}
			}
			if properties > maxProperties || textBytes > maxPropertyText {
				return nil, ErrLimit
			}
			stack = append(stack, entry)
		case xml.CharData:
			if ignoredDepth != 0 || len(stack) == 0 || stack[len(stack)-1].value == nil {
				continue
			}
			text := string(t)
			textBytes += len(text)
			properties++
			if properties > maxProperties || textBytes > maxPropertyText {
				return nil, ErrLimit
			}
			value := any(text)
			if n, err := strconv.ParseInt(strings.TrimLeft(text, " \t\r\n"), 10, 64); err == nil {
				value = int32(n & 0xffff)
			} else if text == "true" {
				value = true
			} else if text == "false" {
				value = false
			}
			object := stack[len(stack)-1].value
			values, _ := object["Value"].([]any)
			object["Value"] = append(values, value)
		case xml.EndElement:
			if ignoredDepth == depth {
				ignoredDepth = 0
			}
			if len(stack) == 0 {
				return nil, ErrMalformed
			}
			stack = stack[:len(stack)-1]
			depth--
		}
	}
}

// msxml scans base64 binData elements in Microsoft Office XML documents.
// The reference applies this extraction to Word and Workbook XML, then scans
// each decoded element as an ordinary child object.
func (s *scanState) msxml(data []byte, depth int) (string, error) {
	var err error
	data, err = repairMSXMLNumericEntities(s.ctx, data, s.maxFile)
	if err != nil {
		return "", err
	}
	dec := xml.NewDecoder(bytes.NewReader(data))
	var encoded strings.Builder
	var binaryDepth, ignoredDepth, xmlDepth, nested, tokens int
	var hadNested bool
	var incomplete error
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return "", incomplete
		}
		if err != nil {
			// The reference repairs some malformed numeric entities before
			// parsing. Do not call a typed document clean if this reader stops.
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
			if xmlDepth >= 20 {
				if binaryDepth != 0 {
					hadNested = true
				}
				if err := skipMSXMLSubtree(s.ctx, dec); err != nil {
					return "", err
				}
				continue
			}
			xmlDepth++
			if ignoredDepth != 0 {
				continue
			}
			if binaryDepth == 0 && (strings.EqualFold(t.Name.Local, "Fonts") || strings.EqualFold(t.Name.Local, "Styles")) {
				ignoredDepth = xmlDepth
				continue
			}
			if binaryDepth != 0 {
				nested++
				hadNested = true
			} else if strings.EqualFold(t.Name.Local, "binData") {
				binaryDepth = xmlDepth
				encoded.Reset()
				hadNested = false
			}
		case xml.CharData:
			if ignoredDepth == 0 && binaryDepth != 0 && nested == 0 {
				encoded.Write(t)
			}
		case xml.EndElement:
			if ignoredDepth != 0 {
				if xmlDepth == ignoredDepth {
					ignoredDepth = 0
				}
				xmlDepth--
				continue
			}
			if binaryDepth != 0 {
				if xmlDepth == binaryDepth {
					if !hadNested {
						decoded, err := s.msxmlDecode(encoded.String(), depth)
						if err != nil && incomplete == nil {
							incomplete = err
						} else if len(decoded) > 0 {
							if sig, err := s.childBytes(decoded, depth); sig != "" {
								return sig, nil
							} else if err != nil && incomplete == nil {
								incomplete = err
							}
						}
					}
					binaryDepth = 0
					encoded.Reset()
				} else if nested > 0 {
					nested--
				}
			}
			xmlDepth--
		}
	}
}

// skipMSXMLSubtree consumes an element that the shared reference parser would
// skip at its recursion limit. Check cancellation during the skip as well:
// xml.Decoder.Skip does not offer that hook for attacker-controlled XML.
func skipMSXMLSubtree(ctx context.Context, dec *xml.Decoder) error {
	depth := 1
	for tokens := 0; depth != 0; tokens++ {
		if tokens&1023 == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		tok, err := dec.Token()
		if err != nil {
			return ErrMalformed
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}

// ClamAV's Office XML reader inserts the missing terminator after decimal
// and lowercase-x hexadecimal numeric entities. Keep the original slice
// when no repair is needed and cap the expanded representation.
func repairMSXMLNumericEntities(ctx context.Context, data []byte, maxBytes int64) ([]byte, error) {
	const (
		plain = iota
		ampersand
		number
		decimal
		hexadecimal
	)
	state := plain
	var out []byte
	for i, b := range data {
		if i&65535 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		insert := false
		switch state {
		case plain:
			if b == '&' {
				state = ampersand
			}
		case ampersand:
			if b == '#' {
				state = number
			} else {
				state = plain
			}
		case number:
			switch {
			case b == 'x':
				state = hexadecimal
			case b >= '0' && b <= '9':
				state = decimal
			default:
				state = plain
			}
		case decimal:
			if b < '0' || b > '9' {
				insert = b != ';'
				state = plain
			}
		case hexadecimal:
			if !((b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')) {
				insert = b != ';'
				state = plain
			}
		}
		if insert {
			if out == nil {
				out = make([]byte, 0, min(len(data)+64, int(maxBytes)))
				out = append(out, data[:i]...)
			}
			if int64(len(out))+2 > maxBytes {
				return nil, ErrLimit
			}
			out = append(out, ';')
		}
		if out != nil {
			if int64(len(out)) >= maxBytes {
				return nil, ErrLimit
			}
			out = append(out, b)
		}
	}
	if out == nil {
		return data, nil
	}
	return out, nil
}

func (s *scanState) msxmlDecode(text string, depth int) ([]byte, error) {
	max, err := s.childBudget(depth)
	if err != nil {
		return nil, err
	}
	packed := make([]byte, 0, min(len(text), int(max/3*4+8)))
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
		return nil, nil
	}
	return decoded, nil
}
