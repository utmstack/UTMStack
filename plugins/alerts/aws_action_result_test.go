package main

import (
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
)

// These are fabricated raw CloudTrail records exercised by the existing offline
// extraction model and the pinned SDK CEL evaluator. They do not run the closed
// EventProcessor or feeds plugin and do not establish that a TI alert was created.
func TestAWSAuthenticationOutcome(t *testing.T) {
	cfg, cache := awsConfig(t), plugins.NewCELCache("aws-authentication-outcome")
	for _, tc := range []struct {
		name, raw, result string
	}{
		{"MFA requirement check", `{"eventVersion":"1.09","eventSource":"signin.amazonaws.com","eventType":"AwsConsoleSignIn","eventName":"CheckMfa","sourceIPAddress":"192.0.2.10","responseElements":{"CheckMfa":"Success"}}`, ""},
		{"failed check", `{"eventVersion":"1.09","eventSource":"signin.amazonaws.com","eventType":"AwsConsoleSignIn","eventName":"CheckMfa","sourceIPAddress":"192.0.2.10","responseElements":{"CheckMfa":"Failure"}}`, "failed"},
		{"completed sign-in", `{"eventVersion":"1.09","eventSource":"signin.amazonaws.com","eventType":"AwsConsoleSignIn","eventName":"ConsoleLogin","sourceIPAddress":"192.0.2.10","responseElements":{"ConsoleLogin":"Success"}}`, "success"},
		{"failed sign-in", `{"eventVersion":"1.09","eventSource":"signin.amazonaws.com","eventType":"AwsConsoleSignIn","eventName":"ConsoleLogin","sourceIPAddress":"192.0.2.10","responseElements":{"ConsoleLogin":"Failure"}}`, "failed"},
		{"missing sign-in outcome", `{"eventVersion":"1.09","eventSource":"signin.amazonaws.com","eventType":"AwsConsoleSignIn","eventName":"ConsoleLogin","sourceIPAddress":"192.0.2.10","responseElements":null}`, ""},
		{"completed role switch", `{"eventVersion":"1.08","eventSource":"signin.amazonaws.com","eventType":"AwsConsoleSignIn","eventName":"SwitchRole","sourceIPAddress":"192.0.2.10","responseElements":{"SwitchRole":"Success"}}`, "success"},
		{"failed role switch", `{"eventVersion":"1.08","eventSource":"signin.amazonaws.com","eventType":"AwsConsoleSignIn","eventName":"SwitchRole","sourceIPAddress":"192.0.2.10","responseElements":{"SwitchRole":"Failure"}}`, "failed"},
		{"Identity Center challenge", `{"eventVersion":"1.08","eventSource":"signin.amazonaws.com","eventType":"AwsServiceEvent","eventName":"CredentialChallenge","sourceIPAddress":"192.0.2.10","serviceEventDetails":{"CredentialChallenge":"Success"}}`, ""},
		{"Identity Center verified credential", `{"eventVersion":"1.08","eventSource":"signin.amazonaws.com","eventType":"AwsServiceEvent","eventName":"CredentialVerification","sourceIPAddress":"192.0.2.10","serviceEventDetails":{"CredentialVerification":"Success"}}`, ""},
		{"Identity Center failed credential", `{"eventVersion":"1.08","eventSource":"signin.amazonaws.com","eventType":"AwsServiceEvent","eventName":"CredentialVerification","sourceIPAddress":"192.0.2.10","serviceEventDetails":{"CredentialVerification":"Failure"}}`, "failed"},
		{"Identity Center completed sign-in", `{"eventVersion":"1.08","eventSource":"signin.amazonaws.com","eventType":"AwsServiceEvent","eventName":"UserAuthentication","sourceIPAddress":"192.0.2.10","serviceEventDetails":{"UserAuthentication":"Success"}}`, "success"},
		{"Grafana change with HTTP 200", `{"eventVersion":"1.08","eventSource":"grafana.amazonaws.com","eventType":"AwsServiceEvent","eventName":"create","sourceIPAddress":"192.0.2.10","errorCode":"200","serviceEventDetails":{"result":{"statusType":"success","statusCode":"200"}}}`, "success"},
		{"Grafana change with HTTP 412", `{"eventVersion":"1.08","eventSource":"grafana.amazonaws.com","eventType":"AwsServiceEvent","eventName":"update","sourceIPAddress":"192.0.2.10","errorCode":"412","errorMessage":"changed by someone else","serviceEventDetails":{"result":{"statusType":"failure","statusCode":"412"}}}`, "failed"},
		{"Grafana failed SSO login", `{"eventVersion":"1.08","eventSource":"grafana.amazonaws.com","eventType":"AwsServiceEvent","eventName":"login-auth.sso","sourceIPAddress":"192.0.2.10","serviceEventDetails":{"action":"login-auth.sso","result":{"statusType":"failure"}}}`, "failed"},
		{"denial overrides success", `{"eventVersion":"1.09","eventSource":"signin.amazonaws.com","eventType":"AwsConsoleSignIn","eventName":"ConsoleLogin","sourceIPAddress":"192.0.2.10","errorCode":"AccessDenied","responseElements":{"ConsoleLogin":"Success"}}`, "denied"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := awsParse(t, cfg, tc.raw, "synthetic-collector", cache)
			result := gjson.Get(out, "actionResult")
			if result.String() != tc.result || (tc.result == "" && result.Exists()) {
				t.Fatalf("actionResult = %s, want %q", result.Raw, tc.result)
			}
			eligible, err := cache.Eval(`equals("actionResult","success") && inCIDR("origin.ip","0.0.0.0/0")`, out)
			if err != nil || eligible != (tc.result == "success") {
				t.Fatalf("SDK success-and-IP predicate = %v, %v", eligible, err)
			}
			if tc.name == "MFA requirement check" && gjson.Get(out, "log.responseElements.CheckMfa").String() != "Success" {
				t.Fatal("vendor check result was lost")
			}
		})
	}
}
