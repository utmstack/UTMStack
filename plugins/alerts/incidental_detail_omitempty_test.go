package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

// Regression test for the "every new alert is silently dropped" bug.
//
// The v11 alert index template (installer/services/search.go, commit
// 7fb74922da6) maps incidentDetail.creationDate as a `date` field. The alerts
// plugin serializes an IncidentDetail for every alert, and for a non-incident
// alert every field is an empty string. An empty string in a `date` field makes
// OpenSearch reject the document (mapper_parsing_exception: cannot parse empty
// date), so after 3 retries the alert was silently lost.
//
// The fix is to add `omitempty` to the IncidentDetail string fields so that a
// zero-value (empty) incident is NOT serialized at all. This test guarantees
// that a default IncidentDetail produces no incidentDetail key in the JSON —
// i.e. the doc no longer contains an empty creationDate to trip the `date`
// mapping.
func TestIncidentDetailOmitsEmptyFields(t *testing.T) {
	// A default (zero) IncidentDetail, as produced for every non-incident alert.
	var detail IncidentDetail

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(detail); err != nil {
		t.Fatalf("encode default IncidentDetail: %v", err)
	}

	// Every field must be omitted for a zero value.
	var got map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, k := range []string{"createdBy", "observation", "creationDate", "source"} {
		if _, present := got[k]; present {
			t.Errorf("default IncidentDetail must omit %q, got %s", k, buf.String())
		}
	}

	// A populated incident must still serialize all four fields (the incident
	// writer depends on creationDate being present and parseable as a date).
	full := IncidentDetail{
		CreatedBy:    "admin",
		Observation:  "obs",
		CreationDate: "2026-11-11T00:00:00Z",
		Source:       "utmstack",
	}
	fbuf := new(bytes.Buffer)
	if err := json.NewEncoder(fbuf).Encode(full); err != nil {
		t.Fatalf("encode full IncidentDetail: %v", err)
	}
	var fgot map[string]interface{}
	if err := json.Unmarshal(fbuf.Bytes(), &fgot); err != nil {
		t.Fatalf("decode full: %v", err)
	}
	if fgot["creationDate"] != "2026-11-11T00:00:00Z" {
		t.Errorf("full IncidentDetail must keep creationDate, got %v", fgot["creationDate"])
	}
	for _, k := range []string{"createdBy", "observation", "source"} {
		if _, present := fgot[k]; !present {
			t.Errorf("full IncidentDetail must keep %q, got %s", k, fbuf.String())
		}
	}
}

// Guard the shape of the AlertFields document that actually goes to OpenSearch:
// for a non-incident alert the incidentDetail object must NOT contain a
// creationDate key. (AlertFields.IncidentDetail is a value struct, so Go still
// emits an empty "incidentDetail":{} object — that is harmless, because the
// `date`-typed creationDate key is what OpenSearch rejected, not the empty
// object itself.)
func TestAlertFieldsOmitsCreationDateWhenZero(t *testing.T) {
	a := AlertFields{}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(a); err != nil {
		t.Fatalf("encode AlertFields: %v", err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("decode AlertFields: %v", err)
	}
	if detail, ok := got["incidentDetail"].(map[string]interface{}); ok {
		if _, present := detail["creationDate"]; present {
			// This is the exact shape that made OpenSearch reject the doc.
			t.Errorf("AlertFields must not emit a creationDate for a non-incident alert; got %s", buf.String())
		}
	}
}
