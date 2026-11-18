package bytecode

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestAPIJSNormalizerFilePipeAndScanLifecycle(t *testing.T) {
	input := bytes.Repeat([]byte("var x=1;"), int(pipeFileWindow())/8+1)
	var mem memory
	s := &apiState{input: input, scanning: true}
	var scanned [][]byte
	s.scanExtracted = func(data []byte, id int32) (string, error) {
		if id != -1 {
			t.Fatalf("normalized child id=%d", id)
		}
		scanned = append(scanned, bytes.Clone(data))
		return "JS.Child.Test", nil
	}
	apiTestCall(t, s, &mem, 43, apiFailure, 0) // no pipe
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 43, 0, 0)
	apiTestCall(t, s, &mem, 44, apiFailure, 1) // invalid normalizer
	for s.pipes[0].read < uint32(len(input)) {
		before := s.pipes[0].read
		apiTestCall(t, s, &mem, 44, 0, 0)
		if s.pipes[0].read <= before {
			t.Fatal("jsnorm_process did not advance file pipe")
		}
	}
	apiTestCall(t, s, &mem, 44, apiFailure, 0) // empty input
	apiTestCall(t, s, &mem, 45, 0, 0)
	apiTestCall(t, s, &mem, 45, apiFailure, 0)
	if len(scanned) != 0 || len(s.jsFiles) != 1 {
		t.Fatalf("normalizer scanned before VM completion: scanned=%d queued=%d", len(scanned), len(s.jsFiles))
	}
	if err := s.scanJSFiles(); err != nil || s.childName != "JS.Child.Test" || len(scanned) != 1 ||
		!bytes.HasPrefix(scanned[0], []byte("<script>")) || !bytes.HasSuffix(scanned[0], []byte("</script>")) {
		t.Fatalf("normalized child: name=%q scanned=%q err=%v", s.childName, scanned, err)
	}
}

func TestAPIJSNormalizerSerializedChildScanner(t *testing.T) {
	input := []byte("danger()")
	m := memoryFixture{
		types:  []string{memoryType(1, 2, 32, 32)},
		apiMax: 45,
		apis: []string{
			memoryAPI(31, 69, "buffer_pipe_new_fromfile"),
			memoryAPI(43, 69, "jsnorm_init"),
			memoryAPI(44, 69, "jsnorm_process"),
			memoryAPI(45, 69, "jsnorm_done"),
		},
		functions: []compatFunction{{locals: []int{32, 32, 32, 32}, result: 32, blocks: [][]string{{
			memoryAPICall(0, 31, compatConstant(0, 32)),
			memoryAPICall(1, 43, compatNum(0)),
			memoryAPICall(2, 44, compatNum(1)),
			memoryAPICall(3, 45, compatNum(1)),
			compatRet(32, compatNum(3)),
		}}}},
	}
	p := memoryParse(t, m)
	var seen []byte
	result, err := p.RunWithMatchContextAndScanner(context.Background(), 0, nil, input, EmptyMatchContext(),
		func(data []byte, _ int32) (string, error) {
			seen = bytes.Clone(data)
			return "JS.Child.Test", nil
		}, 1000)
	if err != nil || result.ChildName != "JS.Child.Test" || !bytes.Equal(seen, []byte("<script>danger()</script>")) {
		t.Fatalf("serialized jsnorm: result=%+v output=%q err=%v", result, seen, err)
	}
	if _, err := p.RunWithMatchContext(context.Background(), 0, nil, input, EmptyMatchContext(), 1000); !errors.Is(err, ErrExtraction) {
		t.Fatalf("missing child scanner = %v", err)
	}
}

func TestAPIJSNormalizerMemoryPipeRefill(t *testing.T) {
	var mem memory
	s := &apiState{scanning: true, scanExtracted: func(data []byte, _ int32) (string, error) {
		if !bytes.Equal(data, []byte("<script>danger()</script>")) {
			t.Fatalf("normalized output = %q", data)
		}
		return "JS.Child.Test", nil
	}}
	apiTestCall(t, s, &mem, 30, 0, 8)
	apiTestCall(t, s, &mem, 43, 0, 0)
	for _, part := range [][]byte{[]byte("danger()"), []byte(" ")} {
		ptr, err := s.invoke(context.Background(), 36, []uint64{0, uint64(len(part))}, &mem, func(uint64) error { return nil })
		if err != nil || ptr == 0 || mem.write(ptr, part) != nil {
			t.Fatalf("write_get = %#x, %v", ptr, err)
		}
		apiTestCall(t, s, &mem, 37, 0, 0, uint64(len(part)))
		apiTestCall(t, s, &mem, 44, 0, 0)
	}
	apiTestCall(t, s, &mem, 45, 0, 0)
	if err := s.scanJSFiles(); err != nil || s.childName != "JS.Child.Test" {
		t.Fatalf("memory pipe child = %q, %v", s.childName, err)
	}
}

func TestAPIJSNormalizerUnmodeledTransformStaysIncomplete(t *testing.T) {
	var mem memory
	s := &apiState{input: []byte(`dF("x", "extra")`), scanning: true,
		scanExtracted: func([]byte, int32) (string, error) { return "", nil }}
	apiTestCall(t, s, &mem, 31, 0, 0)
	apiTestCall(t, s, &mem, 43, 0, 0)
	apiTestCall(t, s, &mem, 44, 0, 0)
	apiTestCall(t, s, &mem, 45, 0, 0)
	if err := s.scanJSFiles(); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("malformed transform = %v; want incomplete", err)
	}
}
