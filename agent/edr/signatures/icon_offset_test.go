package signatures

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type iconOffsetFixture struct {
	name   string
	data   []byte
	marker string
}

func iconOffsetFixtures() []iconOffsetFixture {
	le := binary.LittleEndian
	var fixtures []iconOffsetFixture
	makeIcon := func(w, h, depth, alpha int) labIcon {
		ic := labIcon{w: w, h: h, bpp: depth, mask: make([]bool, w*h)}
		if depth <= 8 {
			ic.palette = make([][3]byte, 1<<depth)
			ic.index = make([]int, w*h)
			for k := range ic.palette {
				ic.palette[k] = [3]byte{byte(k*37 + 17), byte(k*71 + 29), byte(k*19 + 43)}
			}
		} else if depth == 16 {
			ic.raw16 = make([]uint16, w*h)
		} else {
			ic.rgba = make([][4]byte, w*h)
		}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				i := y*w + x
				ic.mask[i] = (x+2*y)%3 == 0
				if depth <= 8 {
					ic.index[i] = (x + 3*y) % (1 << depth)
				} else if depth == 16 {
					ic.raw16[i] = uint16(x*341 + y*177)
				} else {
					a := byte(x*19 + y*31)
					if alpha >= 0 {
						a = byte(alpha)
					}
					ic.rgba[i] = [4]byte{byte(x*13 + y*7), byte(x*3 + y*19), byte(x*23 + y*11), a}
				}
			}
		}
		return ic
	}
	add := func(w, h, depth, alpha int, place string) {
		ic := makeIcon(w, h, depth, alpha)
		pe := labProgram([]labIcon{ic})
		const imageAt = 0x200 + 160
		bitmap := ic.bitmap()
		if depth <= 8 && alpha >= 0 {
			for i := 0; i < 1<<depth; i++ {
				bitmap[40+4*i+3] = byte(alpha)
			}
			copy(pe[imageAt:], bitmap)
		}
		payload := bitmap[40:]
		if place == "maskless" || place == "wrap-maskless" {
			payload = payload[:(w*depth+31)/32*4*h]
		}
		if place != "ordinary" {
			headerAt := len(pe)
			switch place {
			case "maskless":
				pe = append(pe, bitmap[:40]...)
				pe = append(pe, payload...)
			default:
				target := len(pe)
				pe = append(pe, payload...)
				if place != "wrap-maskless" {
					pe = append(pe, make([]byte, 128)...)
				}
				headerAt = len(pe)
				pe = append(pe, bitmap[:40]...)
				if place == "zero" {
					target = 0
				}
				if place == "past-end" {
					target = len(pe) + 16
				}
				if place == "short-tail" {
					target = len(pe) - 8
				}
				le.PutUint32(pe[headerAt:], uint32(target)-uint32(headerAt))
			}
			le.PutUint32(pe[0x280:], uint32(0x1000+headerAt-0x200))
		}
		v := len(pe) - 0x200
		le.PutUint32(pe[0x140:], uint32(v))
		le.PutUint32(pe[0x148:], uint32(v))
		le.PutUint32(pe[0x58+116:], uint32(v))
		le.PutUint32(pe[0x58+56:], uint32(0x1000+(v+4095)/4096*4096))
		marker := fmt.Sprintf("EDR-ICON-OFFSET-%08x", len(fixtures))
		copy(pe[0x20:0x38], marker)
		fixtures = append(fixtures, iconOffsetFixture{fmt.Sprintf("%dx%d-d%d-a%d-%s", w, h, depth, alpha, place), pe, marker})
	}
	for w := 16; w <= 256; w++ {
		for _, depth := range []int{1, 4, 8, 16, 24, 32} {
			for _, place := range []string{"ordinary", "wrap"} {
				add(w, w, depth, -1, place)
			}
		}
	}
	for _, dimensions := range [][2]int{{16, 18}, {18, 16}, {24, 30}, {30, 24}, {32, 40}, {40, 32}} {
		for _, depth := range []int{1, 4, 8, 16, 24, 32} {
			for _, place := range []string{"ordinary", "wrap"} {
				add(dimensions[0], dimensions[1], depth, -1, place)
			}
		}
	}
	for _, w := range []int{16, 24, 32} {
		for _, depth := range []int{1, 4, 8} {
			for _, alpha := range []int{1, 127, 255} {
				for _, place := range []string{"ordinary", "wrap"} {
					add(w, w, depth, alpha, place)
				}
			}
		}
	}
	for _, w := range []int{16, 24, 32} {
		for _, place := range []string{"maskless", "wrap-maskless"} {
			add(w, w, 32, 0, place)
		}
	}
	for _, depth := range []int{1, 4, 8, 16, 24, 32} {
		for _, place := range []string{"zero", "past-end", "short-tail"} {
			add(16, 16, depth, -1, place)
		}
	}
	return fixtures
}

func TestIconOffsetFixtureWorker(t *testing.T) {
	dir := os.Getenv("OPENEDR_ICON_OFFSET_LAB")
	if dir == "" {
		t.Skip("owned inert fixture export worker")
	}
	if err := os.MkdirAll(filepath.Join(dir, "objects"), 0700); err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	for _, fixture := range iconOffsetFixtures() {
		if err := os.WriteFile(filepath.Join(dir, "objects", fixture.name+".exe"), fixture.data, 0600); err != nil {
			t.Fatal(err)
		}
		meta := parsePE(fixture.data)
		if meta == nil || !meta.valid {
			t.Fatal("fixture header rejected", fixture.name)
		}
		icons, known := meta.programIcons(fixture.data)
		hashes := []string{}
		for _, icon := range icons {
			hashes = append(hashes, iconHashWire(icon.hash))
		}
		rows = append(rows, map[string]any{"fixture": fixture.name, "marker": fixture.marker, "known": known, "hashes": hashes})
	}
	body, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "hashes.json"), body, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestIconOffsetDecisionWorker(t *testing.T) {
	dir := os.Getenv("OPENEDR_ICON_OFFSET_LAB")
	if dir == "" {
		t.Skip("owned icon decision worker")
	}
	for _, profile := range []string{"positive", "different"} {
		files := map[string][]byte{}
		for _, name := range []string{"probe.idb", "probe.ldb"} {
			body, err := os.ReadFile(filepath.Join(dir, "db-"+profile, name))
			if err != nil {
				t.Fatal(err)
			}
			files[name] = body
		}
		d, err := Compile(files)
		if err != nil {
			t.Fatal(err)
		}
		var rows []map[string]any
		for _, fixture := range iconOffsetFixtures() {
			got, err := d.MatchObject(nil2ctx(), fixture.data, MatchContext{Type: TypeMSEXE})
			row := map[string]any{"fixture": fixture.name, "name": got.Name}
			if err != nil {
				row["error"] = err.Error()
			}
			rows = append(rows, row)
		}
		body, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(dir, "decisions-"+profile+".json"), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestIconOffsetReference(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "icon_offsets.json"))
	if err != nil {
		t.Fatal(err)
	}
	var golden struct {
		Hashes []struct {
			Fixture string
			Hashes  []string
			Input   struct {
				Bytes  int
				SHA256 string
			}
		}
		Decisions []struct{ Fixture, Profile, Name string }
		Logical   string
		Profiles  map[string]string
	}
	if err = json.Unmarshal(body, &golden); err != nil {
		t.Fatal(err)
	}
	if len(golden.Hashes) != 3042 || len(golden.Decisions) != 6084 {
		t.Fatal("incomplete offset oracle")
	}
	fixtures := map[string][]byte{}
	for _, fixture := range iconOffsetFixtures() {
		fixtures[fixture.name] = fixture.data
	}
	for _, row := range golden.Hashes {
		t.Run("hash/"+row.Fixture, func(t *testing.T) {
			data, found := fixtures[row.Fixture]
			if !found || len(data) != row.Input.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.Input.SHA256 {
				t.Fatal("input differs from independently scanned bytes")
			}
			meta := parsePE(data)
			icons, known := meta.programIcons(data)
			if !known || len(icons) != len(row.Hashes) {
				t.Fatalf("physical icon read: known=%t count=%d want=%d", known, len(icons), len(row.Hashes))
			}
			for i, icon := range icons {
				if got := iconHashWire(icon.hash); got != row.Hashes[i] {
					t.Fatalf("reference hash: got=%s want=%s", got, row.Hashes[i])
				}
			}
		})
	}
	dbs := map[string]*Database{}
	for name, idb := range golden.Profiles {
		d, err := Compile(map[string][]byte{"probe.idb": []byte(idb), "probe.ldb": []byte(golden.Logical)})
		if err != nil {
			t.Fatal(err)
		}
		dbs[name] = d
	}
	for _, row := range golden.Decisions {
		t.Run(row.Profile+"/"+row.Fixture, func(t *testing.T) {
			d := dbs[row.Profile]
			data := fixtures[row.Fixture]
			if d == nil || data == nil {
				t.Fatal("unknown decision oracle input")
			}
			result, err := d.MatchObject(nil2ctx(), data, MatchContext{Type: TypeMSEXE})
			if err != nil || result.Name != row.Name {
				t.Fatalf("icon condition: got=%q err=%v want=%q", result.Name, err, row.Name)
			}
		})
	}
}

func FuzzIconPhysicalWindow(f *testing.F) {
	bitmap := labProgram([]labIcon{{w: 16, bpp: 24, rgba: make([][4]byte, 256)}})
	f.Add(bitmap, uint32(0x2a0), uint32(40))
	f.Add(bitmap, uint32(0x2a0), ^uint32(0))
	f.Fuzz(func(t *testing.T, data []byte, at, size uint32) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		data = append([]byte(nil), data...)
		if len(data) > 0 {
			at %= uint32(len(data))
		}
		if uint64(at)+4 <= uint64(len(data)) {
			binary.LittleEndian.PutUint32(data[at:], size)
		}
		icon, skip, sure := parseIconImageFile(data, int(at))
		if !sure {
			t.Fatal("complete mapped object retained a suffix-only uncertainty")
		}
		if !skip {
			if w := iconComparisonWidth(icon.hash.width); w != 16 && w != 24 && w != 32 {
				t.Fatal("unmodeled icon bucket", w)
			}
			for _, areas := range [][3]iconArea{icon.hash.color, icon.hash.gray, icon.hash.bright, icon.hash.dark, icon.hash.edge, icon.hash.noedge} {
				for _, area := range areas {
					if area.v < 0 || area.v > iconMaxValue12 || area.x < 0 || area.x >= icon.hash.width || area.y < 0 || area.y >= icon.hash.width {
						t.Fatal("decoded measure escaped icon bounds")
					}
				}
			}
		}
	})
}
