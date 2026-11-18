package engine

import (
	"context"
	"encoding/binary"
)

// oneNoteWebCandidates follows MS-ONE page content to embedded-file nodes.
// It returns candidates only after the complete reachable page tree parses;
// callers must not scan a partial result after an error.
type oneNoteWalker struct {
	ctx   context.Context
	graph *fssGraph
	seen  map[fssCell]map[fssID]bool
	steps int
	blobs []fssID
}

func (w *oneNoteWalker) object(space fssSpace, id fssID, jcid uint32) (fssObject, oneFields, error) {
	w.steps++
	if w.steps > 100_000 {
		return fssObject{}, oneFields{}, ErrLimit
	}
	if w.steps&1023 == 0 {
		if err := w.ctx.Err(); err != nil {
			return fssObject{}, oneFields{}, err
		}
	}
	obj, ok := space.objects[id]
	if !ok || (jcid != 0 && obj.jcID != jcid) {
		return fssObject{}, oneFields{}, ErrMalformed
	}
	fields, err := oneParseObjectFields(obj.props)
	if err != nil {
		return fssObject{}, oneFields{}, err
	}
	return obj, fields, nil
}
func oneRequired(f oneFields, id uint32, kind uint32) error {
	field, _, ok := f.find(id)
	if !ok || field.kind != kind {
		return ErrMalformed
	}
	return nil
}
func oneOptional(f oneFields, id uint32, kind uint32) error {
	field, _, ok := f.find(id)
	if ok && field.kind != kind {
		return ErrMalformed
	}
	return nil
}
func oneRequiredUTF16(f oneFields, id uint32) error {
	field, _, ok := f.find(id)
	if !ok {
		return ErrMalformed
	}
	return oneUTF16Field(field)
}
func oneOptionalUTF16(f oneFields, id uint32) error {
	field, _, ok := f.find(id)
	if !ok {
		return nil
	}
	return oneUTF16Field(field)
}
func oneUTF16Field(field oneField) error {
	if field.kind != 7 || len(field.raw)%2 != 0 {
		return ErrMalformed
	}
	for i := 0; i < len(field.raw); i += 2 {
		u := binary.LittleEndian.Uint16(field.raw[i:])
		if u >= 0xd800 && u <= 0xdbff {
			if i+4 > len(field.raw) {
				return ErrMalformed
			}
			v := binary.LittleEndian.Uint16(field.raw[i+2:])
			if v < 0xdc00 || v > 0xdfff {
				return ErrMalformed
			}
			i += 2
		} else if u >= 0xdc00 && u <= 0xdfff {
			return ErrMalformed
		}
	}
	return nil
}
func oneRef(f oneFields, obj fssObject, id uint32) ([]fssID, error) {
	return f.objectReferences(obj, id)
}

func oneNoteWebCandidates(ctx context.Context, g *fssGraph) ([]fssID, error) {
	if g.spaces == nil {
		if _, err := g.reachableBlobs(); err != nil {
			return nil, err
		}
	}
	rootCell, ok := g.roots[fssDataRoot]
	if !ok {
		return nil, ErrMalformed
	}
	root, ok := g.spaces[rootCell]
	if !ok {
		return nil, ErrMalformed
	}
	w := oneNoteWalker{ctx: ctx, graph: g, seen: map[fssCell]map[fssID]bool{}}
	headerCell, ok := g.roots[fssHeaderRoot]
	if !ok {
		return nil, ErrMalformed
	}
	headerSpace, ok := g.spaces[headerCell]
	if !ok {
		return nil, ErrMalformed
	}
	if _, exists := headerSpace.objects[headerSpace.roots[1]]; exists {
		_, headerFields, err := w.object(headerSpace, headerSpace.roots[1], 0x00020001)
		if err != nil {
			return nil, err
		}
		for _, property := range []struct{ id, kind uint32 }{
			{0x14001d93, 5}, {0x1c001d94, 7}, {0x1c001d95, 7},
		} {
			if err := oneRequired(headerFields, property.id, property.kind); err != nil {
				return nil, err
			}
		}
	}
	if err := w.section(root); err != nil {
		return nil, err
	}
	return w.blobs, nil
}
func (w *oneNoteWalker) section(space fssSpace) error {
	meta, _, err := w.object(space, space.roots[2], 0x00020031)
	if err != nil {
		return err
	}
	_ = meta
	section, fields, err := w.object(space, space.roots[1], 0x00060007)
	if err != nil {
		return err
	}
	if err := oneRequired(fields, 0x1c001c30, 7); err != nil {
		return err
	} // entity GUID
	if err := oneRequired(fields, 0x18001c65, 6); err != nil {
		return err
	} // creation time
	series, err := oneRef(fields, section, 0x24001c20)
	if err != nil {
		return err
	}
	for _, id := range series {
		if err := w.pageSeries(space, id); err != nil {
			return err
		}
	}
	return nil
}
func (w *oneNoteWalker) pageSeries(space fssSpace, id fssID) error {
	obj, fields, err := w.object(space, id, 0x00060008)
	if err != nil {
		return err
	}
	if err := oneRequired(fields, 0x1c001c30, 7); err != nil {
		return err
	}
	cells, err := fields.spaceReferences(obj, space, 0x2c001d63)
	if err != nil {
		return err
	}
	for _, cell := range cells {
		pageSpace, ok := w.graph.spaces[cell]
		if !ok {
			return ErrMalformed
		}
		if err := w.page(pageSpace); err != nil {
			return err
		}
	}
	return nil
}
func (w *oneNoteWalker) page(space fssSpace) error {
	_, metadata, err := w.object(space, space.roots[2], 0x00020030)
	if err != nil {
		return err
	}
	if err := oneRequired(metadata, 0x1c001c30, 7); err != nil {
		return err
	}
	if err := oneRequiredUTF16(metadata, 0x1c001cf3); err != nil {
		return err
	}
	if err := oneRequired(metadata, 0x18001c65, 6); err != nil {
		return err
	}
	manifest, fields, err := w.object(space, space.roots[1], 0x00060037)
	if err != nil {
		return err
	}
	pageIDs, err := oneRef(fields, manifest, 0x24001c1f)
	if err != nil || len(pageIDs) == 0 {
		return ErrMalformed
	}
	for _, pageID := range pageIDs {
		page, pageFields, err := w.object(space, pageID, 0x0006000b)
		if err != nil {
			return err
		}
		for _, prop := range []uint32{0x24001d5f, 0x24001c20} { // title, then body
			contents, err := oneRef(pageFields, page, prop)
			if err != nil {
				return err
			}
			for _, id := range contents {
				if err := w.content(space, id, 0); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (w *oneNoteWalker) content(space fssSpace, id fssID, depth int) error {
	if depth > 64 {
		return ErrLimit
	}
	seen := w.seen[space.cell]
	if seen == nil {
		seen = map[fssID]bool{}
		w.seen[space.cell] = seen
	}
	if seen[id] {
		return nil
	}
	seen[id] = true
	obj, fields, err := w.object(space, id, 0)
	if err != nil {
		return err
	}
	switch obj.jcID {
	case 0x0006002c: // page title
		for _, prop := range []uint32{0x14001d7a, 0x14001c14, 0x14001c15} {
			if err := oneRequired(fields, prop, 5); err != nil {
				return err
			}
		}
		if err := oneRequired(fields, 0x24001c20, 9); err != nil {
			return err
		}
		children, err := oneRef(fields, obj, 0x24001c20)
		if err != nil {
			return err
		}
		for _, child := range children {
			if err := w.content(space, child, depth+1); err != nil {
				return err
			}
		}
	case 0x0006000c, 0x00060019: // outline or outline group
		if err := oneRequired(fields, 0x14001d7a, 5); err != nil {
			return err
		}
		if err := oneRequired(fields, 0x0c001c03, 3); err != nil {
			return err
		}
		items, err := oneRef(fields, obj, 0x24001c20)
		if err != nil {
			return err
		}
		for _, child := range items {
			if err := w.content(space, child, depth+1); err != nil {
				return err
			}
		}
	case 0x0006000d: // outline element
		if err := oneRequired(fields, 0x14001d09, 5); err != nil {
			return err
		}
		if err := oneRequired(fields, 0x14001d7a, 5); err != nil {
			return err
		}
		if err := oneRequired(fields, 0x0c001c03, 3); err != nil {
			return err
		}
		for _, prop := range []uint32{0x20001d78, 0x20001d79} { // original and latest author
			if err := oneRequired(fields, prop, 8); err != nil {
				return err
			}
			authors, err := oneRef(fields, obj, prop)
			if err != nil || len(authors) != 1 {
				return ErrMalformed
			}
		}
		for _, prop := range []uint32{0x24001c20, 0x24001c1f, 0x24001c26} {
			children, err := oneRef(fields, obj, prop)
			if err != nil {
				return err
			}
			for _, child := range children {
				if err := w.content(space, child, depth+1); err != nil {
					return err
				}
			}
		}
	case 0x00060035: // embedded-file node
		if err := oneRequired(fields, 0x14001d7a, 5); err != nil {
			return err
		}
		if err := oneRequired(fields, 0x1c001d9c, 7); err != nil {
			return err
		}
		pictures, err := oneRef(fields, obj, 0x20001c3f)
		if err != nil {
			return err
		}
		for _, picture := range pictures {
			pic, _, err := w.object(space, picture, 0)
			if err != nil || (pic.jcID != 0x00080039 && pic.jcID != 0x0008003a) {
				return ErrMalformed
			}
		}
		containers, err := oneRef(fields, obj, 0x20001d9b)
		if err != nil || len(containers) != 1 {
			return ErrMalformed
		}
		container, _, err := w.object(space, containers[0], 0x00080036)
		if err != nil {
			return err
		}
		if container.blob == (fssID{}) {
			return ErrMalformed
		}
		if _, ok := w.graph.blobs[container.blob]; !ok {
			return ErrMalformed
		}
		w.blobs = append(w.blobs, container.blob)
	case 0x00060011: // image node
		if err := oneRequired(fields, 0x14001d7a, 5); err != nil {
			return err
		}
		pictures, err := oneRef(fields, obj, 0x20001c3f)
		if err != nil {
			return err
		}
		for _, picture := range pictures {
			pic, _, err := w.object(space, picture, 0)
			if err != nil || (pic.jcID != 0x00080039 && pic.jcID != 0x0008003a) {
				return ErrMalformed
			}
		}
	case 0x0006000e: // rich text
		if err := oneRequired(fields, 0x14001d7a, 5); err != nil {
			return err
		}
		if err := oneOptional(fields, 0x1c001e12, 7); err != nil { // text-run indices
			return err
		}
		styles, err := oneRef(fields, obj, 0x2000342c)
		if err != nil || len(styles) != 1 {
			return ErrMalformed
		}
		if _, _, err := w.object(space, styles[0], 0); err != nil {
			return err
		}
		for _, prop := range []uint32{0x24001e13, 0x24003458} { // run formats and data objects
			references, err := oneRef(fields, obj, prop)
			if err != nil {
				return err
			}
			for _, reference := range references {
				if _, _, err := w.object(space, reference, 0); err != nil {
					return err
				}
			}
		}
	case 0x00080036, 0x00080039, 0x0008003a, 0x0012004d, 0x00120001: // referenced leaf objects
	default:
		return ErrUnsupported
	}
	return nil
}
