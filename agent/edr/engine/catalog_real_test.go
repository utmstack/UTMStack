package engine

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestRealOpenVPNPEMatchesSignedCatalogHash(t *testing.T) {
	path := os.Getenv("EDR_OPENVPN_CATALOG_ZIP")
	if path == "" {
		t.Skip("set EDR_OPENVPN_CATALOG_ZIP to OpenVPN tap-windows6 9.27.0 dist.win10.zip")
	}
	st, err := os.Stat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() != 189720 {
		t.Fatalf("unexpected real-catalog archive: %v", err)
	}
	archive, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != "36e2609b7ceefedcb978ce5c48caf9e0e5af83423717c4e2e3c1d7ebca8f62a5" {
		t.Fatalf("archive SHA-256 changed: %s", got)
	}
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatal(err)
	}
	var pe []byte
	for _, f := range zr.File {
		if f.Name != "dist.win10/amd64/tap0901.sys" {
			continue
		}
		if f.UncompressedSize64 != 41000 {
			t.Fatal("driver size changed")
		}
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		pe, err = io.ReadAll(io.LimitReader(r, 41001))
		closeErr := r.Close()
		if err != nil || closeErr != nil || len(pe) != 41000 {
			t.Fatalf("driver read: %v, close: %v", err, closeErr)
		}
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(pe)); got != "581dcaace05d5c1ac9512457ff50565aca5d904d2c209bd3fc369ca4d4a0d2b1" {
		t.Fatalf("driver SHA-256 changed: %s", got)
	}
	digest, err := peCatalogDigest(pe, crypto.SHA256)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(digest); got != "9e176400e823172804afb95b5eb0c03c39b8e96a6b15fab745fa9190da45ebbb" {
		t.Fatalf("PE image hash %s does not match signed catalog member", got)
	}
	changed := bytes.Clone(pe)
	changed[0x200] ^= 1
	if digest, err := peCatalogDigest(changed, crypto.SHA256); err == nil && hex.EncodeToString(digest) == "9e176400e823172804afb95b5eb0c03c39b8e96a6b15fab745fa9190da45ebbb" {
		t.Fatal("changed driver kept its catalog hash")
	}
}
