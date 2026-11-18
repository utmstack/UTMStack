package signatures

import "sort"

// Type recognition runs the file-type database's type patterns over a whole
// object during its raw scan, as the reference does, to find HTML or mail
// markers anywhere in text and objects embedded in other objects.

// scanWindow is the reference's matching window. Type results are decided
// per window and the largest window result is the object's.
const scanWindow = 131072

// Bounds on the recorded matches: per pattern, and embedded records (the
// reference keeps at most 10, except for zip self-extractors).
const (
	maxTypeMatches    = 4096
	maxEmbedded       = 10
	maxEmbeddedZIPSFX = 1024
)

type typeHit struct {
	start, order int
	typ          FileType
	fixed        bool // a single-part pattern
}

// recognize returns the detected type and the embedded records of the
// object scanned as scanType.
func (m *patternCounts) recognize(scanType FileType) (FileType, []EmbeddedType, error) {
	var hits []typeHit
	for order, fp := range m.d.ftypes.patterns {
		if fp.rtype != TypeAny && fp.rtype != scanType || !m.candidates.has(fp.id) {
			continue
		}
		p := &m.d.patterns[fp.id]
		if p.offset.relative() && m.info.status == exeFailed {
			continue
		}
		starts, err := m.starts(fp.id, p)
		if err != nil {
			return 0, nil, err
		}
		for _, s := range starts {
			hits = append(hits, typeHit{s, order, fp.typ, p.minWidth() == p.maxWidth()})
		}
	}
	sort.Slice(hits, func(a, b int) bool {
		if hits[a].start != hits[b].start {
			return hits[a].start < hits[b].start
		}
		return hits[a].order < hits[b].order
	})
	var embedded []EmbeddedType
	others, zipSFX := 0, 0
	record := func(h typeHit) {
		switch {
		case h.typ == TypeZIPSFX && zipSFX < maxEmbeddedZIPSFX:
			zipSFX++
		case h.typ != TypeZIPSFX && others < maxEmbedded:
			others++
		default:
			return
		}
		embedded = append(embedded, EmbeddedType{h.typ, h.start})
	}
	result, current, window, stopped := FileType(0), FileType(0), -1, false
	for _, h := range hits {
		if w := h.start / scanWindow; w != window {
			result, current, window, stopped = max(result, current), 0, w, false
		}
		if stopped {
			continue
		}
		var ok bool
		if current, ok = typeResult(current, h.typ); !ok {
			stopped = true // an ignored type ends this window's pass
			continue
		}
		switch {
		case h.typ >= TypeSFX && h.typ != TypeIgnored,
			h.typ == TypeMBR && h.fixed,
			h.typ == TypeMSEXE && (scanType == TypeMSEXE || scanType == TypeZIP || scanType == TypeMSOLE2):
			record(h)
		}
	}
	return max(result, current), embedded, nil
}

// starts lists the match starts of a candidate pattern, in order, at most
// maxTypeMatches of them.
func (m *patternCounts) starts(id int, p *bodyPattern) ([]int, error) {
	pc := m.candidates.startCount(p, id, m.data, m.info.meta)
	var starts []int
	for len(starts) < maxTypeMatches {
		before := pc.n
		if err := p.countMore(m.data, m.w, pc, before+1); err != nil {
			return nil, err
		}
		if pc.n == before {
			break
		}
		starts = append(starts, pc.last)
	}
	return starts, nil
}
