package engine

import (
	"archive/zip"
	"bytes"
	"errors"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// A matched PDF HandlerType signature duplicates the object with PDF forced.
// The child then scans generic rules with the original type as its container.
func TestPDFHandlerTypeForcedRescan(t *testing.T) {
	const marker = "ABCDEFGH"
	const handler = "Test.Handler.PDF;Engine:51-255,HandlerType:CL_TYPE_PDF,Target:0;0;4142434445464748\n"
	data := []byte("hello " + marker + " world\n")
	s := pipelineState(t, map[string]string{"probe.ldb": handler})
	if n := s.db.Report().Unknown[signatures.UnknownHandlerType]; n != 0 {
		t.Fatalf("PDF handler still unknown: %d", n)
	}
	expectScan(t, s, data, "Test.Handler.PDF", "")
	if s.layers.Depth() != 0 {
		t.Fatalf("leaked handler layer: %d", s.layers.Depth())
	}

	// ClamAV 1.5.4 reports Test.Child on these same two rules: the
	// retyped child reaches its Container:TEXT_ASCII rule first.
	s = pipelineState(t, map[string]string{"probe.ldb": "Test.Child;Engine:51-255,Container:CL_TYPE_TEXT_ASCII,Target:0;0;4142434445464748\n" + handler})
	expectScan(t, s, data, "Test.Child", "")

	// The duplicate layer is subject to the same recursion bound.
	s = pipelineState(t, map[string]string{"probe.ldb": handler})
	s.maxDepth = 0
	if _, err := s.visit(data, 0); !errors.Is(err, ErrLimit) {
		t.Fatalf("depth-limited handler rescan: %v, want ErrLimit", err)
	}
}

func TestZIPHandlerTypeScansMembersBeforeRaw(t *testing.T) {
	const handler = "Test.Handler.ZIP;Engine:51-255,HandlerType:CL_TYPE_ZIP,Target:0;0;4142434445464748\n"
	data := []byte("hello ABCDEFGH world\n")
	s := pipelineState(t, map[string]string{"probe.ldb": handler})
	if n := s.db.Report().Unknown[signatures.UnknownHandlerType]; n != 0 {
		t.Fatalf("ZIP handler still unknown: %d", n)
	}
	expectScan(t, s, data, "Test.Handler.ZIP", "")

	// The prefixed archive initially types as text. HandlerType forces the
	// ZIP parser, which scans its child before the duplicate's raw rules.
	buf := bytes.NewBuffer(append([]byte(nil), data...))
	writer := zip.NewWriter(buf)
	writer.SetOffset(int64(len(data)))
	member, err := writer.Create("inner.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := member.Write([]byte("CHILDMARKER content\n")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	s = pipelineState(t, map[string]string{"probe.ldb": "Test.Zip.Child;Engine:51-255,Container:CL_TYPE_ZIP,Target:0;0;4348494c444d41524b4552\n" + handler})
	expectScan(t, s, buf.Bytes(), "Test.Zip.Child", "")
}

func TestGraphicsHandlerTypeForcedRescan(t *testing.T) {
	const handler = "Test.Handler.Graphic;Engine:51-255,HandlerType:CL_TYPE_GRAPHICS,Target:0;0;4142434445464748\n"
	data := []byte("hello ABCDEFGH world\n")
	s := pipelineState(t, map[string]string{"probe.ldb": handler})
	if n := s.db.Report().Unknown[signatures.UnknownHandlerType]; n != 0 {
		t.Fatalf("graphics handler still unknown: %d", n)
	}
	expectScan(t, s, data, "Test.Handler.Graphic", "")
	// The forced graphics root precedes the generic HandlerType rule.
	s = pipelineState(t, map[string]string{"probe.ldb": "Test.Graphic.Child;Engine:51-255,Target:5;0;4142434445464748\n" + handler})
	expectScan(t, s, data, "Test.Graphic.Child", "")
}

func TestScriptHandlerTypeNormalizesBeforeRaw(t *testing.T) {
	const handler = "Test.Handler.Script;Engine:51-255,HandlerType:CL_TYPE_SCRIPT,Target:0;0;4142434445464748\n"
	data := []byte("hello ABCDEFGH VIEWWORD world\n")
	s := pipelineState(t, map[string]string{"probe.ldb": handler})
	if n := s.db.Report().Unknown[signatures.UnknownHandlerType]; n != 0 {
		t.Fatalf("script handler still unknown: %d", n)
	}
	expectScan(t, s, data, "Test.Handler.Script", "")
	// ClamAV 1.5.4 checks the script's normalized text before the raw
	// generic HandlerType rule on its forced duplicate.
	s = pipelineState(t, map[string]string{"probe.ldb": "Test.Script.View;Engine:51-255,Target:7;0;76696577776f7264\n" + handler})
	expectScan(t, s, data, "Test.Script.View", "")
}
