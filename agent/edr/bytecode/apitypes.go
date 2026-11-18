package bytecode

// referenceAPITypes are the compound types of the reference's API table
// (bytecode_api_decl.c cli_apicall_types, IDs from 69) that the supported
// calls use: the environment record, the PE section record and the decoded
// instruction record, each behind a pointer. Element IDs below 69 are base
// types. A program's prototype must equal them structurally, as the
// reference's loader requires (bytecode.c types_equal).
var referenceAPITypes = map[int]bcType{
	86: {kind: pointerType, elements: []int{87}},
	87: {kind: structType, elements: []int{32, 32, 32, 32, 32, 32, 32, 88, 88, 88, 88, 88, 88, 88, 8, 8, 8, 8, 8, 8, 8, 8, 8}},
	88: {kind: arrayType, elements: []int{8}, count: 65},
	91: {kind: pointerType, elements: []int{92}},
	92: {kind: structType, elements: []int{32, 32, 32, 32, 32, 32, 32, 32, 32}},
	95: {kind: pointerType, elements: []int{96}},
	96: {kind: structType, elements: []int{16, 8, 8, 8, 98, 97}},
	97: {kind: arrayType, elements: []int{8}, count: 29},
	98: {kind: arrayType, elements: []int{99}, count: 3},
	99: {kind: arrayType, elements: []int{8}, count: 10},
}

func memberCount(t bcType) int {
	switch t.kind {
	case arrayType:
		return t.count
	case pointerType:
		return 1
	}
	return len(t.elements)
}

// referenceEqual reports whether the program's type id equals the reference
// type ref: base types by ID, compound types by kind, member count and
// members (only the element type of an array). memo records assumed
// equalities so recursive types terminate, as in the reference.
func (t *typeTable) referenceEqual(id, ref int, memo map[int]int) bool {
	if ref < firstDynamicType {
		return id == ref
	}
	if memo[ref] == id+1 {
		return true
	}
	memo[ref] = id + 1
	want, ok := referenceAPITypes[ref]
	if !ok || id < firstCompoundType || id-firstCompoundType >= len(t.types) {
		return false
	}
	have := t.types[id-firstCompoundType]
	if have.kind != want.kind || memberCount(have) != memberCount(want) || len(have.elements) < len(want.elements) {
		return false
	}
	for i, e := range want.elements {
		if !t.referenceEqual(have.elements[i], e, memo) {
			return false
		}
		if have.kind == arrayType {
			break
		}
	}
	return true
}
