package signatures

func nativeACLiteral(p *bodyPattern) *nativeACPattern {
	plain := *p
	plain.nocase = false
	plain.nativeFixed = true
	shape, ok := plain.fixedACShape()
	if !ok {
		return nil
	}
	part := nativeACPart{literal: p.prefix, keyAt: int(shape.prefix), depth: int(shape.depth), minWidth: len(p.prefix), maxWidth: len(p.prefix)}
	part.nocase = p.nocase
	if p.nocase {
		part.keyAt = 0
		part.depth = min(3, len(p.prefix))
	}
	copy(part.key[:], p.prefix[part.keyAt:part.keyAt+part.depth])
	part.prefixMin, part.prefixMax = part.keyAt, part.keyAt
	part.identity = part.fingerprint()
	return &nativeACPattern{variants: [][]nativeACPart{{part}}}
}

func (d *Database) buildNativeAC() {
	d.nativeAC = [maxTargets]*nativeACRoot{}
	var regs [maxTargets][]nativeACRegistration
	var incomplete [maxTargets]bool
	models := map[int]*nativeACPattern{}
	for ri := range d.rules {
		r := &d.rules[ri]
		if !acRegistration(d, r) || d.bytecodeDisabled && (r.bytecode != nil || r.unknown == UnknownBytecode) {
			continue
		}
		for sub, id := range r.patterns {
			p := &d.patterns[id]
			if p.holder != nil {
				continue
			}
			model := models[id]
			if model == nil {
				if p.graph != nil {
					model = p.graph.native
				}
				if model != nil {
					models[id] = model
				}
			}
			if model == nil && !(p.simple && !p.nativeAC && len(p.prefix) >= 2) {
				incomplete[r.target] = true
				continue
			}
			regs[r.target] = append(regs[r.target], nativeACRegistration{id, ri, sub, model, p.offset, r.expr == nil && r.bytecode == nil && r.unknown != UnknownBytecode, 0})
		}
	}
	for root, registrations := range regs {
		if !incomplete[root] && len(registrations) > 0 {
			d.nativeAC[root] = newNativeACRoot(registrations, d.patterns)
		}
	}
}

func (m *patternCounts) nativeRoot(root int) (*nativeACResult, error) {
	if m.native[root] != nil {
		return m.native[root], nil
	}
	reached := m.reached
	if reached == nil || root != 0 && root != m.info.target {
		var err error
		reached, err = m.d.reachedRules(m.candidates, root, m.w)
		if err != nil {
			return nil, err
		}
	}
	var regs []nativeACRegistration
	logical := map[int]bool{}
	registry := m.d.nativeAC[root]
	models := map[int]*nativeACPattern{}
	caps := map[int]int{}
	for ri := range reached.members() {
		r := &m.d.rules[ri]
		if r.target != root || !acRegistration(m.d, r) || m.d.bytecodeDisabled && (r.bytecode != nil || r.unknown == UnknownBytecode) {
			continue
		}
		standalone := r.expr == nil && r.bytecode == nil && r.unknown != UnknownBytecode
		for sub, id := range r.patterns {
			if !m.candidates.has(id) {
				continue
			}
			cap := r.countCap(sub)
			if r.bytecode != nil || r.byteCompareRefs&(uint64(1)<<uint(sub)) != 0 {
				cap = int(^uint(0) >> 1)
			}
			caps[id] = max(caps[id], cap)
			if !standalone && logical[id] {
				continue
			}
			model := models[id]
			if model == nil {
				model = registry.model(id)
				if model != nil {
					models[id] = model
				}
			}
			if model == nil {
				continue
			}
			if !standalone {
				logical[id] = true
			}
			regs = append(regs, nativeACRegistration{id, ri, sub, model, m.d.patterns[id].offset, standalone, 0})
		}
	}
	for i := range regs {
		regs[i].cap = caps[regs[i].pattern]
	}
	result, err := registry.scan(m.data, regs, m.info, m.w)
	if err != nil {
		return nil, err
	}
	m.native[root] = result
	return result, nil
}

func (r *nativeACRoot) modeled(id int) bool { return r.models[id] != nil || r.literals.has(id) }
func (r *nativeACRoot) model(id int) *nativeACPattern {
	if m := r.models[id]; m != nil {
		return m
	}
	if r.literals.has(id) {
		return nativeACLiteral(&r.patterns[id])
	}
	return nil
}

func (d *Database) matchNativeAC(root int, m *patternCounts, info *objectInfo, w *matchWork) (string, error, error) {
	result, err := m.nativeRoot(root)
	if err != nil {
		return "", nil, err
	}
	var pending error
	for _, ri := range result.order {
		name, cause, err := d.evaluateRule(&d.rules[ri], m, info, w)
		if err != nil {
			return "", nil, err
		}
		if name != "" {
			return name, nil, nil
		}
		if cause != nil && pending == nil {
			pending = cause
		}
	}
	return "", pending, nil
}
