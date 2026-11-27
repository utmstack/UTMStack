package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

var testLoggerOnce sync.Once

func ensureTestLogger() {
	testLoggerOnce.Do(func() {
		utils.InitLogger(os.DevNull)
	})
}

func TestUserDataTypeKinds(t *testing.T) {
	ensureTestLogger()

	cases := []struct {
		name    string
		catalog string // "" means the file must not exist
		want    map[string]map[string]bool
		wantErr bool
	}{
		{"missing catalog is an empty set, no error", "", map[string]map[string]bool{}, false},
		{"populated catalog groups by kind", `{"user_datatypes":{"firewall-barracuda":{"kind":"syslog","udp":"7100","tcp":"7101"},"my-user-file":{"kind":"file"}}}`,
			map[string]map[string]bool{"syslog": {"firewall-barracuda": true}, "file": {"my-user-file": true}}, false},
		{"corrupt catalog must error", `{not-json`, nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			orig := CatalogFileName
			CatalogFileName = filepath.Join(tmp, "log-collector-catalog.json")
			t.Cleanup(func() { CatalogFileName = orig })

			if tc.catalog != "" {
				if err := os.WriteFile(CatalogFileName, []byte(tc.catalog), 0644); err != nil {
					t.Fatal(err)
				}
			}

			got, err := UserDataTypeKinds()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("UserDataTypeKinds() = %v, want error for corrupt catalog", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("UserDataTypeKinds() error = %v, want none", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("UserDataTypeKinds() = %v, want %v", got, tc.want)
			}
			for kind, names := range tc.want {
				if len(got[kind]) != len(names) {
					t.Fatalf("UserDataTypeKinds()[%q] = %v, want %v", kind, got[kind], names)
				}
				for name := range names {
					if !got[kind][name] {
						t.Fatalf("UserDataTypeKinds()[%q] = %v, want %q present", kind, got[kind], name)
					}
				}
			}
		})
	}
}
