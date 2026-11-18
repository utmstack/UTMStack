package engine

import (
	"encoding/base64"
	"testing"
)

// Encoded-script uncertainty must not hide independently supported views.
// These are synthetic signatures and inert text; no script is ever executed.
func TestHTMLUnsupportedScriptKeepsSupportedCoverage(t *testing.T) {
	const marker = "independenthtmlmarker"
	const unsupported = `<script>x=1;dF('bmfsu%26392%263%3A1');y=2;</script>`
	rules := map[string]string{"test.ndb": "Independent.HTML:0:*:" + hx(marker) + "\n"}
	for _, page := range []string{
		"<html><body>" + marker + unsupported + "</body></html>",
		"<html><body>" + unsupported + marker + "</body></html>",
	} {
		expectScan(t, pipelineState(t, rules), []byte(page), "Independent.HTML", "")
	}
	// dF now has a complete decoder; the benign page is decided clean.
	expectScan(t, pipelineState(t, rules), []byte("<html><body>ordinary text"+unsupported+"</body></html>"), "", "")
	// The reference's no-marker JScript.Encode path preserves the raw body
	// in its normalized page view, so this independent root-3 rule detects.
	expectScan(t, pipelineState(t, rules), []byte(`<html><script language="JScript.Encode">`+marker+`</script></html>`), "Independent.HTML", "")
}

func TestDFDecodedScriptReachesSignatureScanner(t *testing.T) {
	rules := map[string]string{"test.ndb": "Decoded.DF:3:*:" + hx("alert(1)") + "\n"}
	expectScan(t, pipelineState(t, rules), []byte(`<html><script>dF('bmfsu%26392%263%3A1');</script></html>`), "Decoded.DF", "")
	expectScan(t, pipelineState(t, rules), []byte(`<html><script>dF('cngtv*4+=2');</script></html>`), "", "")
}

func TestHTMLMalformedBaseKeepsEmbeddedCoverage(t *testing.T) {
	const marker = "independentembeddedmarker"
	const unsupported = `<script>dF('bmfsu%26392%263%3A1');</script>`
	rules := map[string]string{
		"test.ndb": "Independent.Embedded:0:*:" + hx(marker) + "\n",
		"test.pdb": "H:amazon.com\n",
	}
	// A malformed base without any links is irrelevant to link analysis.
	// The dF script has known semantics, and the data URI remains scannable.
	page := `<html><base href="http://[broken"><body>` + unsupported + `<img src="data:application/octet-stream;base64,` + base64.StdEncoding.EncodeToString([]byte(marker)) + `"></body></html>`
	expectScan(t, pipelineState(t, rules), []byte(page), "Independent.Embedded", "")
	// The same page without a matching embedded payload is clean.
	benign := `<html><base href="http://[broken"><body>` + unsupported + `<img src="data:application/octet-stream;base64,` + base64.StdEncoding.EncodeToString([]byte("ordinary benign text")) + `"></body></html>`
	expectScan(t, pipelineState(t, rules), []byte(benign), "", "")
	// Continuing independent analysis must retain the recursive scan limits.
	s := pipelineState(t, rules)
	s.maxDepth = 0
	expectIncomplete(t, s, []byte(page), ErrLimit)
}
