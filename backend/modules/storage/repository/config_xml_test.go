package repository

import (
	"os"
	"strings"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/storage/domain"
)

func TestALocalDiskCarriesNoCredentialsOrCache(t *testing.T) {
	dir := t.TempDir()
	r := NewConfigRepository(dir)

	if _, err := r.Write(domain.ColdTarget{Kind: domain.KindLocal}.Normalized()); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(dir + "/" + configFileName)
	for _, want := range []string{"<type>local</type>", "<path>" + domain.LocalMount + "</path>"} {
		if !strings.Contains(string(body), want) {
			t.Errorf("missing %s in\n%s", want, body)
		}
	}
	for _, bad := range []string{"access_key_id", "secret_access_key", "cache_enabled", "metadata_path"} {
		if strings.Contains(string(body), bad) {
			t.Errorf("local disk carries %s", bad)
		}
	}

	got, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != domain.KindLocal || got.Endpoint != domain.LocalMount {
		t.Errorf("read back %+v", got)
	}
}

func TestAnS3DiskRoundTrips(t *testing.T) {
	r := NewConfigRepository(t.TempDir())
	o := domain.ColdTarget{Endpoint: "https://s3.example.com/b", AccessKey: "a", SecretKey: "s"}.Normalized()
	if _, err := r.Write(o); err != nil {
		t.Fatal(err)
	}
	got, _ := r.Read()
	if got.Kind != domain.KindS3 || got.Endpoint != "https://s3.example.com/b/" {
		t.Errorf("read back %+v", got)
	}
}

func TestAzureAndRoleDisksRenderTheirOwnFields(t *testing.T) {
	dir := t.TempDir()
	r := NewConfigRepository(dir)

	az := domain.ColdTarget{Kind: domain.KindAzure, Endpoint: "https://acct.blob.core.windows.net", Container: "cold", AccessKey: "acct", SecretKey: "key"}.Normalized()
	if _, err := r.Write(az); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(dir + "/" + configFileName)
	for _, want := range []string{"<type>object_storage</type>", "<object_storage_type>azure_blob_storage</object_storage_type>", "<container_name>cold</container_name>", "<account_name>acct</account_name>"} {
		if !strings.Contains(string(body), want) {
			t.Errorf("missing %s in\n%s", want, body)
		}
	}
	got, _ := r.Read()
	if got.Kind != domain.KindAzure || got.Endpoint != "https://acct.blob.core.windows.net/cold" {
		t.Errorf("read back %+v", got)
	}

	role := domain.ColdTarget{Endpoint: "https://s3.eu-west-1.amazonaws.com/cold", UseRole: true}.Normalized()
	if _, err := r.Write(role); err != nil {
		t.Fatal(err)
	}
	body, _ = os.ReadFile(dir + "/" + configFileName)
	if !strings.Contains(string(body), "<use_environment_credentials>true</use_environment_credentials>") {
		t.Errorf("role not rendered:\n%s", body)
	}
	if strings.Contains(string(body), "<access_key_id>") {
		t.Error("a role disk carries keys")
	}
}
