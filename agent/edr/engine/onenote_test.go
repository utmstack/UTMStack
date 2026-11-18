package engine

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func oneNoteObject(payload []byte) []byte {
	record := append([]byte(nil), oneNoteDataHeader[:]...)
	var size [8]byte
	binary.LittleEndian.PutUint64(size[:], uint64(len(payload)))
	record = append(record, size[:]...)
	record = append(record, make([]byte, 12)...)
	record = append(record, payload...)
	for len(record)%8 != 0 {
		record = append(record, 0)
	}
	return append(record, oneNoteDataFooter[:]...)
}

func oneNoteScan(t *testing.T, data, marker []byte) (string, error) {
	t.Helper()
	digest := md5.Sum(marker)
	s := pipelineState(t, map[string]string{
		"child.hdb": fmt.Sprintf("%x:%d:OneNote.Child\n", digest, len(marker)),
	})
	typ := signatures.TypeOneNote
	return s.visitObject(data, 0, objectOptions{forcedType: &typ})
}

func TestOneNoteFileDataStoreObjectScansRecursiveChild(t *testing.T) {
	marker := []byte("UTMSTACK-ONENOTE-EMBEDDED-CHILD-CONTROL")
	first := oneNoteObject([]byte("ordinary attachment"))
	second := oneNoteObject(marker)
	container := append(bytes.Repeat([]byte{0x36}, 40), first...)
	container = append(container, bytes.Repeat([]byte{0x55}, 13)...)
	container = append(container, second...)
	if sig, err := oneNoteScan(t, container, marker); sig != "OneNote.Child" || err != nil {
		t.Fatalf("second OneNote attachment: signature=%q error=%v", sig, err)
	}
	packed := oneNoteObject(gzipFixture(t, marker))
	if sig, err := oneNoteScan(t, packed, marker); sig != "OneNote.Child" || err != nil {
		t.Fatalf("compressed attachment's recursive child: signature=%q error=%v", sig, err)
	}
}

func TestOneNoteFileDataStoreObjectRejectsUnprovedBounds(t *testing.T) {
	marker := []byte("UTMSTACK-ONENOTE-EMBEDDED-CHILD-CONTROL")
	valid := oneNoteObject(marker)
	for _, tc := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"header", func(b []byte) []byte { b[0] ^= 1; return b }},
		{"reserved", func(b []byte) []byte { b[24] = 1; return b }},
		{"length-overflow", func(b []byte) []byte { binary.LittleEndian.PutUint64(b[16:24], ^uint64(0)); return b }},
		{"footer", func(b []byte) []byte { b[len(b)-1] ^= 1; return b }},
		{"truncated", func(b []byte) []byte { return b[:len(b)-7] }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := tc.mutate(bytes.Clone(valid))
			if sig, err := oneNoteScan(t, b, marker); sig != "" || err == nil {
				t.Fatalf("unproved OneNote record produced a child or clean verdict: signature=%q error=%v", sig, err)
			}
		})
	}
}

func oneNoteWebPackage(payload []byte, shortBlobHeader bool) []byte {
	file := make([]byte, 72)
	copy(file, []byte("\xe4\x52\x5c\x7b\x8c\xd8\xa7\x4d\xae\xb1\x53\x78\xd0\x29\x96\xd3"))
	copy(file[16:32], []byte("test-file-guid-1"))
	copy(file[32:48], file[16:32])
	binary.LittleEndian.PutUint32(file[68:72], 2|0x7a<<3|33<<17)
	file = append(file, 0x0c) // storage index ExGuid
	file = append(file, []byte("store-guid-12345")...)
	file = append(file, []byte("test-schema-guid")...)
	file = append(file, 0xac, 0x02, 0) // Data Element Package start and reserved
	record := []byte{0x0c, 0x56}       // compound Data Element; 43 fixed-field bytes
	record = append(record, 0x0c)      // ExGuid with a five-bit value
	record = append(record, []byte("test-blob-guid-1")...)
	record = append(record, 0x80) // serial number with GUID and uint64
	record = append(record, []byte("test-serial-guid")...)
	record = append(record, make([]byte, 8)...)
	record = append(record, 0x15) // CompactU64(10), Object Data BLOB
	if len(payload) >= 1<<14 {
		panic("test payload exceeds two-byte CompactU64")
	}
	var size []byte
	if len(payload) < 1<<7 {
		size = []byte{byte(len(payload)<<1 | 1)}
	} else {
		size = make([]byte, 2)
		binary.LittleEndian.PutUint16(size, uint16(len(payload)<<2|2))
	}
	declared := len(size) + len(payload)
	if shortBlobHeader {
		if declared >= 1<<7 {
			panic("test payload exceeds 16-bit stream-object header")
		}
		var h [2]byte
		binary.LittleEndian.PutUint16(h[:], uint16(declared<<9|2<<3))
		record = append(record, h[:]...)
	} else {
		var h [4]byte
		binary.LittleEndian.PutUint32(h[:], uint32(declared<<17|2<<3|2))
		record = append(record, h[:]...)
	}
	record = append(record, size...)
	record = append(record, payload...)
	record = append(record, 0x05) // Data Element end
	file = append(file, record...)
	return append(file, 0x55, 0xeb, 0x01) // package and packaging ends
}

func TestOneNoteFSSHTTPBObjectBlobRequiresGraph(t *testing.T) {
	marker := []byte("UTMSTACK-ONENOTE-WEB-BLOB-CONTROL")
	for _, short := range []bool{false, true} {
		data := oneNoteWebPackage(marker, short)
		if sig, err := oneNoteScan(t, data, marker); sig != "" || err == nil {
			t.Fatalf("short=%v: signature=%q error=%v", short, sig, err)
		}
	}
	packed := oneNoteWebPackage(gzipFixture(t, marker), false)
	if sig, err := oneNoteScan(t, packed, marker); sig != "" || err == nil {
		t.Fatalf("compressed web blob child: signature=%q error=%v", sig, err)
	}
	wrappedDesktopRecord := oneNoteWebPackage(oneNoteObject(marker), false)
	if sig, err := oneNoteScan(t, wrappedDesktopRecord, marker); sig != "" || err == nil {
		t.Fatalf("raw desktop record inside web blob bypassed graph guard: signature=%q error=%v", sig, err)
	}
	malformedWeb := bytes.Clone(wrappedDesktopRecord)
	malformedWeb[68] ^= 1
	if sig, err := oneNoteScan(t, malformedWeb, marker); sig != "" || err == nil {
		t.Fatalf("malformed web package bypassed desktop-record guard: signature=%q error=%v", sig, err)
	}
	terminatorBytes := append(bytes.Clone(marker), 0x55, 0xeb, 0x01)
	if sig, err := oneNoteScan(t, oneNoteWebPackage(terminatorBytes, false), terminatorBytes); sig != "" || err == nil {
		t.Fatalf("package-end bytes inside payload: signature=%q error=%v", sig, err)
	}
}

func TestOneNoteFSSHTTPBRejectsUnprovedBlob(t *testing.T) {
	marker := []byte("UTMSTACK-ONENOTE-WEB-BLOB-CONTROL")
	valid := oneNoteWebPackage(marker, false)
	const blobStart = 153
	for _, tc := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"package-guid", func(b []byte) []byte { b[32] ^= 1; return b }},
		{"non-compound-element", func(b []byte) []byte { b[108] ^= 4; return b }},
		{"parent-length", func(b []byte) []byte { b[109] ^= 2; return b }},
		{"exguid", func(b []byte) []byte { b[110] = 3; return b }},
		{"element-type", func(b []byte) []byte { b[152] = 0x13; return b }},
		{"blob-type", func(b []byte) []byte { b[blobStart] ^= 8; return b }},
		{"blob-length", func(b []byte) []byte { b[blobStart+2] ^= 2; return b }},
		{"compact-length", func(b []byte) []byte { b[blobStart+4] ^= 2; return b }},
		{"end-marker", func(b []byte) []byte { b[len(b)-4] ^= 4; return b }},
		{"package-end", func(b []byte) []byte { b[len(b)-3] ^= 4; return b }},
		{"malformed-tail-after-valid-blob", func(b []byte) []byte {
			return append(b[:len(b)-3], 0xff, 0x55, 0xeb, 0x01)
		}},
		{"truncated", func(b []byte) []byte { return b[:len(b)-5] }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if sig, err := oneNoteScan(t, tc.mutate(bytes.Clone(valid)), marker); sig != "" || err == nil {
				t.Fatalf("unproved web blob produced child or clean verdict: %q, %v", sig, err)
			}
		})
	}
}

func TestOneNoteOfficialWebGraph(t *testing.T) {
	dir := os.Getenv("EDR_CLAMAV_ONENOTE_DIR")
	if dir == "" {
		t.Skip("set EDR_CLAMAV_ONENOTE_DIR to the decoded official fixtures")
	}
	data, err := os.ReadFile(filepath.Join(dir, "webapp.one"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "f5ae4bdb59719b8b9fde153ebcaf12a5b67c214de7c4083c0e968d3c3db0e02f" {
		t.Fatalf("unrecognized OneNote fixture SHA-256: %s", got)
	}
	pos, ok := oneNoteFSSHTTPB(data)
	if !ok {
		t.Fatal("official fixture lost its web envelope")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := fssPackageTree(cancelled, data, pos); err != context.Canceled {
		t.Fatalf("cancelled graph walk: %v", err)
	}
	tree, err := fssPackageTree(context.Background(), data, pos)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := parseFSSGraph(tree)
	if err != nil {
		t.Fatal(err)
	}
	ids, err := graph.reachableBlobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.children) != 31 || len(graph.groups) != 11 || len(ids) != 3 {
		t.Fatalf("official graph changed: %d elements, %d groups, %d reachable blobs", len(tree.children), len(graph.groups), len(ids))
	}
	if len(graph.spaces) != 4 {
		t.Fatalf("official object-space count: %d", len(graph.spaces))
	}
	var objects int
	for _, space := range graph.spaces {
		objects += len(space.objects)
	}
	if objects != 51 {
		t.Fatalf("official object-space objects: %d", objects)
	}
	rootSpace, ok := graph.spaces[graph.roots[fssDataRoot]]
	if !ok {
		t.Fatal("data-root object space missing")
	}
	section, ok := rootSpace.objects[rootSpace.roots[1]]
	if !ok || section.jcID != 0x00060007 {
		t.Fatalf("section root object: %08x %v", section.jcID, ok)
	}
	fields, err := oneParseObjectFields(section.props)
	if err != nil {
		t.Fatal(err)
	}
	seriesIDs, err := fields.objectReferences(section, 0x24001c20)
	if err != nil || len(seriesIDs) != 1 {
		t.Fatalf("section page-series references: %d %v", len(seriesIDs), err)
	}
	series, ok := rootSpace.objects[seriesIDs[0]]
	if !ok || series.jcID != 0x00060008 {
		t.Fatalf("page-series object: %08x %v", series.jcID, ok)
	}
	seriesFields, err := oneParseObjectFields(series.props)
	if err != nil {
		t.Fatal(err)
	}
	pageCells, err := seriesFields.spaceReferences(series, rootSpace, 0x2c001d63)
	if err != nil || len(pageCells) != 1 {
		t.Fatalf("page-space references: %d %v", len(pageCells), err)
	}
	pageSpace, ok := graph.spaces[pageCells[0]]
	if !ok || len(pageSpace.objects) == 0 {
		t.Fatalf("page space missing: %v", ok)
	}
	manifest, ok := pageSpace.objects[pageSpace.roots[1]]
	if !ok || manifest.jcID != 0x00060037 {
		t.Fatalf("page manifest: %08x %v", manifest.jcID, ok)
	}
	manifestFields, err := oneParseObjectFields(manifest.props)
	if err != nil {
		t.Fatal(err)
	}
	pageIDs, err := manifestFields.objectReferences(manifest, 0x24001c1f)
	if err != nil || len(pageIDs) == 0 {
		t.Fatalf("page IDs: %d %v", len(pageIDs), err)
	}
	page, ok := pageSpace.objects[pageIDs[0]]
	if !ok || page.jcID != 0x0006000b {
		t.Fatalf("page object: %08x %v", page.jcID, ok)
	}
	pageFields, err := oneParseObjectFields(page.props)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := pageFields.objectReferences(page, 0x24001c20)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range contents {
		t.Logf("page content JCID %08x", pageSpace.objects[id].jcID)
		outline := pageSpace.objects[id]
		if outline.jcID != 0x0006000c {
			continue
		}
		outlineFields, e := oneParseObjectFields(outline.props)
		if e != nil {
			t.Fatal(e)
		}
		items, e := outlineFields.objectReferences(outline, 0x24001c20)
		if e != nil {
			t.Fatal(e)
		}
		for _, itemID := range items {
			item := pageSpace.objects[itemID]
			t.Logf("outline item JCID %08x", item.jcID)
			if item.jcID != 0x0006000d {
				continue
			}
			itemFields, e := oneParseObjectFields(item.props)
			if e != nil {
				t.Fatal(e)
			}
			contentIDs, e := itemFields.objectReferences(item, 0x24001c1f)
			if e != nil {
				t.Fatal(e)
			}
			for _, contentID := range contentIDs {
				t.Logf("outline content JCID %08x", pageSpace.objects[contentID].jcID)
				content := pageSpace.objects[contentID]
				if content.jcID != 0x00060035 {
					continue
				}
				contentFields, e := oneParseObjectFields(content.props)
				if e != nil {
					t.Fatal(e)
				}
				containerIDs, e := contentFields.objectReferences(content, 0x20001d9b)
				if e != nil || len(containerIDs) != 1 {
					t.Fatalf("embedded-file container refs: %d %v", len(containerIDs), e)
				}
				container, ok := pageSpace.objects[containerIDs[0]]
				if !ok || container.jcID != 0x00080036 || container.blob == (fssID{}) {
					t.Fatalf("embedded-file container: %08x %v", container.jcID, ok)
				}
				t.Logf("embedded file BLOB size %d", len(graph.blobs[container.blob]))
			}
		}
	}
	var executables, images int
	for _, id := range ids {
		data := graph.blobs[id]
		switch {
		case bytes.HasPrefix(data, []byte("MZ")):
			executables++
		case bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")):
			images++
		}
	}
	if executables != 1 || images != 2 {
		t.Fatalf("unexpected reachable blob types: %d PE, %d PNG", executables, images)
	}
	attachments, err := oneNoteWebCandidates(context.Background(), graph)
	if err != nil || len(attachments) != 1 || !bytes.HasPrefix(graph.blobs[attachments[0]], []byte("MZ")) {
		t.Fatalf("embedded-file traversal: %d candidates, %v", len(attachments), err)
	}
	for _, id := range ids {
		delete(graph.blobs, id)
		if _, err := graph.reachableBlobs(); err == nil {
			t.Fatal("missing referenced BLOB was accepted")
		}
		break
	}
}

func TestFSSCompactBounds(t *testing.T) {
	for _, tc := range []struct {
		data []byte
		want uint64
		ok   bool
	}{
		{[]byte{0}, 0, true},
		{[]byte{0xff}, 127, true},
		{[]byte{0x82, 0x08}, 544, true},
		{[]byte{0xd4, 0x8b, 0x10}, 135546, true},
		{[]byte{0x18, 0, 0, 0}, 1, true},
		{[]byte{0x30, 0, 0, 0, 0}, 1, true},
		{[]byte{0x60, 0, 0, 0, 0, 0}, 1, true},
		{[]byte{0xc0, 0, 0, 0, 0, 0, 0}, 1, true},
		{[]byte{0x80, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11}, 0x1122334455667788, true},
		{[]byte{0x82}, 0, false},
		{[]byte{0x80, 1}, 0, false},
	} {
		got, end, ok := fssCompact(tc.data, 0)
		if ok != tc.ok || ok && (got != tc.want || end != len(tc.data)) {
			t.Fatalf("%x: got %d end %d ok %v; want %d ok %v", tc.data, got, end, ok, tc.want, tc.ok)
		}
	}
}

// The official files are XOR-wrapped in Cisco Talos's source tree. Decode
// them outside this repository with its published key and set the directory
// to run this optional, hash-pinned native reference differential.
func TestOneNoteOfficialFixtureDifferential(t *testing.T) {
	dir := os.Getenv("EDR_CLAMAV_ONENOTE_DIR")
	if dir == "" {
		t.Skip("set EDR_CLAMAV_ONENOTE_DIR to the decoded ClamAV 1.5.4 OneNote fixtures")
	}
	const childMD5 = "aa15bcf478d165efd2065190eb473bcb"
	for _, tc := range []struct {
		name, digest string
	}{
		{"2007.one", "d0a9df85d9c3042263dd9c32e2b8110b7a5c3f295e222a8fa48029f8bf0a5583"},
		{"2010.one", "1139a72ab95778e4dcc473603bdcb126cf1d7d796eb9d6fc59d64a3d5468a2c4"},
		{"webapp.one", "f5ae4bdb59719b8b9fde153ebcaf12a5b67c214de7c4083c0e968d3c3db0e02f"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != tc.digest {
				t.Fatalf("unrecognized OneNote fixture SHA-256: %s", got)
			}
			s := pipelineState(t, map[string]string{
				"child.hdb": childMD5 + ":544:OneNote.Child\n",
			})
			name, scanErr := s.visit(data, 0)
			if name != "OneNote.Child" || scanErr != nil {
				t.Fatalf("Go missed official OneNote child: %q, %v", name, scanErr)
			}
			ref := os.Getenv("EDR_CLAMAV_PORTABLE")
			if ref == "" {
				return
			}
			clamscan := filepath.Join(ref, "bin", "clamscan")
			binary, err := os.ReadFile(clamscan)
			if err != nil || fmt.Sprintf("%x", sha256.Sum256(binary)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
				t.Fatalf("reference clamscan is not pinned 1.5.4: %v", err)
			}
			refDB := t.TempDir()
			if err := os.WriteFile(filepath.Join(refDB, "child.hdb"), []byte(childMD5+":544:OneNote.Child\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(clamscan, "--database="+refDB, "--no-summary", path)
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), path+": OneNote.Child.UNOFFICIAL FOUND") {
				t.Fatalf("ClamAV did not reach official OneNote child: %v %s", err, out)
			}
			cmd = exec.Command(clamscan, "--database="+refDB, "--scan-onenote=no", "--no-summary", path)
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err = cmd.CombinedOutput()
			if err != nil || !strings.Contains(string(out), path+": OK") {
				t.Fatalf("reference OneNote-disabled control was not clean: %v %s", err, out)
			}
		})
	}
}

// A detached but structurally valid Data Element is still visited by the
// reference. The Go framing parser must not infer graph reachability from
// those bytes alone; doing so produced false positives on invalid packages.
func TestOneNoteOfficialWebAppDetachedBlob(t *testing.T) {
	dir := os.Getenv("EDR_CLAMAV_ONENOTE_DIR")
	if dir == "" {
		t.Skip("set EDR_CLAMAV_ONENOTE_DIR to run the pinned web-app mutation")
	}
	data, err := os.ReadFile(filepath.Join(dir, "webapp.one"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "f5ae4bdb59719b8b9fde153ebcaf12a5b67c214de7c4083c0e968d3c3db0e02f" {
		t.Fatalf("unrecognized web-app fixture: %s", got)
	}
	const childMD5 = "aa15bcf478d165efd2065190eb473bcb"
	child := bytes.Index(data, []byte("MZ"))
	insert := bytes.LastIndex(data, []byte{0x55, 0xeb, 0x01})
	if child != 33342 || insert != 44056 || data[child+544] != 0x05 {
		t.Fatal("web-app object's pinned structural offsets changed")
	}
	mutated := bytes.Clone(data)
	mutated[child] ^= 1
	appended := append([]byte(nil), mutated[:insert]...)
	appended = append(appended, data[child-51:child+545]...)
	appended = append(appended, mutated[insert:]...)
	afterEnd := append(bytes.Clone(mutated), data[child-51:child+545]...)
	nestedFake := bytes.Clone(mutated)
	copy(nestedFake[20000:], data[child-51:child+545]) // inside the first PNG blob
	objectGroupFake := bytes.Clone(mutated)
	copy(objectGroupFake[35000:], data[child-51:child+545]) // outside a BLOB, inside object-group data
	rawDesktopRecordFake := bytes.Clone(mutated)
	copy(rawDesktopRecordFake[20000:], oneNoteObject(data[child:child+544])) // inside the first PNG blob
	unknownHeader := bytes.Clone(data)
	unknownHeader[153] = 0xfa // unsupported stream-object type before the PE blob
	var graphMutations []struct {
		name string
		data []byte
	}
	for _, at := range []int{113, 26140, 34903, 44055} {
		broken := bytes.Clone(data)
		broken[at] ^= 1
		graphMutations = append(graphMutations, struct {
			name string
			data []byte
		}{fmt.Sprintf("graph-flip-%d.one", at), broken})
	}
	scan := func(b []byte) (string, error) {
		s := pipelineState(t, map[string]string{"child.hdb": childMD5 + ":544:OneNote.Child\n"})
		return s.visit(b, 0)
	}
	if sig, err := scan(mutated); sig != "" || err == nil {
		t.Fatalf("mutated child was not incomplete: %q, %v", sig, err)
	}
	if sig, err := scan(appended); sig != "OneNote.Child" || err != nil {
		t.Fatalf("appended web object did not match reference detection: %q, %v", sig, err)
	}
	if sig, err := scan(afterEnd); sig != "" || err == nil {
		t.Fatalf("post-package object was scanned or silently clean: %q, %v", sig, err)
	}
	if sig, err := scan(nestedFake); sig != "" || err == nil {
		t.Fatalf("nested fake Data Element was scanned or silently clean: %q, %v", sig, err)
	}
	if sig, err := scan(objectGroupFake); sig != "" || err == nil {
		t.Fatalf("object-group fake Data Element was scanned or silently clean: %q, %v", sig, err)
	}
	if sig, err := scan(rawDesktopRecordFake); sig != "" || err == nil {
		t.Fatalf("web export scanned fake desktop record: %q, %v", sig, err)
	}
	if sig, err := scan(unknownHeader); sig != "" || err == nil {
		t.Fatalf("unknown stream-object type was scanned or silently clean: %q, %v", sig, err)
	}
	for _, tc := range graphMutations {
		if sig, err := scan(tc.data); sig != "" || err == nil {
			t.Fatalf("%s: unvalidated OneStore graph produced a verdict: %q, %v", tc.name, sig, err)
		}
	}
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	clamscan := filepath.Join(ref, "bin", "clamscan")
	binary, err := os.ReadFile(clamscan)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(binary)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("reference clamscan is not pinned 1.5.4: %v", err)
	}
	dbDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dbDir, "child.hdb"), []byte(childMD5+":544:OneNote.Child\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	referenceCases := []struct {
		name string
		data []byte
		want string
	}{
		{"mutated.one", mutated, ": OK"},
		{"appended.one", appended, ": OneNote.Child.UNOFFICIAL FOUND"},
		{"after-end.one", afterEnd, ": OK"},
		{"nested-fake.one", nestedFake, ": OK"},
		{"object-group-fake.one", objectGroupFake, ": OK"},
		{"desktop-record-fake.one", rawDesktopRecordFake, ": OK"},
		{"unknown-header.one", unknownHeader, ": OK"},
	}
	for _, tc := range graphMutations {
		referenceCases = append(referenceCases, struct {
			name string
			data []byte
			want string
		}{tc.name, tc.data, ": OK"})
	}
	for _, tc := range referenceCases {
		path := filepath.Join(t.TempDir(), tc.name)
		if err := os.WriteFile(path, tc.data, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(clamscan, "--database="+dbDir, "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if !strings.Contains(string(out), path+tc.want) || (tc.want == ": OK") != (err == nil) {
			t.Fatalf("reference %s: %v %s", tc.name, err, out)
		}
	}
}
