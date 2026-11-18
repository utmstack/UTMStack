package signatures

import "testing"

func stack(layers ...Layer) *Ancestry {
	a := &Ancestry{}
	for _, l := range layers {
		a.Push(l)
	}
	return a
}

func TestAncestryResolvesLikeTheReference(t *testing.T) {
	cases := []struct {
		name  string
		a     *Ancestry
		idx   int
		want  FileType
		depth int
	}{
		{"root has no parent", stack(Layer{Type: TypeTextASCII}), -2, TypeAny, 1},
		{"root's own type", stack(Layer{Type: TypeTextASCII}), -1, TypeTextASCII, 1},
		{"member's parent", stack(Layer{Type: TypeZIP}, Layer{Type: TypeTextASCII}), -2, TypeZIP, 2},
		{"normalized layer on the root counts with it", stack(Layer{Type: TypeTextASCII}, Layer{Type: TypeTextASCII, Normalized: true}), -2, TypeAny, 2},
		{"normalized member layer skips to the archive", stack(Layer{Type: TypeZIP}, Layer{Type: TypeTextASCII}, Layer{Type: TypeTextASCII, Normalized: true}), -2, TypeZIP, 3},
		{"normalized top resolves to the layer below", stack(Layer{Type: TypeZIP}, Layer{Type: TypeHTML}, Layer{Type: TypeHTML, Normalized: true}), -1, TypeHTML, 3},
		{"too high", stack(Layer{Type: TypeZIP}), 5, TypeIgnored, 1},
		{"empty stack", nil, -2, TypeAny, 0},
	}
	for _, c := range cases {
		if got := c.a.TypeAt(c.idx); got != c.want || c.a.Depth() != c.depth {
			t.Errorf("%s: TypeAt(%d) = %s depth %d, want %s depth %d", c.name, c.idx, got, c.a.Depth(), c.want, c.depth)
		}
	}
	a := stack(Layer{Type: TypeZIP, Size: 100}, Layer{Type: TypeTextASCII, Size: 10})
	if a.SizeAt(-2) != 100 || a.SizeAt(-1) != 10 || a.SizeAt(7) != 0 {
		t.Fatalf("sizes %d %d %d", a.SizeAt(-2), a.SizeAt(-1), a.SizeAt(7))
	}
	a.Retype(TypeHTML)
	a.Pop()
	a.Pop()
	a.Pop()
	if a.Depth() != 0 {
		t.Fatalf("depth after pops %d", a.Depth())
	}
}

func TestContainerRulesMatchOnlyInTheirContainer(t *testing.T) {
	d := compileTest(t, "ldb", "InZip;Engine:81-255,Container:CL_TYPE_ZIP,Target:0;0;4d41524b4552\nTop;Engine:81-255,Container:CL_TYPE_ANY,Target:0;0;544f50\n")
	member := "a MARKER in a member"
	expectOutcome(t, d, member, MatchContext{Type: TypeAny, Layers: stack(Layer{Type: TypeZIP}, Layer{Type: TypeTextASCII})}, "InZip", "")
	expectOutcome(t, d, member, MatchContext{Type: TypeAny, Layers: stack(Layer{Type: TypeTextASCII})}, "", "")
	expectOutcome(t, d, member, MatchContext{Type: TypeAny, Layers: stack(Layer{Type: TypeOOXMLWord}, Layer{Type: TypeTextASCII})}, "", "")
	expectOutcome(t, d, member, MatchContext{Type: TypeAny}, "", "")
	top := "TOP of a file"
	expectOutcome(t, d, top, MatchContext{Type: TypeAny, Layers: stack(Layer{Type: TypeTextASCII})}, "Top", "")
	expectOutcome(t, d, top, MatchContext{Type: TypeAny}, "Top", "")
	expectOutcome(t, d, top, MatchContext{Type: TypeAny, Layers: stack(Layer{Type: TypeZIP}, Layer{Type: TypeTextASCII})}, "", "")
	// A normalized layer is matched as part of the object it represents.
	expectOutcome(t, d, member, MatchContext{Type: TypeTextASCII, Layers: stack(Layer{Type: TypeZIP}, Layer{Type: TypeTextASCII}, Layer{Type: TypeTextASCII, Normalized: true})}, "InZip", "")
}

func TestIntermediatesRulesFollowTheWholeChain(t *testing.T) {
	d := compileTest(t, "ldb", "Chain;Engine:81-255,Intermediates:CL_TYPE_ZIP>CL_TYPE_7Z,Target:0;0;4d41524b4552\nSkipAny;Engine:81-255,Intermediates:CL_TYPE_ZIP>CL_TYPE_ANY,Target:0;0;5a5a5a\n")
	data := "a MARKER"
	expectOutcome(t, d, data, MatchContext{Type: TypeAny, Layers: stack(Layer{Type: TypeZIP}, Layer{Type: Type7Z}, Layer{Type: TypeBinaryData})}, "Chain", "")
	expectOutcome(t, d, data, MatchContext{Type: TypeAny, Layers: stack(Layer{Type: Type7Z}, Layer{Type: TypeZIP}, Layer{Type: TypeBinaryData})}, "", "")
	expectOutcome(t, d, data, MatchContext{Type: TypeAny, Layers: stack(Layer{Type: Type7Z}, Layer{Type: TypeBinaryData})}, "", "")
	// An ANY element uses no layer: the parent must be ZIP, and the chain
	// must be deep enough for both elements.
	expectOutcome(t, d, "a ZZZ", MatchContext{Type: TypeAny, Layers: stack(Layer{Type: TypeRAR}, Layer{Type: TypeZIP}, Layer{Type: TypeBinaryData})}, "SkipAny", "")
	expectOutcome(t, d, "a ZZZ", MatchContext{Type: TypeAny, Layers: stack(Layer{Type: TypeZIP}, Layer{Type: TypeBinaryData})}, "", "")
}

func TestContainerIsCheckedOnBytecodeTriggers(t *testing.T) {
	f := callbackWithLabel([]byte("UTMStack.Contained"), true)
	f.logical = "Callback.Trigger;Engine:81-255,Container:CL_TYPE_PDF,Target:0;0;4142"
	d := compileCallback(t, f)
	expectOutcome(t, d, "xxABxx", MatchContext{Type: TypeAny, Layers: stack(Layer{Type: TypePDF}, Layer{Type: TypeBinaryData})}, "UTMStack.Contained", "")
	expectOutcome(t, d, "xxABxx", MatchContext{Type: TypeAny, Layers: stack(Layer{Type: TypeBinaryData})}, "", "")
}
