package engine

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/chm/chmtest"
)

func TestCHMEntriesAreScanned(t *testing.T) {
	page := bytes.Repeat([]byte("<p>help page text</p>"), 50)
	hidden := append(append([]byte("text before "), payloadMarker...), " and after"...)
	b := chmtest.Build([]chmtest.Entry{
		{Name: "/#SYSTEM", Data: []byte("system data")},
		{Name: "/page.htm", Section: 1, Data: page},
		{Name: "/payload.txt", Section: 1, Data: hidden},
	})
	requireHidden(t, b)
	expectDetected(t, payloadState(t), b, "Payload.Test")
	// Entries are matched by their metadata, positions counted from 0.
	rule := fmt.Sprintf("Chm.Entry:CL_TYPE_MSCHM:*:^/payload\\.txt$:0:%d:0:2:*:\n", len(hidden))
	ftm := "0:0:49545346:CHM:CL_TYPE_ANY:CL_TYPE_MSCHM\n"
	expectScan(t, pipelineState(t, map[string]string{"a.cdb": rule, "a.ftm": ftm}), b, "Chm.Entry", "")
	// A directory this reader cannot parse is not decided.
	broken := append([]byte(nil), b...)
	broken[0x78] = 'X' // the ITSP signature
	expectIncomplete(t, payloadState(t), broken, ErrMalformed)
}
