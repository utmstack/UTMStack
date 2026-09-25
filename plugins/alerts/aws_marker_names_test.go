package main

// The parser plugins keep only letters, digits and dots in the field names they write
// (go-sdk utils.SanitizeField), while rules look names up exactly as written. A marker
// added as log.correlationCandidate.mass_resource_deletion is stored as
// ...massresourcedeletion, so a history rule counting the underscored name never fires.
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

// awsHistoryMarkers maps each AWS history rule (file name without extension) to the
// correlation marker that filters/aws/aws.yml adds and the rule counts.
var awsHistoryMarkers = map[string]string{
	"aws_ecs_credential_theft":                           "awsEcsCredentialTheft",
	"aws_golden_saml_attack":                             "samlProviderChange",
	"aws_securityhub_finding_evasion":                    "awsSecurityhubFindingEvasion",
	"aws_ssm_sendcommand_abuse":                          "awsSsmSendcommandAbuse",
	"aws_sso_suspicious_activities":                      "awsSsoSuspiciousActivities",
	"cloudformation_stack_deletion":                      "cloudformationStackDeletion",
	"console_login_impossible_travel":                    "consoleLoginImpossibleTravel",
	"credential_access_aws_iam_assume_role_brute_force":  "credentialAccessAwsIamAssumeRoleBruteForce",
	"credential_access_root_console_failure_brute_force": "credentialAccessRootConsoleFailureBruteForce",
	"cross_account_access_anomalies":                     "crossAccountAccessAnomalies",
	"iam_backdoor_creation_attempts":                     "iamBackdoorCreationAttempts",
	"iam_privilege_escalation_paths":                     "iamPrivilegeEscalationPaths",
	"lambda_privilege_escalation":                        "lambdaPrivilegeEscalation",
	"mass_resource_deletion":                             "massResourceDeletion",
	"route53_dns_hijacking":                              "route53DnsHijacking",
	"s3_bulk_data_exfiltration":                          "s3BulkDataExfiltration",
	"secrets_manager_suspicious_access":                  "secretsManagerSuspiciousAccess",
	"security_group_modifications":                       "securityGroupModifications",
	"ssm_session_abuse":                                  "ssmSessionAbuse",
	"sts_token_abuse":                                    "stsTokenAbuse",
	"unusual_api_call_patterns":                          "unusualApiCallPatterns",
	"vpc_flow_log_anomalies":                             "vpcFlowLogAnomalies",
}

// Every marker the AWS filter adds and every marker an AWS rule reads (where, history
// fields and placeholders, groupBy, deduplicateBy) must be the same string, made only of
// letters, digits and dots, so the stored name is the name the rule looks up.
func TestAWSCorrelationMarkerNames(t *testing.T) {
	const prefix = "log.correlationCandidate."
	clean := regexp.MustCompile(`^[A-Za-z0-9.]+$`)
	kept := func(name string) bool {
		stored := name
		utils.SanitizeField(&stored)
		return stored == name && clean.MatchString(name)
	}

	written := map[string]string{}
	for _, stage := range awsConfig(t).Pipeline {
		if !slices.Contains(stage.DataTypes, "aws") {
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
		if err = json.Unmarshal(b, &head); err != nil || !slices.Contains(head.DataTypes, "aws") {
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
					t.Errorf("%s reads %q, which the AWS filter does not add", path, marker)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{}
	for rule, marker := range awsHistoryMarkers {
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
			if prefix+awsHistoryMarkers[rule] != marker {
				t.Errorf("%s reads %s, expected only %s%s", rule, marker, prefix, awsHistoryMarkers[rule])
			}
		}
	}
	for marker := range written {
		if !expected[marker] {
			t.Errorf("filter adds %s, which no AWS history rule counts", marker)
		}
	}
	if len(written) != 22 || len(awsHistoryMarkers) != 22 {
		t.Errorf("marker coverage: filter adds %d, table lists %d", len(written), len(awsHistoryMarkers))
	}
}
