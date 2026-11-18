package engine

import "encoding/binary"

// An object space is the cell's active revision chain. Later revisions are
// visited first, so an object already resolved there is not replaced by a
// base revision. Root roles follow the declarations in that chain.
type fssObject struct {
	id        fssID
	jcID      uint32
	props     []byte
	objectIDs []fssID
	cellIDs   []fssCell
	blob      fssID
}
type fssSpace struct {
	cell    fssCell
	roots   map[uint32]fssID
	objects map[fssID]fssObject
}

func (g *fssGraph) objectSpaces() (map[fssCell]fssSpace, error) {
	spaces := make(map[fssCell]fssSpace, len(g.storageCells))
	for cell, manifest := range g.storageCells {
		if manifest == (fssID{}) {
			continue
		}
		revisionID, ok := g.cells[manifest]
		if !ok {
			return nil, ErrMalformed
		}
		mapped, ok := g.revisionMappings[revisionID]
		if !ok {
			return nil, ErrMalformed
		}
		space := fssSpace{cell: cell, roots: map[uint32]fssID{}, objects: map[fssID]fssObject{}}
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
			for role, id := range revision.roots {
				space.roots[role] = id
			}
			for _, groupID := range revision.groups {
				group, ok := g.groups[groupID]
				if !ok {
					return nil, ErrMalformed
				}
				byID := map[fssID]map[uint64]fssObjectPart{}
				for _, part := range group.parts {
					parts := byID[part.id]
					if parts == nil {
						parts = map[uint64]fssObjectPart{}
						byID[part.id] = parts
					}
					parts[part.partition] = part
				}
				for id, parts := range byID {
					if _, exists := space.objects[id]; exists {
						continue
					}
					metadata, mok := parts[4]
					data, dok := parts[1]
					if !mok || !dok || metadata.kind != 0x16 || data.kind != 0x16 || len(metadata.data) < 4 {
						return nil, ErrMalformed
					}
					obj := fssObject{id: id, jcID: binary.LittleEndian.Uint32(metadata.data), props: data.data, objectIDs: data.objectIDs, cellIDs: data.cellIDs}
					if file, ok := parts[2]; ok {
						if file.kind != 0x1c {
							return nil, ErrMalformed
						}
						obj.blob = file.blob
					}
					space.objects[id] = obj
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
		spaces[cell] = space
	}
	return spaces, nil
}
