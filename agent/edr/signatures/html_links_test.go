package signatures

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

// sig-6: benign in-memory text under the HTML byte limit must not become
// incomplete only because one text or raw-text run exceeds a small tokenizer
// buffer. Links after a long run are still extracted.
func TestLargeTextRunIsNotIncomplete(t *testing.T) {
	d, err := Compile(map[string][]byte{"p.pdb": []byte("H:paypal.com\n")})
	if err != nil {
		t.Fatal(err)
	}
	script := append(bytes.Repeat([]byte("x = 1 + 2;\n"), 30000), []byte("if (a < b) {}\n")...)
	page := append(append([]byte("<html><body><script>"), bytes.Repeat([]byte("var q = 1;\n"), 30000)...), []byte("</script><p>hello</p></body></html>")...)
	for name, data := range map[string][]byte{"plain script": script, "inline script HTML": page} {
		got, err := d.MatchHTML(context.Background(), data)
		if got != "" || err != nil {
			t.Errorf("%s (%d bytes): benign content without links returned %q, %v", name, len(data), got, err)
		}
	}
	spoof := append(append([]byte(nil), page...), `<a href="https://evil.test">paypal.com</a>`...)
	if got, err := d.MatchHTML(context.Background(), spoof); got != "Heuristics.Phishing.Email.SpoofedDomain" || err != nil {
		t.Fatalf("link after a long script: got %q, %v", got, err)
	}
	// Tags are decoded into strings, so an oversized tag stays incomplete.
	tag := append(append([]byte(`<a title="`), bytes.Repeat([]byte("x"), maxHTMLTagBytes)...), `">paypal.com</a>`...)
	if got, err := d.MatchHTML(context.Background(), tag); got != "" || !errors.Is(err, ErrHTMLIncomplete) {
		t.Fatalf("oversized tag: got %q, %v; want incomplete", got, err)
	}
}
