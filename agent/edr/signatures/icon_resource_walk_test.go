package signatures

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const resourceFlatHash = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
const resourceGradientHash = "10c480a00bbd000bb830a0b3f301014ce04055e70609e60a07e0060bcf0a0b27000057000460040069070b4008073b06031400001700051b090022231f57"

type iconResourceFixture struct {
	name string
	data []byte
}

// These inert PEs deliberately put a directory at EOF while keeping its
// earlier bitmap/data entries readable. Only reference scanner verdicts, not
// this generator, define the golden expectations.
func iconResourceFixtures() []iconResourceFixture {
	le := binary.LittleEndian
	flat := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 256)}
	gradient := labIcon{w: 16, bpp: 24, rgba: make([][4]byte, 256)}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			gradient.rgba[y*16+x] = [4]byte{byte(x * 17), byte(y * 15), byte((x*13 + y*11) & 255), 255}
		}
	}
	var fixtures []iconResourceFixture
	for _, profile := range []struct {
		name string
		icon labIcon
	}{{"flat", flat}, {"gradient", gradient}} {
		for _, level := range []string{"names", "languages"} {
			for _, counts := range [][2]uint16{{0, 1}, {0, 2}, {0, 65535}, {1, 0}, {1, 1}, {65535, 1}, {65535, 2}, {65534, 3}, {32768, 32768}} {
				for _, tail := range []int{0, 4, 8, 12, 16} {
					pe := labProgram([]labIcon{profile.icon})
					// Original relative resource locations in the independent
					// generator: type-3 names=32, icon languages=80, data=128.
					at := len(pe)
					pe = append(pe, make([]byte, 16+tail)...)
					le.PutUint16(pe[at+12:], counts[0])
					le.PutUint16(pe[at+14:], counts[1])
					if tail >= 4 {
						le.PutUint32(pe[at+16:], 1)
					}
					if tail >= 8 {
						le.PutUint32(pe[at+20:], 128)
					}
					if tail >= 12 {
						le.PutUint32(pe[at+24:], 2)
					}
					if tail >= 16 {
						le.PutUint32(pe[at+28:], 0xffffffff)
					}
					switch level {
					case "names":
						le.PutUint32(pe[0x214:], 0x80000000|uint32(at-0x200))
						if tail >= 8 {
							le.PutUint32(pe[at+20:], 0x80000050)
						}
					case "languages":
						le.PutUint32(pe[0x234:], 0x80000000|uint32(at-0x200))
					}
					v := len(pe) - 0x200
					le.PutUint32(pe[0x140:], uint32(v))
					le.PutUint32(pe[0x148:], uint32((v+511)/512*512))
					le.PutUint32(pe[0x58+116:], uint32(v))
					le.PutUint32(pe[0x58+56:], uint32(0x1000+(v+4095)/4096*4096))
					fixtures = append(fixtures, iconResourceFixture{fmt.Sprintf("%s-%s-%d-%d-tail%d", profile.name, level, counts[0], counts[1], tail), pe})
				}
			}
		}
		for _, count := range []uint16{0, 1, 2, 3, 65535} {
			pe := labProgram([]labIcon{profile.icon})
			le.PutUint16(pe[0x20e:], count)
			fixtures = append(fixtures, iconResourceFixture{fmt.Sprintf("%s-root-count%d", profile.name, count), pe})
		}
		// A physically complete language directory with a wrapping count
		// distinguishes a false detection from a conservative old unknown.
		pe := labProgram([]labIcon{profile.icon})
		at := len(pe)
		pe = append(pe, make([]byte, 16+8*65536)...)
		le.PutUint16(pe[at+12:], 65535)
		le.PutUint16(pe[at+14:], 1)
		le.PutUint32(pe[at+16:], 0x409)
		le.PutUint32(pe[at+20:], 128)
		le.PutUint32(pe[0x234:], 0x80000000|uint32(at-0x200))
		v := len(pe) - 0x200
		le.PutUint32(pe[0x140:], uint32(v))
		le.PutUint32(pe[0x148:], uint32((v+511)/512*512))
		le.PutUint32(pe[0x58+116:], uint32(v))
		le.PutUint32(pe[0x58+56:], uint32(0x1000+(v+4095)/4096*4096))
		fixtures = append(fixtures, iconResourceFixture{profile.name + "-language-count-wrap-physical", pe})
	}
	return fixtures
}

func TestIconResourceFixtureWorker(t *testing.T) {
	dir := os.Getenv("OPENEDR_ICON_RESOURCE_LAB")
	if dir == "" {
		t.Skip("owned inert fixture export worker")
	}
	if err := os.MkdirAll(filepath.Join(dir, "objects"), 0700); err != nil {
		t.Fatal(err)
	}
	const flatHash = "1000000000000000000000000000000000000000000000000000000000000000000000040000080000000000000000000000000000040000080000000000"
	const gradientHash = "10c480a00bbd000bb830a0b3f301014ce04055e70609e60a07e0060bcf0a0b27000057000460040069070b4008073b06031400001700051b090022231f57"
	for _, db := range []struct{ name, hash string }{{"flat", flatHash}, {"gradient", gradientHash}} {
		dbdir := filepath.Join(dir, "db-"+db.name)
		if err := os.MkdirAll(dbdir, 0700); err != nil {
			t.Fatal(err)
		}
		for name, body := range map[string]string{"fixture.idb": "Safe.Resource:A:B:" + db.hash + "\n", "fixture.ldb": "Icon.Resource;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a\n"} {
			if err := os.WriteFile(filepath.Join(dbdir, name), []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
		}
		d := iconDatabase(t, "Safe.Resource:A:B:"+db.hash+"\n", "Icon.Resource;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a")
		var rows []map[string]any
		for _, fixture := range iconResourceFixtures() {
			if err := os.WriteFile(filepath.Join(dir, "objects", fixture.name+".exe"), fixture.data, 0600); err != nil {
				t.Fatal(err)
			}
			r, err := d.MatchObject(nil2ctx(), fixture.data, MatchContext{Type: TypeMSEXE})
			row := map[string]any{"fixture": fixture.name, "name": r.Name}
			if err != nil {
				row["error"] = err.Error()
			}
			rows = append(rows, row)
		}
		body, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(dir, "go-"+db.name+".json"), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestIconResourceWalkReference(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("testdata", "icon_resource_walk.json"))
	if err != nil {
		t.Fatal(err)
	}
	var golden []struct {
		Fixture, Profile, Name string
		Input                  struct {
			Bytes  int
			SHA256 string
		}
	}
	if err = json.Unmarshal(body, &golden); err != nil {
		t.Fatal(err)
	}
	fixtures := map[string][]byte{}
	for _, fixture := range iconResourceFixtures() {
		fixtures[fixture.name] = fixture.data
	}
	if len(golden) != 384 || len(fixtures) != 192 {
		t.Fatal("incomplete frozen resource oracle")
	}
	dbs := map[string]*Database{
		"flat":     iconDatabase(t, "Safe.Resource:A:B:"+resourceFlatHash+"\n", "Icon.Resource;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a"),
		"gradient": iconDatabase(t, "Safe.Resource:A:B:"+resourceGradientHash+"\n", "Icon.Resource;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a"),
	}
	seen := map[string]bool{}
	for _, row := range golden {
		t.Run(row.Profile+"/"+row.Fixture, func(t *testing.T) {
			data, ok := fixtures[row.Fixture]
			if !ok || dbs[row.Profile] == nil {
				t.Fatal("unknown oracle input")
			}
			key := row.Profile + "/" + row.Fixture
			if seen[key] {
				t.Fatal("duplicate oracle input")
			}
			seen[key] = true
			if len(data) != row.Input.Bytes || fmt.Sprintf("%x", sha256.Sum256(data)) != row.Input.SHA256 {
				t.Fatal("fixture differs from reference input")
			}
			r, err := dbs[row.Profile].MatchObject(context.Background(), data, MatchContext{Type: TypeMSEXE})
			if err != nil || r.Name != row.Name {
				t.Fatalf("resource oracle: name=%q error=%v want=%q", r.Name, err, row.Name)
			}
		})
	}
}

func repeatedIconLanguagePE(count int) []byte {
	le := binary.LittleEndian
	pe := labProgram([]labIcon{{w: 16, bpp: 24, rgba: make([][4]byte, 256)}})
	at := len(pe)
	pe = append(pe, make([]byte, 16+8*count)...)
	le.PutUint16(pe[at+14:], uint16(count))
	for i := 0; i < count; i++ {
		le.PutUint32(pe[at+16+8*i:], uint32(i))
		le.PutUint32(pe[at+20+8*i:], 128)
	}
	le.PutUint32(pe[0x234:], 0x80000000|uint32(at-0x200))
	v := len(pe) - 0x200
	le.PutUint32(pe[0x140:], uint32(v))
	le.PutUint32(pe[0x148:], uint32((v+511)/512*512))
	le.PutUint32(pe[0x58+116:], uint32(v))
	le.PutUint32(pe[0x58+56:], uint32(0x1000+(v+4095)/4096*4096))
	return pe
}

func TestIconResourceWorkLimitAndEarlyDetection(t *testing.T) {
	pe := repeatedIconLanguagePE(500)
	for _, tc := range []struct {
		name, hash string
		limit      int64
		want       string
		limited    bool
	}{
		{"first-match", resourceFlatHash, 500000, "Icon.Resource", false},
		{"nonmatching-aliases", resourceGradientHash, 500000, "", true},
		{"before-first-image", resourceFlatHash, 1000, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := iconDatabase(t, "Safe.Resource:A:B:"+tc.hash+"\n", "Icon.Resource;Engine:51-255,Target:1,IconGroup1:A,IconGroup2:B;0;4d5a")
			d.limits.MaxMatchSteps, d.limits.MatchStepsPerByte = tc.limit, 0
			r, err := d.MatchObject(context.Background(), pe, MatchContext{Type: TypeMSEXE})
			if tc.limited {
				if !errors.Is(err, ErrMatchLimit) || r.Name != "" {
					t.Fatalf("budget failure became verdict: %+v %v", r, err)
				}
			} else if err != nil || r.Name != tc.want {
				t.Fatalf("earlier detection erased by expensive tail: %+v %v", r, err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	meta := parsePE(pe)
	_, _, err := meta.programIconsWithWork(pe, &matchWork{ctx: ctx, limit: 500000000}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled resource walk continued", err)
	}
}

func FuzzIconResourceWalk(f *testing.F) {
	base := labProgram([]labIcon{{w: 16, bpp: 24, rgba: make([][4]byte, 256)}})
	f.Add([]byte{0xff, 0xff, 0, 0}, uint16(100))
	f.Add([]byte{0, 0, 1, 0}, uint16(0))
	f.Fuzz(func(t *testing.T, mutation []byte, seed uint16) {
		if len(mutation) > 1<<16 {
			t.Skip()
		}
		data := append([]byte(nil), base...)
		// Mutate resource directory/name/language/data bytes while retaining
		// a bounded modeled header. A fixed shared work budget bounds aliases.
		for i, b := range mutation {
			at := 0x200 + (int(seed)+i)%(len(data)-0x200)
			data[at] = b
		}
		meta := parsePE(data)
		if meta == nil || !meta.valid {
			return
		}
		work := &matchWork{ctx: context.Background(), limit: 200000}
		_, _, err := meta.programIconsWithWork(data, work, nil)
		if err != nil && !errors.Is(err, ErrMatchLimit) {
			t.Fatal(err)
		}
		if work.steps > work.limit {
			t.Fatal("resource work exceeded budget")
		}
	})
}
