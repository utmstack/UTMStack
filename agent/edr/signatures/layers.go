package signatures

// Layer is one object of a scan's recursion stack: the scanned file at index
// 0, then each extracted or normalized object down to the current one.
type Layer struct {
	Type FileType
	Size int64
	// Normalized marks a representation of the layer below it (normalized
	// text or HTML): it is not a container of its own for Container and
	// Intermediates checks.
	Normalized bool
}

// Ancestry is the layer stack the Container and Intermediates descriptors
// read. The zero value is an empty stack.
type Ancestry struct {
	layers []Layer
}

func (a *Ancestry) Push(l Layer) { a.layers = append(a.layers, l) }

func (a *Ancestry) Pop() {
	if len(a.layers) > 0 {
		a.layers = a.layers[:len(a.layers)-1]
	}
}

// Retype changes the current layer's type, as a raw scan retypes text it
// recognizes as HTML or mail.
func (a *Ancestry) Retype(t FileType) {
	if len(a.layers) > 0 {
		a.layers[len(a.layers)-1].Type = t
	}
}

// Depth is the number of layers.
func (a *Ancestry) Depth() int {
	if a == nil {
		return 0
	}
	return len(a.layers)
}

// PhysicalDepth omits normalized views, which share their containing file's
// hook-trigger lifetime in ClamAV.
func (a *Ancestry) PhysicalDepth() int {
	if a == nil {
		return 0
	}
	n := 0
	for _, layer := range a.layers {
		if !layer.Normalized {
			n++
		}
	}
	return n
}

// resolve maps a layer index to a stack position the reference way: negative
// indexes count from the top (-1 is the current layer, -2 its parent), and
// normalized layers above layer 0 are counted with the layer below them. A
// result above the top means "too high"; below 0 means "none".
func (a *Ancestry) resolve(idx int) int {
	top := len(a.layers) - 1
	desired := idx
	if idx < 0 {
		desired = top + idx + 1
	}
	if desired > top {
		return top + 1
	}
	for cur := top; cur >= desired && cur > 0; cur-- {
		if a.layers[cur].Normalized {
			desired--
		}
	}
	return desired
}

// TypeAt is the type of the layer at idx: ANY when there is no such layer
// below, IGNORED when idx is above the top.
func (a *Ancestry) TypeAt(idx int) FileType {
	if a == nil || len(a.layers) == 0 {
		return TypeAny
	}
	switch r := a.resolve(idx); {
	case r < 0:
		return TypeAny
	case r >= len(a.layers):
		return TypeIgnored
	default:
		return a.layers[r].Type
	}
}

// SizeAt is the size of the layer at idx: the scanned file's size when there
// is no such layer below, 0 when idx is above the top.
func (a *Ancestry) SizeAt(idx int) int64 {
	if a == nil || len(a.layers) == 0 {
		return 0
	}
	switch r := a.resolve(idx); {
	case r < 0:
		return a.layers[0].Size
	case r >= len(a.layers):
		return 0
	default:
		return a.layers[r].Size
	}
}
