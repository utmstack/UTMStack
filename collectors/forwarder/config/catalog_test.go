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

func TestUserDataTypeNames(t *testing.T) {
	ensureTestLogger()

	cases := []struct {
		name    string
		catalog string // "" means the file must not exist
		want    map[string]bool
		wantErr bool
	}{
		{"missing catalog is an empty set, no error", "", map[string]bool{}, false},
		{"populated catalog", `{"user_datatypes":{"firewall-barracuda":{"kind":"syslog","udp":"7100","tcp":"7101"}}}`,
			map[string]bool{"firewall-barracuda": true}, false},
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

			got, err := UserDataTypeNames()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("UserDataTypeNames() = %v, want error for corrupt catalog", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("UserDataTypeNames() error = %v, want none", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("UserDataTypeNames() = %v, want %v", got, tc.want)
			}
			for k := range tc.want {
				if !got[k] {
					t.Fatalf("UserDataTypeNames() = %v, want %v to be present", got, tc.want)
				}
			}
		})
	}
}
