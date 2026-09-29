package main

// The parser plugins keep only letters, digits and dots in the field names they write
// (go-sdk utils.SanitizeField), while rules look names up exactly as written. A marker
// added as log.correlationCandidate.key_vault_access_spikes is stored as
// ...keyvaultaccessspikes, so a history rule counting the underscored name never fires.
import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// azureHistoryMarkers maps each Azure history rule (file name without extension) to the
// correlation marker that filters/azure/azure-eventhub.yml adds and the rule counts.
var azureHistoryMarkers = map[string]string{
	"aks_security_threats":           "aksSecurityThreats",
	"app_registration_abuse":         "appRegistrationAbuse",
	"application_gateway_waf_alerts": "applicationGatewayWafAlerts",
	"azure_ad_password_spray":        "azureAdPasswordSpray",
	"azure_bulk_role_changes":        "azureBulkRoleChanges",
	"azure_kubernetes_secret_access": "azureKubernetesSecretAccess",
	"azure_laps_credential_dump":     "azureLapsCredentialDump",
	"azure_ropc_authentication":      "azureRopcAuthentication",
	"key_vault_access_spikes":        "keyVaultAccessSpikes",
	"pim_role_activation_abuse":      "pimRoleActivationAbuse",
}

// Every marker the Azure filter adds and every marker an Azure rule reads (where, history
// fields and placeholders, groupBy, deduplicateBy) must be the same string, made only of
// letters, digits and dots, so the stored name is the name the rule looks up.
func TestAzureCorrelationMarkerNames(t *testing.T) {
	const prefix = "log.correlationCandidate."
	clean := regexp.MustCompile(`^[A-Za-z0-9.]+$`)
	kept := func(name string) bool {
		stored := name
		utils.SanitizeField(&stored)
		return stored == name && clean.MatchString(name)
	}

	written := map[string]string{}
	for _, stage := range azureConfig(t).Pipeline {
		if !slices.Contains(stage.DataTypes, "azure") {
			continue
		}
		for _, step := range stage.Steps {
			names := []string{}
			if s := step.Grok; s != nil {
				for _, p := range s.Patterns {
					if p.FieldName != "" {
						names = append(names, p.FieldName)
					}
				}
			}
			if s := step.Rename; s != nil {
				names = append(names, s.To)
			}
			if s := step.Csv; s != nil {
				names = append(names, s.Headers...)
			}
			if s := step.Add; s != nil {
				key := s.Params["key"].GetStringValue()
				names = append(names, key)
				if strings.HasPrefix(key, prefix) {
					value := s.Params["value"].GetStringValue()
					if previous, ok := written[key]; ok && previous != value {
						t.Errorf("filter adds %s as %q and %q", key, previous, value)
					}
					written[key] = value
				}
			}
			for _, name := range names {
				if !kept(name) {
					t.Errorf("filter writes %q, which the parser stores under another name", name)
				}
			}
		}
	}

	reference := regexp.MustCompile(`log\.correlationCandidate\.[^"'\s,()\[\]{}]*`)
	read := map[string]map[string]bool{}
	err := filepath.WalkDir("../../rules", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || (filepath.Ext(path) != ".yml" && filepath.Ext(path) != ".yaml") {
			return err
		}
		b, err := utils.ReadPbYaml(path)
		if err != nil {
			return err
		}
		var head struct {
			DataTypes []string `json:"dataTypes"`
		}
		if err = json.Unmarshal(b, &head); err != nil || !slices.Contains(head.DataTypes, "azure") {
			return err
		}
		rule := new(plugins.Rule)
		if err = protojson.Unmarshal(b, rule); err != nil {
			return err
		}
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		texts := []string{rule.Where}
		texts = append(texts, rule.GroupBy...)
		texts = append(texts, rule.DeduplicateBy...)
		var searches func([]*plugins.SearchRequest)
		searches = func(list []*plugins.SearchRequest) {
			for _, search := range list {
				for _, term := range search.With {
					value := term.Value.GetStringValue()
					texts = append(texts, term.Field, value)
					if strings.HasPrefix(term.Field, prefix) && written[term.Field] != value {
						t.Errorf("%s counts %s = %q; the filter adds %q", path, term.Field, value, written[term.Field])
					}
				}
				searches(search.Or)
			}
		}
		searches(rule.AfterEvents)
		searches(rule.Correlation)
		for _, text := range texts {
			for _, marker := range reference.FindAllString(text, -1) {
				marker = strings.TrimSuffix(marker, ".keyword")
				if read[marker] == nil {
					read[marker] = map[string]bool{}
				}
				read[marker][name] = true
				if !kept(marker) {
					t.Errorf("%s reads %q, which the parser never stores under that name", path, marker)
				}
				if _, ok := written[marker]; !ok {
					t.Errorf("%s reads %q, which the Azure filter does not add", path, marker)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{}
	for rule, marker := range azureHistoryMarkers {
		expected[prefix+marker] = true
		if _, ok := written[prefix+marker]; !ok {
			t.Errorf("filter does not add %s%s for %s", prefix, marker, rule)
		}
		if !read[prefix+marker][rule] {
			t.Errorf("%s does not read %s%s", rule, prefix, marker)
		}
	}
	for marker, rules := range read {
		for rule := range rules {
			if prefix+azureHistoryMarkers[rule] != marker {
				t.Errorf("%s reads %s, expected only %s%s", rule, marker, prefix, azureHistoryMarkers[rule])
			}
		}
	}
	for marker := range written {
		if !expected[marker] {
			t.Errorf("filter adds %s, which no Azure history rule counts", marker)
		}
	}
	if len(written) != 10 || len(azureHistoryMarkers) != 10 {
		t.Errorf("marker coverage: filter adds %d, table lists %d", len(written), len(azureHistoryMarkers))
	}
}
