package amsi

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
)

func pipeTestFrame(versioned bool, status byte, app string, content []byte) []byte {
	payload := append(append([]byte(app), 0), content...)
	if versioned {
		payload = append(append([]byte(pipeMagic), status), payload...)
	}
	n := uint32(len(payload))
	if versioned {
		n |= pipeVersionFlag
	}
	frame := make([]byte, 4, 4+len(payload))
	binary.LittleEndian.PutUint32(frame, n)
	return append(frame, payload...)
}

type oneByteReader struct{ io.Reader }

func TestPipeClientAccessExcludesServerCreationAndSecurityChanges(t *testing.T) {
	// Individual Windows rights: data read/write, read attributes, owner/DACL
	// query and synchronization. Generic write would also permit new instances.
	want := uint32(0x1 | 0x2 | 0x80 | 0x20000 | 0x100000)
	forbidden := uint32(0x4 | 0x40000 | 0x80000 | 0x10000000 | 0x40000000)
	if pipeClientRights != want || pipeClientRights&forbidden != 0 {
		t.Fatalf("client access %#x grants unexpected rights", pipeClientRights)
	}
}

func (r oneByteReader) Read(p []byte) (int, error) {
	return r.Reader.Read(p[:min(1, len(p))])
}

func TestPipeRequestCompleteAndIncompleteFrames(t *testing.T) {
	content := append(bytes.Repeat([]byte("ordinary harmless text\x00"), 4000), []byte("END-OF-COMPLETE-BUFFER")...)
	for _, versioned := range []bool{false, true} {
		for status := pipeComplete; status <= pipeInputTimeout; status++ {
			frame := pipeTestFrame(versioned, status, "PowerShell-\u03b1", content)
			got, err := readPipeRequest(oneByteReader{bytes.NewReader(frame)})
			wantStatus := status
			if !versioned {
				wantStatus = pipeComplete
			}
			if err != nil || got.versioned != versioned || got.status != wantStatus || got.appName != "PowerShell-\u03b1" || !bytes.Equal(got.content, content) {
				t.Fatalf("versioned=%v status=%d: %+v %v", versioned, status, got, err)
			}
		}
	}
	got, err := readPipeRequest(bytes.NewReader(pipeTestFrame(true, pipeComplete, "", nil)))
	if err != nil || got.appName != "" || len(got.content) != 0 {
		t.Fatalf("empty complete buffer: %+v %v", got, err)
	}
}

func TestPipeRequestRejectsMalformedBeforeScanning(t *testing.T) {
	valid := pipeTestFrame(true, pipeComplete, "host", []byte("inert content"))
	badMagic := bytes.Clone(valid)
	badMagic[4] = 'X'
	badStatus := bytes.Clone(valid)
	badStatus[8] = 255
	noDelimiter := append([]byte(nil), valid[:len(valid)-len("inert content")-1]...)
	binary.LittleEndian.PutUint32(noDelimiter, pipeVersionFlag|uint32(len(noDelimiter)-4))
	var oversized [4]byte
	binary.LittleEndian.PutUint32(oversized[:], pipeVersionFlag|uint32(maxPipePayload+1))
	for name, frame := range map[string][]byte{
		"bad-magic": badMagic, "bad-status": badStatus,
		"no-delimiter": noDelimiter, "zero-length": {0, 0, 0, 0},
		"short-header": {1, 0}, "short-payload": valid[:len(valid)-1],
		"invalid-utf8":        pipeTestFrame(true, pipeComplete, "\xff", nil),
		"long-name":           pipeTestFrame(true, pipeComplete, strings.Repeat("x", maxPipeAppName+1), nil),
		"legacy-no-delimiter": {3, 0, 0, 0, 'a', 'b', 'c'},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := readPipeRequest(bytes.NewReader(frame)); !errors.Is(err, engine.ErrMalformed) {
				t.Fatalf("error=%v; want malformed", err)
			}
		})
	}
	// A claimed oversized body is rejected from its header without consuming
	// or allocating that body. The fixture itself is just four safe bytes.
	if _, err := readPipeRequest(bytes.NewReader(oversized[:])); !errors.Is(err, engine.ErrLimit) {
		t.Fatalf("oversized claim: %v", err)
	}
}

func TestPipeDecisionNeverScansIncompletePrefixes(t *testing.T) {
	for _, mode := range []string{"open", "closed"} {
		for _, status := range []byte{pipeSizeLimit, pipeReadFailure, pipeInvalidAttributes, pipeInputTimeout} {
			cfg := config.Default()
			cfg.FailMode = mode
			s, path := newTestScanner(t, cfg)
			s.scanSnapshot = func(context.Context, []byte) (bool, string, string, error) {
				t.Fatal("incomplete provider prefix reached the engine")
				return false, "", "", nil
			}
			req := pipeRequest{versioned: true, status: status, appName: "host", content: []byte("UNSCANNED-PREFIX")}
			want := pipeIncompleteAllow
			if mode == "closed" {
				want = pipeIncompleteBlock
			}
			if got := s.pipeDecision(context.Background(), req, nil); got != want {
				t.Fatalf("mode=%s status=%d: reply=%d want=%d", mode, status, got, want)
			}
			lines := spoolLines(t, path)
			if len(lines) != 1 || !strings.Contains(lines[0], `"action":"health"`) || !strings.Contains(lines[0], "Analysis.Incomplete.") || strings.Contains(lines[0], `"file_hash":"`) || strings.Contains(lines[0], "UNSCANNED-PREFIX") {
				t.Fatalf("incomplete input reporting: %v", lines)
			}
		}
	}
}

func TestPipeDecisionPreservesActualEngineVerdictAndFailPolicy(t *testing.T) {
	for _, tc := range []struct {
		name      string
		clean     bool
		sig       string
		err       error
		wantOpen  byte
		wantClose byte
	}{
		{"clean", true, "", nil, pipeAllow, pipeAllow},
		{"detected", false, "Test.Harmless.Marker", nil, pipeBlock, pipeBlock},
		{"unsupported", false, "", engine.ErrUnsupported, pipeIncompleteAllow, pipeIncompleteBlock},
		{"limit", false, "", engine.ErrLimit, pipeIncompleteAllow, pipeIncompleteBlock},
	} {
		for _, mode := range []string{"open", "closed"} {
			for _, versioned := range []bool{false, true} {
				t.Run(tc.name+"/"+mode+"/"+map[bool]string{true: "v2", false: "legacy"}[versioned], func(t *testing.T) {
					cfg := config.Default()
					cfg.FailMode = mode
					s, _ := newTestScanner(t, cfg)
					ctx := context.WithValue(context.Background(), struct{}{}, "caller")
					called := 0
					s.scanSnapshot = func(actual context.Context, data []byte) (bool, string, string, error) {
						called++
						if actual != ctx || string(data) != "entire input" {
							t.Fatal("caller context or complete bytes changed")
						}
						return tc.clean, tc.sig, "actual-generation", tc.err
					}
					want := tc.wantOpen
					if mode == "closed" {
						want = tc.wantClose
					}
					if !versioned && want >= pipeIncompleteAllow {
						want -= pipeIncompleteAllow // old DLLs understand only 0/1
					}
					got := s.pipeDecision(ctx, pipeRequest{versioned: versioned, appName: "host", content: []byte("entire input")}, nil)
					if got != want || called != 1 {
						t.Fatalf("reply=%d want=%d calls=%d", got, want, called)
					}
				})
			}
		}
	}
}

func TestPipeDecisionInputErrorsAndCancellationAreIncomplete(t *testing.T) {
	s, _ := newTestScanner(t, config.Default())
	s.scanSnapshot = func(context.Context, []byte) (bool, string, string, error) {
		t.Fatal("invalid frame reached scanner")
		return true, "", "", nil
	}
	if got := s.pipeDecision(context.Background(), pipeRequest{versioned: true}, engine.ErrMalformed); got != pipeIncompleteAllow {
		t.Fatal(got)
	}
	if got := s.pipeDecision(context.Background(), pipeRequest{versioned: true, status: 255}, nil); got != pipeIncompleteAllow {
		t.Fatal(got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := s.pipeDecision(ctx, pipeRequest{versioned: true}, nil); got != pipeIncompleteAllow {
		t.Fatal(got)
	}
}

func FuzzPipeRequest(f *testing.F) {
	f.Add(pipeTestFrame(true, pipeComplete, "host", []byte("inert content")))
	f.Add(pipeTestFrame(false, pipeComplete, "", nil))
	f.Add([]byte{255, 255, 255, 255})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 8192 {
			t.Skip()
		}
		// Avoid allocating a large declared-but-absent body during fuzzing;
		// oversized claims have a separate deterministic rejection test.
		if len(data) >= 4 && binary.LittleEndian.Uint32(data[:4])&^pipeVersionFlag > 8192 {
			return
		}
		req, err := readPipeRequest(oneByteReader{bytes.NewReader(data)})
		if err == nil && (len(req.appName) > maxPipeAppName || len(req.content) > maxPipeContent || req.status > pipeInputTimeout) {
			t.Fatalf("invalid successful frame: %+v", req)
		}
	})
}
