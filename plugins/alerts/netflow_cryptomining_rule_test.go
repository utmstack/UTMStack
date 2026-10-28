package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"gopkg.in/yaml.v3"
)

// The mining rule was not covered by any contract manifest, so this test
// compiles its `where` against the real CEL cache and drives it with a real
// FP sample and a real mining sample (#2985 fix):
//   - an established high-volume session to a mining port MUST match;
//   - the observed FP (SYN+ACK one-way flow, 132 bytes, no actionResult)
//     MUST NOT match.
func TestNetflowCryptominingRule(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "rules", "netflow", "netflow_cryptomining_traffic.yml"))
	if err != nil {
		t.Fatalf("read rule: %v", err)
	}
	var doc struct {
		Where string `yaml:"where"`
	}
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatalf("parse rule yaml: %v", err)
	}
	where := doc.Where
	if strings.TrimSpace(where) == "" {
		t.Fatal("empty where in rule")
	}

	cache := plugins.NewCELCache("netflow-crypto-mining-rule-test")
	eval := func(sample string) (bool, error) {
		got, err := cache.Eval(where, sample)
		if err != nil {
			// A direct-selector eval can fail on an incomplete sample after a
			// successful compile; that is not a syntax error (same convention
			// as filter_contract_test.go).
			if strings.Contains(err.Error(), "failed to evaluate program") {
				return false, nil
			}
			return false, err
		}
		return got, nil
	}

	cases := []struct {
		name   string
		sample string
		want   bool
	}{
		{
			name:   "established stratum session with sustained data matches",
			sample: `{"origin":{"ip":"10.1.2.3"},"target":{"ip":"185.69.126.194","port":7777},"protocol":"TCP","actionResult":"success","event":{"totalBytes":524288}}`,
			want:   true,
		},
		{
			name:   "observed FP: one-way SYN+ACK probe, 132 bytes, no actionResult does NOT match",
			sample: `{"origin":{"ip":"57.135.227.31"},"target":{"ip":"185.69.126.194","port":7777},"protocol":"TCP","event":{"totalBytes":132}}`,
			want:   false,
		},
		{
			name:   "low-volume blip on a mining port does NOT match",
			sample: `{"origin":{"ip":"10.1.2.3"},"target":{"ip":"83.220.177.229","port":4444},"protocol":"TCP","actionResult":"success","event":{"totalBytes":120}}`,
			want:   false,
		},
		{
			name:   "internal target (192.168.0.0/16) is excluded",
			sample: `{"origin":{"ip":"10.1.2.3"},"target":{"ip":"192.168.1.50","port":3333},"protocol":"TCP","actionResult":"success","event":{"totalBytes":999999}}`,
			want:   false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := eval(c.sample)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			if got != c.want {
				t.Fatalf("where = %v, want %v", got, c.want)
			}
		})
	}
}
