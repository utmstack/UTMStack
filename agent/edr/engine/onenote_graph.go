package engine

import (
	"context"
	"encoding/binary"
	"fmt"
)

// The FSSHTTPB package is a tree of length-delimited stream objects. Build
// that tree before interpreting any ID: an apparent object inside binary data
// must never become a graph edge.
type fssNode struct {
	typ        uint32
	body       []byte
	start, end int
	children   []*fssNode
}

func fssPackageTree(ctx context.Context, data []byte, pos int) (*fssNode, error) {
	root := &fssNode{typ: 0x15}
	stack := []*fssNode{root}
	for count := 0; count < 1_000_000; count++ {
		if count&1023 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		t, ok := fssReadToken(data, pos)
		if !ok || !fssKnownObjectType(t.typ) {
			return nil, ErrMalformed
		}
		if t.end {
			if len(stack) == 0 || stack[len(stack)-1].typ != t.typ {
				return nil, ErrMalformed
			}
			stack = stack[:len(stack)-1]
			pos = t.next
			if len(stack) == 0 {
				end, ok := fssReadToken(data, pos)
				if !ok || !end.end || end.typ != 0x7a || end.next != pos+2 {
					return nil, ErrMalformed
				}
				return root, nil
			}
			continue
		}
		parent := stack[len(stack)-1]
		end, ok := fssLeafEnd(data, t, parent.typ)
		if !ok {
			return nil, ErrMalformed
		}
		node := &fssNode{typ: t.typ, body: data[t.next:end], start: t.next, end: end}
		parent.children = append(parent.children, node)
		pos = end
		if t.compound {
			if len(stack) >= 64 {
				return nil, ErrLimit
			}
			stack = append(stack, node)
		}
	}
	return nil, ErrLimit
}

// BinaryItem has its own CompactU64 length. The OneNote parser reads that
// length for BLOB and ObjectGroupDataObject bodies and does not rely on the
// redundant outer stream-object length. Keep the same bounded boundary here.
func fssLeafEnd(data []byte, t fssToken, parentType uint32) (int, bool) {
	if t.next > len(data) {
		return 0, false
	}
	if !t.compound && t.typ == 2 && parentType == 1 {
		size, start, ok := fssCompact(data, t.next)
		if !ok || size > uint64(len(data)-start) {
			return 0, false
		}
		return start + int(size), true
	}
	if !t.compound && t.typ == 0x16 && parentType == 0x1e {
		w := fssWire{data: data[t.next:]}
		if _, ok := w.idArray(); !ok {
			return 0, false
		}
		if _, ok := w.cellArray(); !ok {
			return 0, false
		}
		if _, ok := w.binaryItem(); !ok {
			return 0, false
		}
		return t.next + w.pos, true
	}
	if t.length > uint64(len(data)-t.next) {
		return 0, false
	}
	return t.next + int(t.length), true
}

type fssID struct {
	guid  [16]byte
	value uint32
}

// MS-ONESTORE 2.7.1 assigns these roots to the header and data-root cells.
var fssHeaderRoot = fssID{guid: [16]byte{0x9c, 0x31, 0x5a, 0x1a, 0x6b, 0xc2, 0xaa, 0x41, 0xb9, 0xc5, 0x9b, 0xd8, 0xc4, 0x4e, 0x07, 0xd4}, value: 1}
var fssDataRoot = fssID{guid: [16]byte{0xb9, 0xfa, 0xde, 0x84, 0xa3, 0xaa, 0x0d, 0x4a, 0xa3, 0xa8, 0x52, 0x0c, 0x77, 0xac, 0x70, 0x73}, value: 2}
var fssRevisionRoleGUID = [16]byte{0xf8, 0x17, 0x37, 0x4a, 0x14, 0x1c, 0xe7, 0x49, 0x95, 0x26, 0x81, 0xd9, 0x42, 0xde, 0x17, 0x41}

type fssCell struct{ first, second fssID }
type fssWire struct {
	data []byte
	pos  int
}

func (w *fssWire) bytes(n int) ([]byte, bool) {
	if n < 0 || n > len(w.data)-w.pos {
		return nil, false
	}
	b := w.data[w.pos : w.pos+n]
	w.pos += n
	return b, true
}
func (w *fssWire) compact() (uint64, bool) {
	v, end, ok := fssCompact(w.data, w.pos)
	if ok {
		w.pos = end
	}
	return v, ok
}
func (w *fssWire) id() (fssID, bool) {
	var id fssID
	if w.pos >= len(w.data) {
		return id, false
	}
	first := w.data[w.pos]
	w.pos++
	switch {
	case first == 0:
		return id, true
	case first&7 == 4:
		id.value = uint32(first >> 3)
	case first&63 == 32:
		b, ok := w.bytes(1)
		if !ok {
			return id, false
		}
		id.value = uint32(b[0])<<2 | uint32(first>>6)
	case first&127 == 64:
		b, ok := w.bytes(2)
		if !ok {
			return id, false
		}
		id.value = uint32(binary.LittleEndian.Uint16(b))<<1 | uint32(first>>7)
	case first == 128:
		b, ok := w.bytes(4)
		if !ok {
			return id, false
		}
		id.value = binary.LittleEndian.Uint32(b)
	default:
		return id, false
	}
	b, ok := w.bytes(16)
	if !ok {
		return id, false
	}
	copy(id.guid[:], b)
	return id, true
}
func (w *fssWire) cell() (fssCell, bool) {
	a, ok := w.id()
	if !ok {
		return fssCell{}, false
	}
	b, ok := w.id()
	return fssCell{a, b}, ok
}
func (w *fssWire) serial() bool {
	b, ok := w.bytes(1)
	if !ok {
		return false
	}
	if b[0] == 0 {
		return true
	}
	_, ok = w.bytes(24)
	return ok
}
func (w *fssWire) idArray() ([]fssID, bool) {
	n, ok := w.compact()
	if !ok || n > 1_000_000 || n > uint64(len(w.data)-w.pos) {
		return nil, false
	}
	out := make([]fssID, 0, int(n))
	for i := uint64(0); i < n; i++ {
		id, ok := w.id()
		if !ok {
			return nil, false
		}
		out = append(out, id)
	}
	return out, true
}
func (w *fssWire) cellArray() ([]fssCell, bool) {
	n, ok := w.compact()
	if !ok || n > 1_000_000 || n > uint64((len(w.data)-w.pos)/2) {
		return nil, false
	}
	out := make([]fssCell, 0, int(n))
	for i := uint64(0); i < n; i++ {
		id, ok := w.cell()
		if !ok {
			return nil, false
		}
		out = append(out, id)
	}
	return out, true
}
func (w *fssWire) binaryItem() (fssSpan, bool) {
	n, ok := w.compact()
	if !ok || n > uint64(len(w.data)-w.pos) {
		return fssSpan{}, false
	}
	start := w.pos
	w.pos += int(n)
	return fssSpan{start, w.pos}, true
}
func (w *fssWire) done() bool { return w.pos == len(w.data) }

type fssObjectPart struct {
	id                   fssID
	partition            uint64
	kind                 uint32
	data                 []byte
	dataStart            int
	blob                 fssID
	objectIDs            []fssID
	cellIDs              []fssCell
	objectRefs, cellRefs int
}
type fssGroup struct {
	blobs []fssID
	parts []fssObjectPart
}
type fssRevision struct {
	base   fssID
	groups []fssID
	roots  map[uint32]fssID
}
type fssGraph struct {
	storageCells     map[fssCell]fssID
	revisionMappings map[fssID]fssID
	roots            map[fssID]fssCell
	cells            map[fssID]fssID
	revisions        map[fssID]fssRevision
	groups           map[fssID]fssGroup
	blobs            map[fssID][]byte
	spaces           map[fssCell]fssSpace
}

// parseFSSGraph validates every known data-element shape and records the
// references needed to prove that a blob belongs to a live OneStore cell.
// A structurally valid graph is only a prerequisite: identifying embedded
// file nodes still requires MS-ONE property-set traversal.
func parseFSSGraph(root *fssNode) (*fssGraph, error) {
	g := &fssGraph{storageCells: map[fssCell]fssID{}, revisionMappings: map[fssID]fssID{}, roots: map[fssID]fssCell{}, cells: map[fssID]fssID{}, revisions: map[fssID]fssRevision{}, groups: map[fssID]fssGroup{}, blobs: map[fssID][]byte{}}
	for index, elem := range root.children {
		if elem.typ != 1 {
			return nil, ErrMalformed
		}
		w := fssWire{data: elem.body}
		id, ok := w.id()
		if !ok || !w.serial() {
			return nil, ErrMalformed
		}
		kind, ok := w.compact()
		if !ok || !w.done() {
			return nil, ErrMalformed
		}
		switch kind {
		case 1:
			if err := g.storageIndex(elem); err != nil {
				return nil, fmt.Errorf("storage index element %d: %w", index, err)
			}
		case 2:
			if err := g.storageManifest(elem); err != nil {
				return nil, fmt.Errorf("storage manifest element %d: %w", index, err)
			}
		case 3:
			if len(elem.children) != 1 || elem.children[0].typ != 0x0b {
				return nil, ErrMalformed
			}
			w := fssWire{data: elem.children[0].body}
			rev, ok := w.id()
			if !ok || !w.done() {
				return nil, ErrMalformed
			}
			g.cells[id] = rev
		case 4:
			if err := g.revisionManifest(id, elem); err != nil {
				return nil, fmt.Errorf("revision manifest element %d: %w", index, err)
			}
		case 5:
			group, err := parseFSSGroup(elem)
			if err != nil {
				return nil, fmt.Errorf("object group element %d: %w", index, err)
			}
			g.groups[id] = group
		case 6:
			if err := parseFSSFragment(elem); err != nil {
				return nil, err
			}
		case 10:
			if len(elem.children) != 1 || elem.children[0].typ != 2 {
				return nil, ErrMalformed
			}
			w := fssWire{data: elem.children[0].body}
			span, ok := w.binaryItem()
			if !ok || !w.done() {
				return nil, ErrMalformed
			}
			g.blobs[id] = elem.children[0].body[span.start:span.end]
		default:
			return nil, ErrMalformed
		}
	}
	return g, nil
}

func (g *fssGraph) storageIndex(elem *fssNode) error {
	for _, child := range elem.children {
		w := fssWire{data: child.body}
		switch child.typ {
		case 0x11:
			_, ok := w.id()
			if !ok || !w.serial() || !w.done() {
				return ErrMalformed
			}
		case 0x0e:
			cell, ok := w.cell()
			if !ok {
				return ErrMalformed
			}
			id, ok := w.id()
			if !ok || !w.serial() || !w.done() {
				return ErrMalformed
			}
			g.storageCells[cell] = id
		case 0x0d:
			id, ok := w.id()
			if !ok {
				return ErrMalformed
			}
			mapped, ok := w.id()
			if !ok || !w.serial() || !w.done() {
				return ErrMalformed
			}
			g.revisionMappings[id] = mapped
		default:
			return ErrMalformed
		}
	}
	return nil
}
func (g *fssGraph) storageManifest(elem *fssNode) error {
	if len(elem.children) < 1 || elem.children[0].typ != 0x0c || len(elem.children[0].body) != 16 {
		return ErrMalformed
	}
	for _, child := range elem.children[1:] {
		if child.typ != 7 {
			return ErrMalformed
		}
		w := fssWire{data: child.body}
		id, ok := w.id()
		if !ok {
			return ErrMalformed
		}
		cell, ok := w.cell()
		if !ok || !w.done() {
			return ErrMalformed
		}
		g.roots[id] = cell
	}
	return nil
}
func (g *fssGraph) revisionManifest(id fssID, elem *fssNode) error {
	if len(elem.children) < 1 || elem.children[0].typ != 0x1a {
		return ErrMalformed
	}
	w := fssWire{data: elem.children[0].body}
	_, ok := w.id()
	if !ok {
		return ErrMalformed
	}
	base, ok := w.id()
	if !ok || !w.done() {
		return ErrMalformed
	}
	var refs []fssID
	roots := map[uint32]fssID{}
	for _, child := range elem.children[1:] {
		w := fssWire{data: child.body}
		switch child.typ {
		case 0x0a:
			role, valid := w.id()
			if !valid || role.guid != fssRevisionRoleGUID || role.value < 1 || role.value > 4 {
				return ErrMalformed
			}
			var objectID fssID
			objectID, ok = w.id()
			if ok {
				roots[role.value] = objectID
			}
		case 0x19:
			var ref fssID
			ref, ok = w.id()
			refs = append(refs, ref)
		default:
			return ErrMalformed
		}
		if !ok || !w.done() {
			return ErrMalformed
		}
	}
	g.revisions[id] = fssRevision{base: base, groups: refs, roots: roots}
	return nil
}

func parseFSSGroup(elem *fssNode) (fssGroup, error) {
	var group fssGroup
	if len(elem.children) < 2 || len(elem.children) > 3 || elem.children[0].typ != 0x1d {
		return group, ErrMalformed
	}
	decl := elem.children[0]
	if len(decl.body) != 0 {
		return group, ErrMalformed
	}
	declParts := make([]fssObjectPart, 0, len(decl.children))
	for _, child := range decl.children {
		w := fssWire{data: child.body}
		objectID, ok := w.id()
		if !ok {
			return group, ErrMalformed
		}
		var blob fssID
		if child.typ == 5 {
			blob, ok = w.id()
			if !ok {
				return group, ErrMalformed
			}
		} else if child.typ != 0x18 {
			return group, ErrMalformed
		}
		partition, ok := w.compact()
		if !ok {
			return group, ErrMalformed
		}
		if child.typ != 5 {
			if _, ok = w.compact(); !ok {
				return group, ErrMalformed
			}
		} // inline data size
		objectRefs, ok := w.compact()
		if !ok {
			return group, ErrMalformed
		}
		cellRefs, ok := w.compact()
		if !ok {
			return group, ErrMalformed
		}
		if !w.done() {
			return group, ErrMalformed
		}
		_ = blob // declaration ID is not the data record's blob reference
		if objectRefs > 1_000_000 || cellRefs > 1_000_000 {
			return group, ErrLimit
		}
		declParts = append(declParts, fssObjectPart{id: objectID, partition: partition, kind: child.typ, objectRefs: int(objectRefs), cellRefs: int(cellRefs)})
	}
	dataIndex := 1
	if elem.children[1].typ == 0x79 {
		meta := elem.children[1]
		if len(meta.body) != 0 {
			return group, ErrMalformed
		}
		for _, child := range meta.children {
			if child.typ != 0x78 {
				return group, ErrMalformed
			}
			w := fssWire{data: child.body}
			n, ok := w.compact()
			if !ok || n > 4 || !w.done() {
				return group, ErrMalformed
			}
		}
		dataIndex++
	}
	if dataIndex != len(elem.children)-1 || elem.children[dataIndex].typ != 0x1e {
		return group, ErrMalformed
	}
	objects := elem.children[dataIndex]
	if len(objects.body) != 0 || len(objects.children) != len(declParts) {
		return group, ErrMalformed
	}
	for i, child := range objects.children {
		w := fssWire{data: child.body}
		if (declParts[i].kind == 5) != (child.typ == 0x1c) {
			return group, ErrMalformed
		}
		refs, ok := w.idArray()
		if !ok {
			return group, ErrMalformed
		}
		cells, ok := w.cellArray()
		if !ok {
			return group, ErrMalformed
		}
		switch child.typ {
		case 3:
			_, ok = w.compact()
		case 0x16:
			var span fssSpan
			span, ok = w.binaryItem()
			if ok {
				declParts[i].data = child.body[span.start:span.end]
				declParts[i].dataStart = child.start + span.start
			}
		case 0x1c:
			var blob fssID
			blob, ok = w.id()
			group.blobs = append(group.blobs, blob)
			declParts[i].blob = blob
		default:
			return group, ErrMalformed
		}
		if !ok || !w.done() {
			return group, ErrMalformed
		}
		// The data record's bounded arrays are authoritative. The parser
		// reads declaration counts but does not require them to equal the
		// actual object/cell arrays in a later data record.
		declParts[i].objectRefs = len(refs)
		declParts[i].cellRefs = len(cells)
		declParts[i].kind = child.typ
		declParts[i].objectIDs = refs
		declParts[i].cellIDs = cells
	}
	group.parts = declParts
	return group, nil
}

func parseFSSFragment(elem *fssNode) error {
	if len(elem.children) != 1 || elem.children[0].typ != 0x6a {
		return ErrMalformed
	}
	w := fssWire{data: elem.children[0].body}
	_, ok := w.id()
	if !ok {
		return ErrMalformed
	}
	size, ok := w.compact()
	if !ok {
		return ErrMalformed
	}
	_, ok = w.compact()
	if !ok {
		return ErrMalformed
	}
	_, ok = w.compact()
	if !ok || size != uint64(len(w.data)-w.pos) {
		return ErrMalformed
	}
	return nil
}

// reachableBlobs follows the storage manifest through cell/revision/group
// mappings. Detached BLOB elements never become scan candidates.
func (g *fssGraph) reachableBlobs() ([]fssID, error) {
	if len(g.roots) == 0 || len(g.storageCells) == 0 {
		return nil, ErrMalformed
	}
	if _, ok := g.roots[fssHeaderRoot]; !ok {
		return nil, ErrMalformed
	}
	if _, ok := g.roots[fssDataRoot]; !ok {
		return nil, ErrMalformed
	}
	seen := map[fssID]bool{}
	var ids []fssID
	for _, cell := range g.roots {
		if _, ok := g.storageCells[cell]; !ok {
			return nil, ErrMalformed
		}
	}
	for _, manifest := range g.storageCells {
		if manifest == (fssID{}) {
			continue
		}
		rev, ok := g.cells[manifest]
		if !ok {
			return nil, ErrMalformed
		}
		mapped, ok := g.revisionMappings[rev]
		if !ok {
			return nil, ErrMalformed
		}
		visited := map[fssID]bool{}
		for mapped != (fssID{}) {
			if visited[mapped] {
				return nil, ErrMalformed
			}
			visited[mapped] = true
			revision, ok := g.revisions[mapped]
			if !ok {
				return nil, ErrMalformed
			}
			for _, groupID := range revision.groups {
				group, ok := g.groups[groupID]
				if !ok {
					return nil, ErrMalformed
				}
				if err := validateFSSGroupObjects(group); err != nil {
					return nil, fmt.Errorf("object group %v: %w", groupID, err)
				}
				for _, blob := range group.blobs {
					if _, ok := g.blobs[blob]; !ok {
						return nil, ErrMalformed
					}
					if !seen[blob] {
						seen[blob] = true
						ids = append(ids, blob)
					}
				}
			}
			if revision.base == (fssID{}) {
				break
			}
			mapped, ok = g.revisionMappings[revision.base]
			if !ok {
				return nil, ErrMalformed
			}
		}
	}
	spaces, err := g.objectSpaces()
	if err != nil {
		return nil, err
	}
	g.spaces = spaces
	return ids, nil
}
