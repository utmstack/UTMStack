package feed

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// NormalizeArtifact converts a verified artifact into its named native signature
// database. Publishers use the same function to validate bundles before signing.
func NormalizeArtifact(item ManifestFile, data []byte) ([]byte, error) {
	out, _, err := normalizeArtifact(item, data, false)
	return out, err
}

// Hash list formats are explicit: algorithm-list, algorithm-list.gz, or
// algorithm-list.tar.gz. Archive members are read in memory and never extracted
// to the filesystem. There must be exactly one regular member. Empty-object
// digests are dropped because they cannot match in the reference engine.
func normalizeArtifact(item ManifestFile, data []byte, keepDecoded bool) ([]byte, *signatures.PreparedHashList, error) {
	if item.Format == "" || item.Format == "signature" {
		return data, nil, nil
	}
	format := item.Format
	isTar := strings.HasSuffix(format, ".tar.gz")
	if strings.HasSuffix(format, ".gz") {
		z, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, nil, errors.New("invalid hash-list gzip")
		}
		decoded, err := readBounded(z, maxArtifactBytes)
		_ = z.Close()
		if err != nil {
			return nil, nil, err
		}
		data = decoded
		if isTar {
			format = strings.TrimSuffix(format, ".tar.gz")
		} else {
			format = strings.TrimSuffix(format, ".gz")
		}
	}
	if isTar {
		t := tar.NewReader(bytes.NewReader(data))
		h, err := t.Next()
		if err != nil || h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA || h.Size < 0 || h.Size > maxArtifactBytes {
			return nil, nil, errors.New("hash-list archive requires one bounded regular member")
		}
		data, err = readBounded(t, maxArtifactBytes)
		if err != nil {
			return nil, nil, err
		}
		if _, err = t.Next(); err != io.EOF {
			return nil, nil, errors.New("hash-list archive requires exactly one member")
		}
	}
	var algorithm, ext string
	switch format {
	case "md5-list":
		algorithm, ext = "MD5", ".hdb"
	case "sha1-list":
		algorithm, ext = "SHA1", ".hsb"
	case "sha256-list":
		algorithm, ext = "SHA256", ".hsb"
	default:
		return nil, nil, errors.New("unsupported signature artifact format")
	}
	if path.Ext(item.Name) != ext {
		return nil, nil, fmt.Errorf("hash-list output requires %s extension", ext)
	}
	if keepDecoded {
		return signatures.PrepareHashList(data, algorithm, maxArtifactBytes)
	}
	out, err := signatures.NormalizeHashList(data, algorithm, maxArtifactBytes)
	return out, nil, err
}
