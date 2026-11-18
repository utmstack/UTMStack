package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// rootPreclassMetadata builds the fields ClamAV initializes or proves for a
// root fmap before BC_PRECLASS. The map remains partial: only the supported
// Office/HWPML parser properties are included; child objects, indicators and
// optional extra hashes are not synthesized. A missing query remains undecided.
func rootPreclassMetadata(ctx context.Context, data []byte, typ signatures.FileType) (map[string]any, error) {
	return rootPreclassMetadataForPath(ctx, data, typ, "")
}

// rootPreclassMetadataForPath adds the root name and path ClamAV records for
// file scans. Byte-buffer callers pass an empty path and retain the partial
// metadata tree used before path-aware scanning was introduced.
func rootPreclassMetadataForPath(ctx context.Context, data []byte, typ signatures.FileType, path string) (map[string]any, error) {
	meta, err := initialRootMetadata(ctx, data, typ, path)
	if err != nil {
		return nil, err
	}
	return finishRootMetadata(ctx, data, typ, meta)
}

// initialRootMetadata records the root identity and hash before any parser or
// raw rule runs. JSON handles in every child and derived view refer to this
// same root; RootFileType and parser properties arrive only after analysis.
func initialRootMetadata(ctx context.Context, data []byte, typ signatures.FileType, path string) (map[string]any, error) {
	meta := map[string]any{
		"Magic":    "CLAMJSONv0",
		"FileSize": uint64(len(data)),
		"ObjectID": uint64(0),
	}
	if path != "" {
		name := nativeMetadataBase(path)
		if _, nameOK := jsonCString(name); !nameOK {
			return nil, unknownRepresentation("metadata-json")
		}
		if _, pathOK := jsonCString(path); !pathOK {
			return nil, unknownRepresentation("metadata-json")
		}
		meta["FileName"] = name
		meta["FilePath"] = path
	}
	if typ != signatures.TypeError && typ != signatures.TypeOther {
		if name := typ.String(); name != "" {
			meta["FileType"] = "CL_TYPE_" + name
		}
	}
	h := sha256.New()
	for start := 0; start < len(data); start += 1 << 20 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		end := min(start+1<<20, len(data))
		_, _ = h.Write(data[start:end])
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	meta["sha2-256"] = hex.EncodeToString(h.Sum(nil))
	return meta, nil
}

func finishRootMetadata(ctx context.Context, data []byte, typ signatures.FileType, meta map[string]any) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if meta == nil {
		return nil, unknownRepresentation("metadata-json")
	}
	if name, ok := meta["FileType"].(string); ok {
		meta["RootFileType"] = name
	}
	// The reference's document parser has already run before PRECLASS. Add
	// only proved property subtrees; MetadataPartial stays true at the call
	// site because extracted children and other fields remain unmodeled.
	var properties map[string]any
	var err error
	switch typ {
	case signatures.TypeXMLHWP:
		properties, err = hwpMLProperties(ctx, data, int64(len(data))+(1<<20))
	case signatures.TypeXMLWord, signatures.TypeXMLXL:
		properties, err = msxmlProperties(ctx, data, int64(len(data))+(1<<20), "CL_TYPE_"+typ.String())
	}
	if err != nil {
		// A parser failure leaves the already proved root fields valid.
		// Return that shared partial tree so independent PRECLASS hooks can
		// still detect; no unproved parser subtree is published.
		return meta, err
	}
	for key, value := range properties {
		switch key {
		case "HWPML", "WordDocument", "Workbook", "DocumentProperties", "BinaryData":
			meta[key] = value
		}
	}
	return meta, nil
}

// plainScalarMetadataJSON is the json-c PRETTY byte layout for a root binary
// or plain ASCII text object with no extracted children or parser properties.
// The insertion order follows scan_common, change_type, SHA-256 and the final
// RootFileType alias. Other root types can add parser fields or nested objects
// before the deprecated target-13 pass and remain incomplete.
func plainScalarMetadataJSON(meta map[string]any) []byte {
	size, sizeOK := meta["FileSize"].(uint64)
	fileType, typeOK := meta["FileType"].(string)
	hash, hashOK := meta["sha2-256"].(string)
	if !sizeOK || !typeOK || !hashOK || meta["Magic"] != "CLAMJSONv0" || meta["ObjectID"] != uint64(0) ||
		(fileType != "CL_TYPE_BINARY_DATA" && fileType != "CL_TYPE_TEXT_ASCII") || meta["RootFileType"] != fileType {
		return nil
	}
	for key := range meta {
		switch key {
		case "Magic", "FileName", "FilePath", "FileSize", "ObjectID", "FileType", "sha2-256", "RootFileType":
		default:
			return nil
		}
	}
	b := make([]byte, 0, 160)
	b = append(b, "{\n  \"Magic\":\"CLAMJSONv0\","...)
	if name, hasName := meta["FileName"]; hasName {
		fileName, nameOK := name.(string)
		filePath, pathOK := meta["FilePath"].(string)
		if !nameOK || !pathOK || fileName == "" || filePath == "" || nativeMetadataBase(filePath) != fileName {
			return nil
		}
		// Match json-c's path-string escapes, including raw UTF-8 for the
		// U+2028/U+2029 separators that encoding/json escapes by default.
		nameJSON, nameOK := jsonCString(fileName)
		pathJSON, pathOK := jsonCString(filePath)
		if !nameOK || !pathOK {
			return nil
		}
		b = append(b, "\n  \"FileName\":"...)
		b = append(b, nameJSON...)
		b = append(b, ",\n  \"FilePath\":"...)
		b = append(b, pathJSON...)
		b = append(b, ',')
	} else if _, hasPath := meta["FilePath"]; hasPath {
		return nil
	}
	b = append(b, "\n  \"FileSize\":"...)
	b = strconv.AppendUint(b, size, 10)
	b = append(b, ",\n  \"ObjectID\":0,\n  \"FileType\":\""...)
	b = append(b, fileType...)
	b = append(b, "\",\n  \"sha2-256\":\""...)
	b = append(b, hash...)
	b = append(b, "\",\n  \"RootFileType\":\""...)
	b = append(b, fileType...)
	b = append(b, "\"\n}"...)
	return b
}

func jsonCString(s string) ([]byte, bool) {
	if !utf8.ValidString(s) || strings.IndexByte(s, 0) >= 0 {
		return nil, false
	}
	out := make([]byte, 0, len(s)+2)
	out = append(out, '"')
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"', '\\':
			out = append(out, '\\', s[i])
		case '\b':
			out = append(out, '\\', 'b')
		case '\f':
			out = append(out, '\\', 'f')
		case '\n':
			out = append(out, '\\', 'n')
		case '\r':
			out = append(out, '\\', 'r')
		case '\t':
			out = append(out, '\\', 't')
		default:
			if s[i] < 0x20 {
				const hexDigits = "0123456789abcdef"
				out = append(out, '\\', 'u', '0', '0', hexDigits[s[i]>>4], hexDigits[s[i]&0x0f])
			} else {
				out = append(out, s[i])
			}
		}
	}
	return append(out, '"'), true
}

// ClamAV uses both path separators to derive FileName, even on Unix where a
// backslash may be a literal on-disk character. FilePath retains the input.
func nativeMetadataBase(path string) string {
	if i := strings.LastIndexAny(path, "/\\"); i >= 0 {
		return path[i+1:]
	}
	return path
}
