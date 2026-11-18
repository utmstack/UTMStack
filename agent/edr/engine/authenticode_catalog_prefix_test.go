package engine

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Catalog parsing has no scan context in the reference. A blocked certificate
// prevents member-hash admission and chained promotions, but does not stop the
// direct-approval pass or undo independently approved certificates in that bag.
func TestExportAuthenticodeBlockedCatalogPrefix(t *testing.T) {
	if os.Getenv("EDR_CATALOG_PREFIX_EXPORT") != "1" || os.Getenv("EDR_HISTORICAL_CHAIN_EXPORT") == "" {
		t.Skip("set an owned EDR_HISTORICAL_CHAIN_EXPORT and EDR_CATALOG_PREFIX_EXPORT=1 to regenerate public signed controls")
	}
	orders := [][]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		chain := historicalPEChainFixture(t, hash, 0, true, x509.ExtKeyUsageCodeSigning)
		blocked := historicalPEChainFixture(t, hash, 0, true, x509.ExtKeyUsageCodeSigning)
		noise := historicalPEChainFixture(t, hash, 0, true, x509.ExtKeyUsageCodeSigning)
		member := simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
		memberDigest, err := peCatalogDigest(member, crypto.SHA256)
		if err != nil {
			t.Fatal(err)
		}
		withoutSigner, _ := signedPEWithSigner(t, chain.leaf, chain.leafKey, noise.root.Raw)
		for _, direct := range []bool{false, true} {
			trust := chain.crb
			if direct {
				trust = replaceHistoricalCRBField(t, policyBlockRow(t, chain.leaf, "Trusted.Direct", "wildcard"), 1, "1")
			}
			for _, blockKind := range []string{"exact", "wildcard", "other-key"} {
				for order, indexes := range orders {
					parts := [][]byte{chain.root.Raw, chain.leaf.Raw, blocked.leaf.Raw}
					bag := bytes.Join([][]byte{parts[indexes[0]], parts[indexes[1]], parts[indexes[2]]}, nil)
					for _, tamper := range []bool{false, true} {
						catalog := signedCatalogWithSignerForEngine(t, memberDigest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, bag)
						if tamper {
							catalog[len(catalog)-1] ^= 1
						}
						for _, surface := range []string{"learned-signer", "catalog-member"} {
							label := fmt.Sprintf("blocked-prefix-%s-direct%v", hash, direct)
							kind := fmt.Sprintf("%s-order%d-tamper%v-%s", blockKind, order, tamper, surface)
							t.Run(label+"/"+kind, func(t *testing.T) {
								inputs := map[string][]byte{
									"trust.crb":  []byte(trust + policyBlockRow(t, blocked.leaf, "Blocked.Unrelated", blockKind)),
									"raw.ndb":    []byte("Raw.Marker:0:*:9090909090\n"),
									"prefix.cat": catalog,
								}
								body := withoutSigner
								want := direct || blockKind == "other-key"
								if surface == "catalog-member" {
									body = member
									want = blockKind == "other-key" && !tamper
								}
								db, err := signatures.Compile(inputs)
								if err != nil {
									t.Fatal(err)
								}
								e := New(config.EDRConfig{})
								e.current.Store(&generation{db: db})
								clean, name, scanErr := e.Scan(context.Background(), body)
								exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
								if scanErr != nil || clean != want || !clean && name != "Raw.Marker" {
									t.Errorf("clean=%v name=%q err=%v want=%v", clean, name, scanErr, want)
								}
							})
						}
					}
				}
			}
		}
	}
}

func TestAuthenticodeCatalogPrefixReference(t *testing.T) {
	file, err := os.Open("testdata/authenticode_catalog_prefix.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	count := 0
	for {
		var row struct {
			Case      string            `json:"case"`
			Sample    []byte            `json:"sample"`
			Databases map[string][]byte `json:"databases"`
			Clean     bool              `json:"clean"`
		}
		if err := decoder.Decode(&row); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		count++
		t.Run(row.Case, func(t *testing.T) {
			db, err := signatures.Compile(row.Databases)
			if err != nil {
				t.Fatal(err)
			}
			e := New(config.EDRConfig{})
			e.current.Store(&generation{db: db})
			clean, name, scanErr := e.Scan(context.Background(), row.Sample)
			if scanErr != nil || clean != row.Clean || !clean && name != "Raw.Marker" {
				t.Errorf("fixed signed reference control: clean=%v signature=%q error=%v want=%v", clean, name, scanErr, row.Clean)
			}
		})
	}
	if count != 288 {
		t.Fatalf("fixed signed control count=%d want=288", count)
	}
}
