// feedstage acquires public ThreatWinds hash snapshots for feedpack.
package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/feed"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

const publicBase = "https://apis.threatwinds.com/api/feeds/v1"
const maxMetadata = 2 << 20
const maxArtifact = 256 << 20
const maxTotal = 512 << 20

type inputSpec struct {
	Version                   string         `json:"version"`
	Levels                    []string       `json:"levels"`
	Files                     []officialFile `json:"files"`
	RequireDetachedSignatures bool           `json:"require_detached_signatures,omitempty"`
}

type officialFile struct {
	Name              string `json:"name"`
	Path              string `json:"path"`
	Official          bool   `json:"official"`
	DetachedSignature string `json:"detached_signature,omitempty"`
}

type catalogItem struct {
	Name     string `json:"name"`
	Accuracy string `json:"accuracy"`
	Type     string `json:"type"`
}

type outputSpec struct {
	Version string       `json:"version"`
	Files   []outputFile `json:"files"`
}

type outputFile struct {
	Name                  string `json:"name"`
	Path                  string `json:"path"`
	Format                string `json:"format,omitempty"`
	Official              bool   `json:"official,omitempty"`
	DetachedSignature     string `json:"detached_signature,omitempty"`
	DedupeAgainstOfficial bool   `json:"dedupe_against_official,omitempty"`
}

type receipt struct {
	Source       string    `json:"source"`
	FetchedAt    time.Time `json:"fetched_at"`
	CatalogSHA   string    `json:"catalog_sha256"`
	SumsSHA      string    `json:"checksum_manifest_sha256"`
	SumsAfterSHA string    `json:"checksum_manifest_after_sha256"`
	Official     []record  `json:"official"`
	Files        []record  `json:"files"`
}

type record struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	Bytes           int    `json:"bytes"`
	SHA256          string `json:"sha256"`
	MD5             string `json:"publisher_md5"`
	SignatureBytes  int    `json:"detached_signature_bytes,omitempty"`
	SignatureSHA256 string `json:"detached_signature_sha256,omitempty"`
	Omitted         bool   `json:"omitted_empty,omitempty"`
}

var validLevel = regexp.MustCompile(`^level[123]$`)
var validChecksumName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*_level[123]$`)
var validOfficial = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

func main() {
	input := flag.String("spec", "", "JSON with version, official files, and optional accuracy levels")
	output := flag.String("out", "", "new private staging directory")
	flag.Parse()
	if *input == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "feedstage: -spec and -out are required")
		os.Exit(2)
	}
	client := &http.Client{Transport: &http.Transport{ResponseHeaderTimeout: 30 * time.Second},
		CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("redirects disabled") }}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	if err := stage(ctx, client, publicBase, *input, *output, authFromEnv()); err != nil {
		fmt.Fprintln(os.Stderr, "feedstage:", err)
		os.Exit(1)
	}
	fmt.Println("Verified feedpack inputs staged in", *output)
}

type credentials struct{ token, key, secret string }

func authFromEnv() credentials {
	return credentials{os.Getenv("THREATWINDS_TOKEN"), os.Getenv("THREATWINDS_API_KEY"), os.Getenv("THREATWINDS_API_SECRET")}
}

func get(ctx context.Context, client *http.Client, url string, limit int64, auth credentials) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("User-Agent", "OpenEDR-feedstage/1")
	if auth.token != "" {
		req.Header.Set("Authorization", "Bearer "+auth.token)
	} else if auth.key != "" && auth.secret != "" {
		req.Header.Set("api-key", auth.key)
		req.Header.Set("api-secret", auth.secret)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("feed request failed") // Never expose credential-bearing URLs or headers.
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed HTTP status %d", resp.StatusCode)
	}
	if resp.ContentLength > limit {
		return nil, errors.New("feed exceeds size limit")
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, errors.New("feed body read failed")
	}
	if int64(len(b)) > limit {
		return nil, errors.New("feed exceeds size limit")
	}
	return b, nil
}

func decodeInput(path string) (inputSpec, error) {
	var in inputSpec
	st, err := os.Stat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() > maxMetadata {
		return in, errors.New("input spec must be a bounded regular file")
	}
	b, err := os.ReadFile(path)
	if err != nil || len(b) > maxMetadata {
		return in, errors.New("unable to read bounded input spec")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		return in, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return in, errors.New("trailing input JSON")
	}
	if strings.TrimSpace(in.Version) == "" || len(in.Version) > 128 || len(in.Files) == 0 || len(in.Files) > 16 {
		return in, errors.New("version and 1–16 official files required")
	}
	if len(in.Levels) == 0 {
		in.Levels = []string{"level1", "level2", "level3"}
	}
	seen := make(map[string]bool)
	for _, level := range in.Levels {
		if !validLevel.MatchString(level) || seen[level] {
			return in, errors.New("levels must be unique level1, level2 or level3 values")
		}
		seen[level] = true
	}
	seen = make(map[string]bool)
	for i := range in.Files {
		f := &in.Files[i]
		if !f.Official || !validOfficial.MatchString(f.Name) || f.Path == "" || seen[strings.ToLower(f.Name)] {
			return in, errors.New("official files need unique portable names, paths and official=true")
		}
		isCVD := strings.EqualFold(filepath.Ext(f.Name), ".cvd")
		if f.DetachedSignature != "" && !isCVD || in.RequireDetachedSignatures && isCVD && f.DetachedSignature == "" {
			return in, fmt.Errorf("official CVD %s requires a detached signature", f.Name)
		}
		seen[strings.ToLower(f.Name)] = true
		if !filepath.IsAbs(f.Path) {
			f.Path = filepath.Join(filepath.Dir(path), f.Path)
		}
		absolute, err := filepath.Abs(f.Path)
		if err != nil {
			return in, err
		}
		f.Path = absolute
		st, err := os.Stat(f.Path)
		if err != nil || !st.Mode().IsRegular() || st.Size() > maxArtifact {
			return in, fmt.Errorf("official input %s is not a bounded regular file", f.Name)
		}
		if f.DetachedSignature != "" {
			if !filepath.IsAbs(f.DetachedSignature) {
				f.DetachedSignature = filepath.Join(filepath.Dir(path), f.DetachedSignature)
			}
			f.DetachedSignature, err = filepath.Abs(f.DetachedSignature)
			if err != nil {
				return in, err
			}
			st, err = os.Stat(f.DetachedSignature)
			if err != nil || !st.Mode().IsRegular() || st.Size() > 1<<20 {
				return in, fmt.Errorf("detached signature for %s is not a bounded regular file", f.Name)
			}
		}
	}
	return in, nil
}

func parseSums(data []byte) (map[string]string, error) {
	out := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 || len(fields[0]) != 32 || !strings.HasSuffix(fields[1], ".list.tar.gz") || !validChecksumName.MatchString(strings.TrimSuffix(fields[1], ".list.tar.gz")) || out[fields[1]] != "" {
			return nil, errors.New("invalid or repeated checksum entry")
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			return nil, errors.New("invalid checksum hex")
		}
		out[fields[1]] = strings.ToLower(fields[0])
	}
	if len(out) == 0 {
		return nil, errors.New("empty checksum manifest")
	}
	return out, nil
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}

func sum256(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func copyOfficial(source, dest string, limit int64) (record, error) {
	in, err := os.Open(source)
	if err != nil {
		return record{}, err
	}
	defer in.Close()
	st, err := in.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Size() > limit {
		return record{}, errors.New("official input is not a bounded regular file")
	}
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return record{}, err
	}
	h := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(out, h), io.LimitReader(in, limit+1))
	closeErr := out.Close()
	if copyErr != nil {
		return record{}, copyErr
	}
	if closeErr != nil {
		return record{}, closeErr
	}
	if n > limit || n != st.Size() {
		return record{}, errors.New("official input changed or exceeds size limit")
	}
	return record{Bytes: int(n), SHA256: hex.EncodeToString(h.Sum(nil))}, nil
}

func stage(ctx context.Context, client *http.Client, base, input, output string, auth credentials) error {
	return stageWithVerifier(ctx, client, base, input, output, auth, signatures.VerifyOfficialDetachedCVD)
}

func stageWithVerifier(ctx context.Context, client *http.Client, base, input, output string, auth credentials, verifyDetached func([]byte, []byte) error) error {
	if auth.token == "" && (auth.key == "") != (auth.secret == "") {
		return errors.New("api-key and api-secret must be configured together")
	}
	in, err := decodeInput(input)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(output); err == nil {
		return errors.New("output already exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	parent := filepath.Dir(output)
	temp, err := os.MkdirTemp(parent, ".feedstage-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	if err := os.Chmod(temp, 0700); err != nil {
		return err
	}
	catalogData, err := get(ctx, client, base+"/list", maxMetadata, auth)
	if err != nil {
		return fmt.Errorf("catalog: %w", err)
	}
	var catalog []catalogItem
	if err := json.Unmarshal(catalogData, &catalog); err != nil {
		return fmt.Errorf("catalog: %w", err)
	}
	if len(catalog) == 0 || len(catalog) > 1000 {
		return errors.New("catalog count out of bounds")
	}
	allowed := map[string]bool{"md5": true, "sha1": true, "sha256": true}
	levels := make(map[string]bool)
	for _, l := range in.Levels {
		levels[l] = true
	}
	selected := make(map[string]catalogItem)
	for _, item := range catalog {
		if allowed[item.Name] && levels[item.Accuracy] && item.Type == "accumulative" {
			filename := item.Name + "_" + item.Accuracy + ".list.tar.gz"
			if selected[filename].Name != "" {
				return errors.New("duplicate selected catalog entry")
			}
			selected[filename] = item
		}
	}
	if len(selected) == 0 {
		return errors.New("no advertised hash snapshots at requested levels")
	}
	beforeData, err := get(ctx, client, base+"/download/checksum", maxMetadata, auth)
	if err != nil {
		return fmt.Errorf("checksum manifest: %w", err)
	}
	before, err := parseSums(beforeData)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(selected))
	for name := range selected {
		if before[name] == "" {
			return fmt.Errorf("selected feed %s missing from checksum manifest", name)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	if err := os.Mkdir(filepath.Join(temp, "artifacts"), 0700); err != nil {
		return err
	}
	out := outputSpec{Version: in.Version}
	rec := receipt{Source: base, FetchedAt: time.Now().UTC(), CatalogSHA: sum256(catalogData), SumsSHA: sum256(beforeData)}
	total := int64(0)
	downloaded := int64(0)
	for _, f := range in.Files {
		path := filepath.Join("artifacts", "official-"+f.Name)
		entry, err := copyOfficial(f.Path, filepath.Join(temp, path), maxArtifact)
		if err != nil {
			return fmt.Errorf("copying official %s: %w", f.Name, err)
		}
		var signaturePath string
		if f.DetachedSignature != "" {
			signaturePath = path + ".sign"
			signatureRecord, err := copyOfficial(f.DetachedSignature, filepath.Join(temp, signaturePath), 1<<20)
			if err != nil {
				return fmt.Errorf("copying detached signature for %s: %w", f.Name, err)
			}
			cvd, err := os.ReadFile(filepath.Join(temp, path))
			if err != nil {
				return err
			}
			sign, err := os.ReadFile(filepath.Join(temp, signaturePath))
			if err != nil {
				return err
			}
			if err := verifyDetached(cvd, sign); err != nil {
				return fmt.Errorf("verifying detached signature for %s: %w", f.Name, err)
			}
			entry.SignatureBytes = signatureRecord.Bytes
			entry.SignatureSHA256 = signatureRecord.SHA256
		}
		total += int64(entry.Bytes)
		if total > maxTotal {
			return errors.New("feed generation exceeds size limit")
		}
		entry.Name = f.Name
		rec.Official = append(rec.Official, entry)
		out.Files = append(out.Files, outputFile{Name: f.Name, Path: path, Official: true, DetachedSignature: signaturePath})
	}
	for _, name := range names {
		item := selected[name]
		url := base + "/download/list/" + item.Accuracy + "/accumulative/" + item.Name
		data, err := get(ctx, client, url, maxArtifact, auth)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		downloaded += int64(len(data))
		if downloaded > maxTotal {
			return errors.New("downloaded feeds exceed size limit")
		}
		md := md5.Sum(data)
		mdHex := hex.EncodeToString(md[:])
		if mdHex != before[name] {
			return fmt.Errorf("%s: publisher checksum mismatch", name)
		}
		ext := ".hsb"
		if item.Name == "md5" {
			ext = ".hdb"
		}
		outputName := "threatwinds-" + item.Name + "-" + item.Accuracy + ext
		normalized, err := feed.NormalizeArtifact(feed.ManifestFile{Name: outputName, Format: item.Name + "-list.tar.gz"}, data)
		entry := record{Name: name, URL: url, Bytes: len(data), SHA256: sum256(data), MD5: mdHex}
		if errors.Is(err, signatures.ErrEmptyHashList) {
			entry.Omitted = true
			rec.Files = append(rec.Files, entry)
			continue
		}
		if err != nil {
			return fmt.Errorf("%s: invalid hash archive: %w", name, err)
		}
		total += int64(max(len(data), len(normalized)))
		if total > maxTotal {
			return errors.New("feed generation exceeds size limit")
		}
		path := filepath.Join("artifacts", name)
		if err := os.WriteFile(filepath.Join(temp, path), data, 0600); err != nil {
			return err
		}
		out.Files = append(out.Files, outputFile{Name: outputName, Path: path, Format: item.Name + "-list.tar.gz", DedupeAgainstOfficial: true})
		rec.Files = append(rec.Files, entry)
	}
	afterData, err := get(ctx, client, base+"/download/checksum", maxMetadata, auth)
	if err != nil {
		return fmt.Errorf("rechecking checksum manifest: %w", err)
	}
	after, err := parseSums(afterData)
	if err != nil {
		return err
	}
	rec.SumsAfterSHA = sum256(afterData)
	for _, name := range names {
		if after[name] != before[name] {
			return fmt.Errorf("%s: publisher snapshot changed during acquisition", name)
		}
	}
	if err := writeJSON(filepath.Join(temp, "feedpack-spec.json"), out); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(temp, "receipt.json"), rec); err != nil {
		return err
	}
	if err := os.Rename(temp, output); err != nil {
		return err
	}
	return nil
}
