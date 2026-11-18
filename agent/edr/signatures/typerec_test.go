package signatures

import (
	"context"
	"strings"
	"testing"
)

const recognitionFTM = `0:0:4d5a:DOS executable:CL_TYPE_ANY:CL_TYPE_MSEXE
0:0:504b0304:ZIP:CL_TYPE_ANY:CL_TYPE_ZIP
1:*:3c68746d6c3e:HTML:CL_TYPE_ANY:CL_TYPE_HTML
1:*:46726f6d3a20:mail:CL_TYPE_ANY:CL_TYPE_MAIL
1:*:5375626a6563743a20:mail subject in text:CL_TYPE_TEXT_ASCII:CL_TYPE_MAIL
1:*:504b0304:ZIP self-extractor:CL_TYPE_ANY:CL_TYPE_ZIPSFX
1:*:4d5a{60-300}50450000:embedded executable:CL_TYPE_ANY:CL_TYPE_MSEXE
1:EOF-8:6b6f6c79:disk image trailer:CL_TYPE_ANY:CL_TYPE_DMG
`

func recognize(t *testing.T, d *Database, data string, typ FileType) MatchResult {
	t.Helper()
	res, err := d.MatchObject(context.Background(), []byte(data), MatchContext{Type: typ, TypeRecognition: true})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestTypeRecognitionFindsMarkersAnywhere(t *testing.T) {
	d := compileFTM(t, recognitionFTM)
	text := strings.Repeat("words ", 40000) + "<html>" // the marker is past 200 KiB
	if got := recognize(t, d, text, TypeAny).DetectedType; got != TypeHTML {
		t.Fatalf("HTML marker far in text: %s", got)
	}
	if got := recognize(t, d, text+"\nFrom: someone\n", TypeAny).DetectedType; got != TypeMail {
		t.Fatalf("mail and HTML markers: %s, want MAIL (the higher type)", got)
	}
	// A required type must equal the scan type; a raw scan of TEXT_ASCII
	// passes ANY, as the reference does.
	if got := recognize(t, d, "Subject: hello\n"+strings.Repeat("x", 100), TypeAny).DetectedType; got == TypeMail {
		t.Fatalf("a text-only pattern applied to an ANY scan")
	}
	if got := recognize(t, d, "Subject: hello\n"+strings.Repeat("x", 100), TypeTextASCII).DetectedType; got != TypeMail {
		t.Fatalf("a text-only pattern did not apply to a text scan: %s", got)
	}
	res, err := d.MatchObject(context.Background(), []byte(text), MatchContext{Type: TypeAny})
	if err != nil || res.DetectedType != 0 || len(res.Embedded) != 0 {
		t.Fatalf("without recognition: %+v %v", res, err)
	}
}

func TestTypeRecognitionRecordsEmbeddedObjects(t *testing.T) {
	d := compileFTM(t, recognitionFTM)
	pe := "MZ" + strings.Repeat("\x00", 62) + "PE\x00\x00" + strings.Repeat("\x00", 100)
	exe := pe + "overlay PK\x03\x04 archive"
	res := recognize(t, d, exe, TypeMSEXE)
	// The executable's own header is recorded at offset 0, as the reference
	// records it; extraction only considers later offsets.
	want := []EmbeddedType{{TypeMSEXE, 0}, {TypeZIPSFX, strings.Index(exe, "PK")}}
	if len(res.Embedded) != 2 || res.Embedded[0] != want[0] || res.Embedded[1] != want[1] {
		t.Fatalf("embedded = %+v, want %+v", res.Embedded, want)
	}
	// An embedded executable is recorded only in executables, zips and OLE2
	// files.
	inner := "junk data ahead " + pe
	for _, c := range []struct {
		typ  FileType
		want int
	}{{TypeMSEXE, 1}, {TypeZIP, 1}, {TypeMSOLE2, 1}, {TypeBinaryData, 0}, {TypeAny, 0}} {
		n := 0
		for _, e := range recognize(t, d, inner, c.typ).Embedded {
			if e.Type == TypeMSEXE {
				n++
			}
		}
		if n != c.want {
			t.Errorf("%s: %d embedded executables, want %d", c.typ, n, c.want)
		}
	}
	// Relative offsets resolve during a raw scan.
	image := strings.Repeat("\x01", 600) + "koly" + "\x00\x00\x00\x00"
	res = recognize(t, d, image, TypeBinaryData)
	if len(res.Embedded) != 1 || res.Embedded[0].Type != TypeDMG || res.DetectedType != TypeDMG {
		t.Fatalf("trailer: %+v", res)
	}
}

func TestTypeRecognitionWindowsTakeTheLargestResult(t *testing.T) {
	d := compileFTM(t, recognitionFTM)
	// Window 1 ends on HTML (560); window 2 holds only the self-extractor
	// marker (563): across windows the larger wins.
	data := "<html>" + strings.Repeat("x", 140000) + "PK\x03\x04"
	if got := recognize(t, d, data, TypeAny).DetectedType; got != TypeZIPSFX {
		t.Fatalf("got %s, want ZIPSFX", got)
	}
	// In one window a later HTML match does not replace a self-extractor.
	data = "PK\x03\x04 then <html>"
	if got := recognize(t, d, data, TypeAny).DetectedType; got != TypeZIPSFX {
		t.Fatalf("got %s, want ZIPSFX", got)
	}
}
