package registry

import "fmt"

func (d *Dataset) GetOrBuildIndex(matchColumn string) (map[string][]int, error) {
	d.indicesMu.RLock()
	idx, ok := d.Indices[matchColumn]
	d.indicesMu.RUnlock()
	if ok {
		return idx, nil
	}

	colIdx, exists := d.ColIndex[matchColumn]
	if !exists {
		return nil, fmt.Errorf("column %q not in dataset headers", matchColumn)
	}

	d.indicesMu.Lock()
	defer d.indicesMu.Unlock()
	if idx, ok = d.Indices[matchColumn]; ok {
		return idx, nil
	}

	newIdx := make(map[string][]int, len(d.RawRows))
	for i, row := range d.RawRows {
		if colIdx < len(row) {
			key := NormalizeKey(row[colIdx])
			newIdx[key] = append(newIdx[key], i)
		}
	}
	d.Indices[matchColumn] = newIdx
	return newIdx, nil
}
