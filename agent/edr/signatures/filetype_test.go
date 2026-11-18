package signatures

import "testing"

func TestParseFileTypeUsesReferenceValues(t *testing.T) {
	cases := map[string]FileType{
		"CL_TYPE_ANY":          0,
		"CL_TYPE_TEXT_ASCII":   500,
		"CL_TYPE_BINARY_DATA":  504,
		"CL_TYPE_MSEXE":        506,
		"CL_TYPE_MSOLE2":       522,
		"CL_TYPE_HTML_UTF16":   539,
		"CL_TYPE_7Z":           541,
		"CL_TYPE_AI_MODEL":     557,
		"CL_TYPE_PART_HFSPLUS": 559,
		"CL_TYPE_HTML":         561,
		"CL_TYPE_MAIL":         562,
		"CL_TYPE_SFX":          563,
		"CL_TYPE_ZIPSFX":       564,
		"CL_TYPE_ALZ":          585,
		"CL_TYPE_IGNORED":      587,
	}
	for name, want := range cases {
		got, ok := ParseFileType(name)
		if !ok || got != want {
			t.Errorf("ParseFileType(%q) = %d, %v; want %d", name, got, ok, want)
		}
	}
	// The reference's name table has no entry for these two values, and
	// names are compared exactly.
	for _, name := range []string{"CL_TYPE_ERROR", "CL_TYPE_OTHER", "cl_type_msexe", "CL_TYPE_MSEXE ", "MSEXE", ""} {
		if _, ok := ParseFileType(name); ok {
			t.Errorf("ParseFileType(%q) accepted a name the reference does not know", name)
		}
	}
}

func TestFileTypeNameTableIsComplete(t *testing.T) {
	// Every value from ANY and TEXT_ASCII to IGNORED has a name except
	// ERROR and OTHER: 1 + 88 - 2.
	count := 0
	for v := FileType(500); v <= TypeIgnored; v++ {
		if s := v.String(); s != "" && v != TypeError && v != TypeOther {
			if got, ok := ParseFileType("CL_TYPE_" + s); !ok || got != v {
				t.Errorf("value %d named %q does not parse back", v, s)
			}
			count++
		}
	}
	if _, ok := ParseFileType("CL_TYPE_ANY"); ok {
		count++
	}
	if count != 87 {
		t.Fatalf("named types = %d, want 87", count)
	}
	if TypeError.String() != "ERROR" || TypeOther.String() != "OTHER" || TypeAny.String() != "ANY" {
		t.Fatalf("report names: %q %q %q", TypeError, TypeOther, TypeAny)
	}
}

func TestTargetOfFollowsReferenceTargetTable(t *testing.T) {
	cases := []struct {
		typ    FileType
		target int
	}{
		{TypeMSEXE, 1}, {TypeMSOLE2, 2}, {TypeHTML, 3}, {TypeMail, 4}, {TypeGraphics, 5}, {TypeGIF, 5},
		{TypePNG, 5}, {TypeJPEG, 5}, {TypeTIFF, 5}, {TypeELF, 6}, {TypeTextASCII, 7}, {TypeError, 8},
		{TypeMachO, 9}, {TypeMachOUnibin, 9}, {TypePDF, 10}, {TypeSWF, 11}, {TypeJava, 12},
		{TypeInternal, 13}, {TypeOther, 14},
		{TypeAny, 0}, {TypeTextUTF8, 0}, {TypeTextUTF16LE, 0}, {TypeTextUTF16BE, 0}, {TypeBinaryData, 0},
		{TypeScript, 0}, {TypeZIP, 0}, {TypeAIModel, 0}, {TypeHTMLUTF16, 0}, {TypeMHTML, 0}, {TypeHWPOLE2, 0}, {TypeIgnored, 0},
	}
	for _, c := range cases {
		if got := targetOf(c.typ); got != c.target {
			t.Errorf("targetOf(%s) = %d, want %d", c.typ, got, c.target)
		}
	}
}
