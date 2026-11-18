package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func archive(t *testing.T, name, content string) []byte {
	t.Helper()
	var b bytes.Buffer
	z := gzip.NewWriter(&b)
	tw := tar.NewWriter(z)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func input(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.cvd"), []byte("placeholder"), 0600); err != nil {
		t.Fatal(err)
	}
	spec := `{"version":"test","levels":["level1"],"files":[{"name":"main.cvd","path":"main.cvd","official":true}]}`
	path := filepath.Join(dir, "input.json")
	if err := os.WriteFile(path, []byte(spec), 0600); err != nil {
		t.Fatal(err)
	}
	return path, filepath.Join(dir, "staged")
}

func TestStageSnapshotAndRejectRotation(t *testing.T) {
	const digest = "44d88612fea8a8f36de82e1278abb02f"
	good := archive(t, "md5_level1.list", digest+"\n")
	h := md5.Sum(good)
	checksum := hex.EncodeToString(h[:])
	for _, tc := range []struct {
		name, after, body string
		wantError         string
	}{
		{name: "stable"},
		{name: "rotated checksum", after: strings.Repeat("f", 32), wantError: "changed during acquisition"},
		{name: "corrupt body", body: "corrupt", wantError: "publisher checksum mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in, out := input(t)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/list":
					fmt.Fprint(w, `[{"name":"md5","accuracy":"level1","type":"accumulative"},{"name":"url","accuracy":"level1","type":"accumulative"}]`)
				case "/download/checksum":
					value := checksum
					if tc.after != "" && r.Header.Get("X-Second") == "true" {
						value = tc.after
					}
					fmt.Fprintf(w, "%s  md5_level1.list.tar.gz\n", value)
				case "/download/list/level1/accumulative/md5":
					if tc.body != "" {
						_, _ = w.Write([]byte(tc.body))
					} else {
						_, _ = w.Write(good)
					}
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			// The checksum endpoint is fetched twice; rotate on the second call.
			calls := 0
			client := server.Client()
			client.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/download/checksum" {
					calls++
					if calls == 2 {
						r.Header.Set("X-Second", "true")
					}
				}
				return http.DefaultTransport.RoundTrip(r)
			})
			err := stage(context.Background(), client, server.URL, in, out, credentials{})
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("error = %v; want %q", err, tc.wantError)
				}
				if _, statErr := os.Stat(out); !os.IsNotExist(statErr) {
					t.Fatalf("failed stage published output: %v", statErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var spec outputSpec
			b, err := os.ReadFile(filepath.Join(out, "feedpack-spec.json"))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(b, &spec); err != nil {
				t.Fatal(err)
			}
			if len(spec.Files) != 2 || !spec.Files[0].Official || !spec.Files[1].DedupeAgainstOfficial || spec.Files[1].Format != "md5-list.tar.gz" {
				t.Fatalf("unexpected feedpack spec: %+v", spec)
			}
			official, err := os.ReadFile(filepath.Join(out, spec.Files[0].Path))
			if err != nil || string(official) != "placeholder" {
				t.Fatalf("official snapshot differs: %v", err)
			}
			staged, err := os.ReadFile(filepath.Join(out, spec.Files[1].Path))
			if err != nil || !bytes.Equal(staged, good) {
				t.Fatalf("staged bytes differ: %v", err)
			}
			var receiptData receipt
			b, err = os.ReadFile(filepath.Join(out, "receipt.json"))
			if err != nil || json.Unmarshal(b, &receiptData) != nil || len(receiptData.Official) != 1 || len(receiptData.Files) != 1 {
				t.Fatalf("incomplete receipt: %v", err)
			}
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestStageRejectsMalformedArchive(t *testing.T) {
	in, out := input(t)
	body := []byte("not-gzip")
	h := md5.Sum(body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/list":
			fmt.Fprint(w, `[{"name":"sha256","accuracy":"level1","type":"accumulative"}]`)
		case "/download/checksum":
			fmt.Fprintf(w, "%x  sha256_level1.list.tar.gz\n", h)
		case "/download/list/level1/accumulative/sha256":
			_, _ = w.Write(body)
		}
	}))
	defer server.Close()
	err := stage(context.Background(), server.Client(), server.URL, in, out, credentials{})
	if err == nil || !strings.Contains(err.Error(), "invalid hash archive") {
		t.Fatalf("error = %v", err)
	}
}

func TestStageRecordsEmptySnapshotWithoutPublishingIt(t *testing.T) {
	in, out := input(t)
	body := archive(t, "md5_level1.list", "# no indicators\n")
	h := md5.Sum(body)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/list":
			fmt.Fprint(w, `[{"name":"md5","accuracy":"level1","type":"accumulative"}]`)
		case "/download/checksum":
			fmt.Fprintf(w, "%x  md5_level1.list.tar.gz\n", h)
		case "/download/list/level1/accumulative/md5":
			_, _ = w.Write(body)
		}
	}))
	defer server.Close()
	if err := stage(context.Background(), server.Client(), server.URL, in, out, credentials{}); err != nil {
		t.Fatal(err)
	}
	var spec outputSpec
	b, err := os.ReadFile(filepath.Join(out, "feedpack-spec.json"))
	if err != nil || json.Unmarshal(b, &spec) != nil || len(spec.Files) != 1 {
		t.Fatalf("empty list was published: %v", err)
	}
	var receiptData receipt
	b, err = os.ReadFile(filepath.Join(out, "receipt.json"))
	if err != nil || json.Unmarshal(b, &receiptData) != nil || len(receiptData.Files) != 1 || !receiptData.Files[0].Omitted {
		t.Fatalf("empty list lacks audit record: %v", err)
	}
}

func TestStageCarriesVerifiedDetachedSignatureIntoPackSpec(t *testing.T) {
	in, out := input(t)
	dir := filepath.Dir(in)
	cvd := []byte("ClamAV-VDB:" + strings.Repeat("x", 512))
	sign := []byte("#clamsign-1.0\nfixture\n")
	if err := os.WriteFile(filepath.Join(dir, "main.cvd"), cvd, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.cvd.sign"), sign, 0600); err != nil {
		t.Fatal(err)
	}
	spec := `{"version":"test","levels":["level1"],"require_detached_signatures":true,"files":[{"name":"main.cvd","path":"main.cvd","official":true,"detached_signature":"main.cvd.sign"}]}`
	if err := os.WriteFile(in, []byte(spec), 0600); err != nil {
		t.Fatal(err)
	}
	feedData := archive(t, "md5_level1.list", "44d88612fea8a8f36de82e1278abb02f\n")
	feedMD5 := md5.Sum(feedData)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/list":
			fmt.Fprint(w, `[{"name":"md5","accuracy":"level1","type":"accumulative"}]`)
		case "/download/checksum":
			fmt.Fprintf(w, "%x  md5_level1.list.tar.gz\n", feedMD5)
		case "/download/list/level1/accumulative/md5":
			_, _ = w.Write(feedData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	verified := false
	verifier := func(gotCVD, gotSign []byte) error {
		if !bytes.Equal(gotCVD, cvd) || !bytes.Equal(gotSign, sign) {
			return fmt.Errorf("detached signature input changed")
		}
		verified = true
		return nil
	}
	if err := stageWithVerifier(context.Background(), server.Client(), server.URL, in, out, credentials{}, verifier); err != nil {
		t.Fatal(err)
	}
	if !verified {
		t.Fatal("staged CVD was not verified")
	}
	var packed outputSpec
	b, err := os.ReadFile(filepath.Join(out, "feedpack-spec.json"))
	if err != nil || json.Unmarshal(b, &packed) != nil || len(packed.Files) != 2 {
		t.Fatalf("invalid feedpack spec: %v", err)
	}
	signaturePath := packed.Files[0].DetachedSignature
	if signaturePath != filepath.Join("artifacts", "official-main.cvd.sign") {
		t.Fatalf("detached signature not passed to feedpack: %q", signaturePath)
	}
	b, err = os.ReadFile(filepath.Join(out, signaturePath))
	if err != nil || !bytes.Equal(b, sign) {
		t.Fatalf("staged detached signature differs: %v", err)
	}
	var receiptData receipt
	b, err = os.ReadFile(filepath.Join(out, "receipt.json"))
	if err != nil || json.Unmarshal(b, &receiptData) != nil || len(receiptData.Official) != 1 || receiptData.Official[0].SignatureBytes != len(sign) || receiptData.Official[0].SignatureSHA256 != sum256(sign) {
		t.Fatalf("detached signature missing from receipt: %v", err)
	}
}

func TestStageDetachedSignatureFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		name, spec, want string
	}{
		{"required but missing", `{"version":"test","require_detached_signatures":true,"files":[{"name":"main.cvd","path":"main.cvd","official":true}]}`, "requires a detached signature"},
		{"non CVD", `{"version":"test","files":[{"name":"main.hdb","path":"main.cvd","official":true,"detached_signature":"main.cvd.sign"}]}`, "requires a detached signature"},
		{"missing detached file", `{"version":"test","files":[{"name":"main.cvd","path":"main.cvd","official":true,"detached_signature":"missing.sign"}]}`, "not a bounded regular file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in, _ := input(t)
			if err := os.WriteFile(in, []byte(tc.spec), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := decodeInput(in); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("decodeInput error = %v; want %q", err, tc.want)
			}
		})
	}
	in, out := input(t)
	dir := filepath.Dir(in)
	if err := os.WriteFile(filepath.Join(dir, "main.cvd.sign"), []byte("bad signature"), 0600); err != nil {
		t.Fatal(err)
	}
	spec := `{"version":"test","files":[{"name":"main.cvd","path":"main.cvd","official":true,"detached_signature":"main.cvd.sign"}]}`
	if err := os.WriteFile(in, []byte(spec), 0600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/list":
			fmt.Fprint(w, `[{"name":"md5","accuracy":"level1","type":"accumulative"}]`)
		case "/download/checksum":
			fmt.Fprint(w, "44d88612fea8a8f36de82e1278abb02f  md5_level1.list.tar.gz\n")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	if err := stage(context.Background(), server.Client(), server.URL, in, out, credentials{}); err == nil || !strings.Contains(err.Error(), "verifying detached signature") {
		t.Fatalf("invalid detached signature accepted: %v", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatalf("failed stage published output: %v", err)
	}
}
