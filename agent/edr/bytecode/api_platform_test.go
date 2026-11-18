package bytecode

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPlatformHostAPIs(t *testing.T) {
	var mem memory
	state := &apiState{}
	for _, tc := range []struct {
		left, right string
		want        uint64
	}{
		{"1.2", "1.10", apiFailure},
		{"2.10", "2.2", 1},
		{"v1.2", "v1.2", 0},
		{"v1.2", "v1.2.0", apiFailure},
		{"v2", "V2", 1},
		// The pinned reference retains its numeric accumulator across
		// fields: 1 then 09 compares as 109, not 9.
		{"1.09", "1.9", 1},
	} {
		left := apiTestBytes(t, &mem, []byte(tc.left))
		right := apiTestBytes(t, &mem, []byte(tc.right))
		apiTestCall(t, state, &mem, 75, tc.want, left, uint64(len(tc.left)), right, uint64(len(tc.right)))
	}
	valid := apiTestBytes(t, &mem, []byte("1.0"))
	for _, args := range [][]uint64{{0, 3, valid, 3}, {valid, 3, 0, 3}} {
		if _, err := state.invoke(context.Background(), 75, args, &mem, func(uint64) error { return nil }); !errors.Is(err, ErrRuntime) {
			t.Fatalf("version_compare invalid pointer %v: %v; want ErrRuntime", args, err)
		}
	}

	env := environmentRecord()
	a := binary.LittleEndian.Uint32(env)
	b := binary.LittleEndian.Uint32(env[4:])
	c := binary.LittleEndian.Uint32(env[8:])
	apiTestCall(t, state, nil, 76, 1, uint64(a), uint64(b), uint64(c))
	apiTestCall(t, state, nil, 76, 1, 0xffffffff, 0xffffffff, 0xffffffff)
	wrongOS := (a & 0x00ffffff) | uint32((byte(a>>24)+1)%0xff)<<24
	apiTestCall(t, state, nil, 76, 0, uint64(wrongOS), uint64(b), uint64(c))
	apiTestCall(t, state, nil, 90, 0)
}

func TestPlatformHostAPIsSerializedDeclarations(t *testing.T) {
	for _, tc := range []struct {
		id   int
		name string
		typ  string
		args []string
		want uint64
	}{
		{75, "version_compare", memoryType(1, 5, 32, 65, 32, 65, 32), nil, 0},
		{76, "check_platform", memoryType(1, 4, 32, 32, 32, 32), []string{
			compatConstant(0xffffffff, 32), compatConstant(0xffffffff, 32), compatConstant(0xffffffff, 32),
		}, 1},
		{90, "running_on_jit", memoryType(1, 1, 32), []string{}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := memoryFixture{types: []string{tc.typ}, apiMax: tc.id,
				apis: []string{memoryAPI(tc.id, 69, tc.name)}}
			if tc.args == nil {
				m.functions = []compatFunction{{result: 32, blocks: [][]string{{
					compatRet(32, compatConstant(0, 32)),
				}}}}
				memoryParse(t, m) // declaration is accepted without executing it
				return
			}
			m.functions = []compatFunction{{locals: []int{32}, result: 32, blocks: [][]string{{
				memoryAPICall(0, tc.id, tc.args...), compatRet(32, compatNum(0)),
			}}}}
			p := memoryParse(t, m)
			got, err := p.Run(context.Background(), 0, nil, 1000)
			if err != nil || got != tc.want {
				t.Fatalf("wire API %d = %#x, %v; want %#x", tc.id, got, err, tc.want)
			}
		})
	}
}

// The output is an intentionally harmless unsigned ClamBC fixture for a
// network-disabled pinned ClamAV differential. The opt-in directory must be
// supplied by the caller; ordinary tests never write this artifact.
func TestWritePlatformOracleFixture(t *testing.T) {
	dir := os.Getenv("EDR_PLATFORM_ORACLE_DIR")
	if dir == "" {
		t.Skip("set EDR_PLATFORM_ORACLE_DIR for the native ClamAV oracle")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	write := func(file, label string, m memoryFixture, detected bool) {
		t.Helper()
		m.kind = 256
		m.logical = label + ";Engine:56-255,Target:0;0;4142"
		p, err := Parse(m.encode())
		if err != nil {
			t.Fatal(err)
		}
		result, err := p.RunWithInput(context.Background(), 0, nil, []byte("AB"), 1000)
		if err != nil || (result.VirusName == label) != detected {
			t.Fatalf("Go oracle %s: %+v, %v", file, result, err)
		}
		if err := os.WriteFile(filepath.Join(dir, file), m.encode(), 0600); err != nil {
			t.Fatal(err)
		}
	}
	const jitLabel = "UTM.JIT"
	write("jit.cbc", jitLabel, engineLevelProgram(90, "running_on_jit", 0, jitLabel), true)

	platform := func(label string, a, b, c uint64) memoryFixture {
		m := engineLevelProgram(76, "check_platform", 1, label)
		m.types[2] = memoryType(1, 4, 32, 32, 32, 32)
		m.functions[0].blocks[0][0] = memoryAPICall(0, 76, compatConstant(a, 32),
			compatConstant(b, 32), compatConstant(c, 32))
		return m
	}
	const platformLabel = "UTM.Platform"
	write("platform.cbc", platformLabel, platform(platformLabel, 0xffffffff, 0xffffffff, 0xffffffff), true)
	const missLabel = "UTM.PlatMiss"
	write("platform-miss.cbc", missLabel, platform(missLabel, 0, 0xffffffff, 0xffffffff), false)
	// The retained portable reference is macOS ARM64 with this exact
	// environment profile. Keep the fixture host-specific; wildcard and
	// mismatch checks above remain portable.
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		const exactLabel = "UTM.PlatExact"
		write("platform-exact.cbc", exactLabel,
			platform(exactLabel, 0x0403eaea, 0x08000000, 0x00040201), true)
	}

	const versionLabel = "UTM.Version"
	left, right := "1.09", "1.9"
	version := engineLevelProgram(75, "version_compare", 1, versionLabel)
	version.types[2] = memoryType(1, 5, 32, 65, 32, 65, 32)
	version.globals = append(version.globals,
		memoryGlobal(69, stringComponents(left, 16)...), memoryGlobal(65, 0, 3),
		memoryGlobal(69, stringComponents(right, 16)...), memoryGlobal(65, 0, 5))
	version.functions[0].blocks[0][0] = memoryAPICall(0, 75,
		memoryGlobalRef(4), compatConstant(uint64(len(left)), 32),
		memoryGlobalRef(6), compatConstant(uint64(len(right)), 32))
	write("version.cbc", versionLabel, version, true)

	if err := os.WriteFile(filepath.Join(dir, "sample.bin"), append([]byte("AB"), bytes.Repeat([]byte{'X'}, 1024)...), 0600); err != nil {
		t.Fatal(err)
	}
}

func stringComponents(s string, size int) []uint64 {
	result := make([]uint64, size)
	for i := 0; i < len(s); i++ {
		result[i] = uint64(s[i])
	}
	return result
}
