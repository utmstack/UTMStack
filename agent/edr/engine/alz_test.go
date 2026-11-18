package engine

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

const alzOracle = "../../../docs/edr/native-engine/audits/2026-10-28-v142-alz/"

func TestALZReferenceChildHashes(t *testing.T) {
	b, err := os.ReadFile(alzOracle + "reference-verdicts.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Fixture    string `json:"fixture"`
		ArchiveSHA string `json:"archive_sha256"`
		ChildMD5   string `json:"child_md5"`
		ChildSize  int    `json:"child_size"`
	}
	if err := json.Unmarshal(b, &cases); err != nil || len(cases) != 8 {
		t.Fatalf("reference receipt: %d cases, %v", len(cases), err)
	}
	for _, tc := range cases {
		t.Run(tc.Fixture, func(t *testing.T) {
			archive, err := os.ReadFile(alzOracle + "oracle/" + tc.Fixture)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != tc.ArchiveSHA {
				t.Fatalf("ALZ oracle changed: %s", got)
			}
			rule := map[string]string{"child.hdb": fmt.Sprintf("%s:%d:Reference.ALZ.Child\n", tc.ChildMD5, tc.ChildSize)}
			expectDetected(t, pipelineState(t, rule), archive, "Reference.ALZ.Child")
			clean := pipelineState(t, map[string]string{"other.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"})
			if strings.Contains(tc.Fixture, "before") {
				expectIncomplete(t, clean, archive, ErrUnsupported)
			} else {
				expectScan(t, clean, archive, "", "")
			}
		})
	}
}

func TestALZMalformedAndLimited(t *testing.T) {
	archive, err := os.ReadFile(alzOracle + "oracle/stored.alz")
	if err != nil {
		t.Fatal(err)
	}
	clean := map[string]string{"other.hdb": "9a2f72a9a15b29e841489be539214a7e:5:Other\n"}
	truncated := append([]byte(nil), archive...)
	truncated[27] = 250 // compressed member reaches past the archive's EOF
	expectIncomplete(t, pipelineState(t, clean), truncated, ErrMalformed)
	limited := pipelineState(t, clean)
	limited.maxFile = 50
	expectIncomplete(t, limited, archive, ErrLimit)
	if _, err := alzMemberOf(archive, len(archive)-2); !errors.Is(err, ErrMalformed) {
		t.Fatalf("truncated local header: %v", err)
	}
}

func TestALZContainerMetadata(t *testing.T) {
	archive, err := os.ReadFile(alzOracle + "oracle/encrypted-before-stored.alz")
	if err != nil {
		t.Fatal(err)
	}
	rules := map[string]string{"alz.cdb": "ALZ.Secret:CL_TYPE_ALZ:*:^secret\\.txt$:6:6:1:1:*:\n"}
	expectDetected(t, pipelineState(t, rules), archive, "ALZ.Secret")
	rules = map[string]string{"alz.cdb": "ALZ.Child:CL_TYPE_ALZ:*:^child\\.txt$:102:102:0:2:*:\n"}
	expectDetected(t, pipelineState(t, rules), archive, "ALZ.Child")
}
