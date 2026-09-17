package main

import (
	"github.com/tidwall/gjson"
	"testing"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAlertGroupingValue(t *testing.T) {
	eventLog, err := structpb.NewStruct(map[string]any{"eventCode": 4625, "isFailure": true, "accounts": []any{"alice", "bob"}})
	if err != nil {
		t.Fatal(err)
	}
	alert := &plugins.Alert{
		Adversary: &plugins.Side{Ip: "203.0.113.7"},
		Target:    &plugins.Side{Host: "dc01"},
		Events: []*plugins.Event{
			{Action: "previous", Origin: &plugins.Side{User: "previous-user"}},
			{Action: "login", Origin: &plugins.Side{User: "alice"}, Log: eventLog.Fields},
		},
	}
	serialized, err := utils.ProtoMessageToString(alert)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ field, want string }{
		{"adversary.ip", "203.0.113.7"}, {"target.host", "dc01"},
		{"lastEvent.action", "login"}, {"lastEvent.origin.user", "alice"},
		{"lastEvent.log.eventCode", "4625"}, {"lastEvent.log.isFailure", "true"},
		{"lastEvent.log.accounts.1", "bob"},
	} {
		t.Run(tc.field, func(t *testing.T) {
			got := alertGroupingValue(*serialized, tc.field)
			if !got.Exists() || got.String() != tc.want {
				t.Fatalf("got %v, want %s", got, tc.want)
			}
		})
	}
	for _, field := range []string{"lastEvent.log.missing", "origin.ip"} {
		if alertGroupingValue(*serialized, field).Exists() {
			t.Errorf("unexpected value for %s", field)
		}
	}
}
func TestAlertGroupingValueWithoutEvents(t *testing.T) {
	for _, input := range []string{`{}`, `{"events":[]}`, `{"events":[null]}`} {
		if alertGroupingValue(input, "lastEvent.origin.ip").Exists() {
			t.Errorf("unexpected lastEvent for %s", input)
		}
	}
}

func TestScalarGroupingValue(t *testing.T) {
	for _, tc := range []struct {
		input string
		valid bool
	}{
		{`"alice"`, true}, {`0`, true}, {`false`, true}, {`true`, true},
		{`null`, false}, {`[]`, false}, {`["alice"]`, false}, {`{"user":"alice"}`, false},
	} {
		if _, ok := scalarGroupingValue(gjson.Parse(tc.input)); ok != tc.valid {
			t.Errorf("scalarGroupingValue(%s) = %v, want %v", tc.input, ok, tc.valid)
		}
	}
}
