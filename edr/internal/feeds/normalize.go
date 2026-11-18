// edr/internal/feeds/normalize.go
// Package feeds syncs ThreatWinds indicator lists into the mirror tree in the
// exact wire format the agent's netblock feed client consumes.
package feeds

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"sort"
	"strings"
)

// Normalize converts an upstream accumulative artifact — a gzip'd tarball of
// member lists, or an already-gzip'd plain list — into the agent wire format:
// a gzip'd, newline-separated, deduped, sorted value list. Sorting makes the
// output bytes (and therefore the .sha256 sibling) deterministic.
func Normalize(upstream []byte) ([]byte, int, error) {
	vals, err := extractValues(upstream)
	if err != nil {
		return nil, 0, err
	}
	return gzipValues(vals), len(vals), nil
}

// extractValues gunzips an upstream accumulative artifact — a gzip'd tarball of
// member lists, or an already-gzip'd plain list — then trims each line, drops
// blanks and '#' comments, dedupes, and returns the values sorted.
func extractValues(upstream []byte) ([]string, error) {
	gr, err := gzip.NewReader(bytes.NewReader(upstream))
	if err != nil {
		return nil, errors.New("upstream artifact is not gzip")
	}
	raw, err := io.ReadAll(gr)
	if err != nil {
		return nil, err
	}

	// A tarball's members are concatenated; anything else is one plain list.
	var text strings.Builder
	tr := tar.NewReader(bytes.NewReader(raw))
	isTar := false
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		isTar = true
		b, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		text.Write(b)
		text.WriteByte('\n')
	}
	if !isTar {
		text.Reset()
		text.Write(raw)
	}

	seen := map[string]struct{}{}
	var values []string
	for _, line := range strings.Split(text.String(), "\n") {
		v := strings.TrimSpace(line)
		if v == "" || strings.HasPrefix(v, "#") {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		values = append(values, v)
	}
	sort.Strings(values)
	return values, nil
}

// gzipValues encodes a value list into the agent wire format: a gzip'd,
// newline-terminated list (one value per line).
func gzipValues(values []string) []byte {
	var out bytes.Buffer
	gw := gzip.NewWriter(&out)
	for _, v := range values {
		gw.Write([]byte(v))
		gw.Write([]byte{'\n'})
	}
	// gzip.Writer.Close only flushes to the in-memory buffer; it cannot fail.
	_ = gw.Close()
	return out.Bytes()
}

// parseGzipValues reads a previously-written accumulative artifact (the gzip'd
// wire format produced by gzipValues) back into a value list: gunzip, then
// trim each line and drop blanks and '#' comments. Unlike extractValues it does
// not expect a tarball — it only reads our own emitted artifact.
func parseGzipValues(b []byte) ([]string, error) {
	gr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, errors.New("artifact is not gzip")
	}
	raw, err := io.ReadAll(gr)
	if err != nil {
		return nil, err
	}
	var values []string
	for _, line := range strings.Split(string(raw), "\n") {
		v := strings.TrimSpace(line)
		if v == "" || strings.HasPrefix(v, "#") {
			continue
		}
		values = append(values, v)
	}
	return values, nil
}
