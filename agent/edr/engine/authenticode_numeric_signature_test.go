package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// All four inert PEs have independently authenticated certificate signatures.
// The reference additionally checks the numeric signature length before modular
// exponentiation. S and S+N have identical RSA proofs but straddle that gate.
func TestAuthenticodeCertificateNumericSignatureBoundary(t *testing.T) {
	var rows []struct {
		Label, Mode       string
		Body, Certificate []byte
		IssuerKey         []byte `json:"issuer_key"`
		Inputs            map[string][]byte
		Hash              crypto.Hash
		Clean             bool
		NumericBytes      int `json:"numeric_bytes"`
		KeyBytes          int `json:"key_bytes"`
	}
	data, err := os.ReadFile("testdata/authenticode_numeric_signature.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatal("incomplete independent numeric cohort")
	}
	positive := make(map[string][]byte)
	for _, row := range rows {
		if row.Mode == "plus-modulus" {
			cert, err := x509.ParseCertificate(row.Certificate)
			if err != nil {
				t.Fatal(err)
			}
			positive[row.Label] = cert.Signature
		}
	}
	for _, row := range rows {
		t.Run(row.Label+"/"+row.Mode, func(t *testing.T) {
			cert, err := x509.ParseCertificate(row.Certificate)
			if err != nil {
				t.Fatal(err)
			}
			public, err := x509.ParsePKIXPublicKey(row.IssuerKey)
			if err != nil {
				t.Fatal(err)
			}
			key := public.(*rsa.PublicKey)
			value := new(big.Int).SetBytes(cert.Signature)
			if len(value.Bytes()) != row.NumericBytes || key.Size() != row.KeyBytes {
				t.Fatal("numeric boundary changed")
			}
			if (row.NumericBytes < key.Size()-1) != (row.Mode == "short") {
				t.Fatal("fixture no longer straddles reference length gate")
			}
			if !bytes.Equal(referenceSizedFixtureSignature(key, cert.Signature), positive[row.Label]) {
				t.Fatal("positive fixture helper differs from independently accepted representation")
			}
			value.Mod(value, key.N)
			h := row.Hash.New()
			h.Write(cert.RawTBSCertificate)
			if err := rsa.VerifyPKCS1v15(key, row.Hash, h.Sum(nil), value.FillBytes(make([]byte, key.Size()))); err != nil {
				t.Fatal("fixture lost its actual RSA proof", err)
			}
			db, err := signatures.Compile(row.Inputs)
			if err != nil {
				t.Fatal(err)
			}
			e := New(config.EDRConfig{})
			e.current.Store(&generation{db: db})
			clean, name, err := e.Scan(context.Background(), row.Body)
			if err != nil || clean != row.Clean || !clean && name != "Suspicious.Code" {
				t.Fatalf("numeric certificate: clean=%v name=%q err=%v", clean, name, err)
			}
		})
	}
}
