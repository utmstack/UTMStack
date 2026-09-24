package main

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

// Consumer rules that the normalized protocol field reaches. Events use the shape the
// EventProcessor json plugin produces: top-level EVE keys are sanitized and renamed by the
// filter, nested keys such as flow.bytes_toclient keep their underscores.
func TestSuricataConsumerRules(t *testing.T) {
	cache := plugins.NewCELCache("suricata-consumer-rules")
	load := func(name string) *plugins.Rule {
		b, err := utils.ReadPbYaml(filepath.Join("../..", "rules/nids/suricata", name+".yml"))
		if err != nil {
			t.Fatal(err)
		}
		rule := new(plugins.Rule)
		if err := protojson.Unmarshal(b, rule); err != nil {
			t.Fatal(err)
		}
		rule.Normalize()
		return rule
	}
	tunneling, ddos := load("tunneling_detection"), load("ddos_attack_patterns")
	flow443 := func(log string) string {
		return `{"dataType":"suricata","protocol":"TCP","origin":{"ip":"192.0.2.10","port":50000},` +
			`"target":{"ip":"203.0.113.20","port":443},"log":` + log + `}`
	}
	ntp := func(toClient, toServer int) string {
		return `{"dataType":"suricata","protocol":"UDP","origin":{"ip":"192.0.2.10","port":50000},` +
			`"target":{"ip":"203.0.113.20","port":123},"log":{"eventType":"flow","flow":{"bytes_toclient":` +
			strconv.Itoa(toClient) + `,"bytes_toserver":` + strconv.Itoa(toServer) + `,"pkts_toserver":2}}}`
	}
	cases := []struct {
		name  string
		rule  *plugins.Rule
		event string
		want  bool
	}{
		{"tunneling-443-http-flow", tunneling, flow443(`{"eventType":"flow","appProto":"http"}`), true},
		{"tunneling-443-tls-flow", tunneling, flow443(`{"eventType":"flow","appProto":"tls"}`), false},
		{"tunneling-443-unidentified-flow", tunneling, flow443(`{"eventType":"flow"}`), false},
		{"tunneling-443-failed-detection-flow", tunneling, flow443(`{"eventType":"flow","appProto":"failed"}`), false},
		{"tunneling-443-alert-without-tunnel-signature", tunneling, flow443(`{"eventType":"alert","appProto":"http","alert":{"signature":"ET POLICY example"}}`), false},
		{"ddos-ntp-amplification", ddos, ntp(50000, 100), true},
		{"ddos-ntp-normal-response", ddos, ntp(500, 100), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cache.Eval(tc.rule.Where, tc.event)
			if err != nil || got != tc.want {
				t.Errorf("%s = %v (%v); want %v", tc.rule.Name, got, err, tc.want)
			}
		})
	}
}
