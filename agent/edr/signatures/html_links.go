package signatures

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

const maxHTMLBytes = 32 << 20
const maxHTMLPairs = 10000
const maxHTMLTokens = 1000000
const maxHTMLTagBytes = 256 << 10

type htmlLinkPair struct {
	real, display string
	ssl           bool
}
type htmlAnchor struct {
	destination string
	text        strings.Builder
}

// extractHTMLLinks is a streaming extraction pass, not a browser or DOM renderer.
// It bounds token, attribute, text and pair growth, decodes entities once via the
// HTML tokenizer, and uses the first base href when resolving relative links.
func extractHTMLLinks(ctx context.Context, data []byte, keepEmptyDisplay bool) ([]htmlLinkPair, string, error) {
	if len(data) > maxHTMLBytes {
		return nil, "", fmt.Errorf("%w: HTML byte limit exceeded", ErrHTMLIncomplete)
	}
	// The input is already in memory and capped at maxHTMLBytes, so a text or
	// script run may span all of it: that is ordinary content, not a reason to
	// leave the object incomplete. Tags, which are decoded into strings, keep a
	// per-token bound below, and link and display text keep their own limits.
	tokenizer := html.NewTokenizer(bytes.NewReader(data))
	tokenizer.SetMaxBuf(maxHTMLBytes + 1)
	var pairs []htmlLinkPair
	var anchor *htmlAnchor
	var form, base string
	baseSeen := false
	skipTag := ""
	skipDepth := 0
	add := func(real, display string, ssl bool) error {
		if real == "" || display == "" && !keepEmptyDisplay {
			return nil
		}
		if len(real) > maxPhishingURL || len(display) > maxPhishingURL {
			return fmt.Errorf("%w: URL length limit exceeded", ErrHTMLIncomplete)
		}
		if len(pairs) >= maxHTMLPairs {
			return fmt.Errorf("%w: URL pair limit exceeded", ErrHTMLIncomplete)
		}
		pairs = append(pairs, htmlLinkPair{real, display, ssl})
		return nil
	}
	flush := func() error {
		if anchor == nil {
			return nil
		}
		err := add(anchor.destination, anchor.text.String(), true)
		anchor = nil
		return err
	}
	for tokens := 0; ; tokens++ {
		if err := ctx.Err(); err != nil {
			return nil, "", err
		}
		if tokens >= maxHTMLTokens {
			return nil, "", fmt.Errorf("%w: HTML token count limit exceeded", ErrHTMLIncomplete)
		}
		kind := tokenizer.Next()
		if kind == html.ErrorToken {
			if err := tokenizer.Err(); err != io.EOF {
				return nil, "", fmt.Errorf("%w: HTML tokenizer: %v", ErrHTMLIncomplete, err)
			}
			if err := flush(); err != nil {
				return nil, "", err
			}
			return pairs, base, nil
		}
		switch kind {
		case html.TextToken:
			if anchor != nil && skipDepth == 0 {
				text := tokenizer.Text()
				if anchor.text.Len()+len(text) > maxPhishingURL {
					return nil, "", fmt.Errorf("%w: link display text limit exceeded", ErrHTMLIncomplete)
				}
				anchor.text.Write(text)
			}
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			if len(tokenizer.Raw()) >= maxHTMLTagBytes {
				return nil, "", fmt.Errorf("%w: HTML tag size limit exceeded", ErrHTMLIncomplete)
			}
			token := tokenizer.Token()
			tag := token.Data
			if skipDepth > 0 {
				if tag == skipTag {
					if kind == html.EndTagToken {
						skipDepth--
					} else if kind == html.StartTagToken {
						skipDepth++
					}
					if skipDepth > 64 {
						return nil, "", fmt.Errorf("%w: HTML hidden-content nesting limit exceeded", ErrHTMLIncomplete)
					}
				}
				continue
			}
			if (tag == "script" || tag == "style" || tag == "template" || tag == "noscript") && kind == html.StartTagToken {
				skipTag, skipDepth = tag, 1
				continue
			}
			if kind == html.EndTagToken {
				switch tag {
				case "a":
					if err := flush(); err != nil {
						return nil, "", err
					}
				case "form":
					form = ""
				}
				continue
			}
			attrs := make(map[string]string, len(token.Attr))
			for _, attribute := range token.Attr {
				key := attribute.Key
				if _, exists := attrs[key]; exists {
					switch key {
					case "href", "action", "src", "dynsrc", "title":
						return nil, "", fmt.Errorf("%w: duplicate URL attribute", ErrHTMLIncomplete)
					}
					continue
				}
				attrs[key] = attribute.Val
			}
			switch tag {
			case "base":
				if !baseSeen {
					if href, ok := attrs["href"]; ok {
						base, baseSeen = href, true
					}
				}
			case "a":
				if err := flush(); err != nil {
					return nil, "", err
				}
				href := attrs["href"]
				if err := add(form, href, false); err != nil {
					return nil, "", err
				}
				if err := add(href, attrs["title"], true); err != nil {
					return nil, "", err
				}
				anchor = &htmlAnchor{destination: href}
				if kind == html.SelfClosingTagToken {
					anchor = nil
				}
			case "form":
				if form != "" {
					return nil, "", fmt.Errorf("%w: nested form origins are ambiguous", ErrHTMLIncomplete)
				}
				form = attrs["action"]
			case "img", "area", "iframe":
				parent := form
				if anchor != nil {
					parent = anchor.destination
				}
				for _, key := range []string{"src", "dynsrc"} {
					if err := add(parent, attrs[key], false); err != nil {
						return nil, "", err
					}
				}
				if tag == "area" {
					if err := add(parent, attrs["href"], false); err != nil {
						return nil, "", err
					}
				}
			}
		}
	}
}
