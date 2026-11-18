package engine

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"strconv"
	"strings"
)

type hwpMLPropertySpec struct {
	name   string
	attrs  bool
	values bool
	ignore bool
}

// The names and selected fields follow ClamAV's HWPML key table. These are
// the proved document properties only; child-object metadata is separate.
var hwpMLPropertySpecs = map[string]hwpMLPropertySpec{
	"hwpml":              {name: "HWPML", attrs: true},
	"docsummary":         {name: "DocumentProperties"},
	"title":              {name: "Title", values: true},
	"author":             {name: "Author", values: true},
	"date":               {name: "Date", values: true},
	"docsetting":         {name: "DocumentSettings"},
	"beginnumber":        {name: "BeginNumber", attrs: true},
	"caretpos":           {name: "CaretPos", attrs: true},
	"bindatastorage":     {name: "BinaryDataStorage"},
	"bindata":            {name: "BinaryData", attrs: true},
	"scriptcode":         {name: "ScriptCodeStorage", attrs: true},
	"scriptheader":       {name: "ScriptHeader", values: true},
	"scriptsource":       {name: "ScriptSource", values: true},
	"body":               {ignore: true},
	"facenamelist":       {ignore: true},
	"borderfilllist":     {ignore: true},
	"charshapelist":      {ignore: true},
	"tabdeflist":         {ignore: true},
	"numberinglist":      {ignore: true},
	"parashapelist":      {ignore: true},
	"stylelist":          {ignore: true},
	"compatibledocument": {ignore: true},
}

// hwpMLProperties collects the HWPML-specific portion of ClamAV's JSON
// metadata. It is deliberately partial: a scanner must still mark the tree
// partial for queries about extracted children and other parser fields.
func hwpMLProperties(ctx context.Context, input []byte, maxBytes int64) (map[string]any, error) {
	data, err := repairMSXMLNumericEntities(ctx, input, maxBytes)
	if err != nil {
		return nil, err
	}
	data, err = hwpMLMaskCDATA(data)
	if err != nil {
		return nil, err
	}
	const maxProperties = 1 << 16
	const maxPropertyText = 1 << 20
	root := map[string]any{}
	type frame struct {
		current map[string]any
		values  bool
	}
	var stack []frame
	var properties, textBytes, tokens int
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			if len(stack) == 0 {
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
			if len(stack) >= 20 {
				if err := skipMSXMLSubtree(ctx, dec); err != nil {
					return nil, err
				}
				continue
			}
			parent := root
			if len(stack) != 0 {
				parent = stack[len(stack)-1].current
			}
			spec := hwpMLPropertySpecs[strings.ToLower(t.Name.Local)]
			if spec.ignore {
				if err := skipMSXMLSubtree(ctx, dec); err != nil {
					return nil, err
				}
				continue
			}
			current := parent
			if spec.name != "" {
				current, _ = parent[spec.name].(map[string]any)
				if current == nil {
					current = map[string]any{}
					parent[spec.name] = current
				}
				properties++
				if spec.attrs && len(t.Attr) != 0 {
					attrs, _ := current["Attributes"].(map[string]any)
					if attrs == nil {
						attrs = map[string]any{}
						current["Attributes"] = attrs
					}
					for _, attr := range t.Attr {
						attrs[attr.Name.Local] = attr.Value
						properties++
						textBytes += len(attr.Value)
					}
				}
			}
			if properties > maxProperties || textBytes > maxPropertyText {
				return nil, ErrLimit
			}
			stack = append(stack, frame{current: current, values: spec.values})
		case xml.CharData:
			if len(stack) == 0 || !stack[len(stack)-1].values || len(bytes.TrimSpace(t)) == 0 {
				continue
			}
			value := string(t)
			textBytes += len(value)
			properties++
			if properties > maxProperties || textBytes > maxPropertyText {
				return nil, ErrLimit
			}
			var parsed any = value
			if n, err := strconv.ParseInt(strings.TrimLeft(value, " \t\r\n"), 10, 64); err == nil {
				parsed = int32(n & 0xffff)
			} else if value == "true" {
				parsed = true
			} else if value == "false" {
				parsed = false
			}
			current := stack[len(stack)-1].current
			values, _ := current["Value"].([]any)
			current["Value"] = append(values, parsed)
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, ErrMalformed
			}
			stack = stack[:len(stack)-1]
		}
	}
}
