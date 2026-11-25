package domain

import (
	"encoding/json"
	"testing"
)

// The event processor writes the asn of a geolocation as text. An alert carrying one has to stay readable.
func TestAnAlertWithAGeolocationWhoseAsnIsTextIsReadable(t *testing.T) {
	doc := `{"id":"a1","name":"n","adversary":{"ip":"1.2.3.4","geolocation":{"country":"Ukraine","asn":"6849","aso":"Ukrtelecom","latitude":50.4,"accuracy":100}}}`

	var a UtmAlert
	if err := json.Unmarshal([]byte(doc), &a); err != nil {
		t.Fatalf("the alert could not be read: %v", err)
	}

	g := a.Adversary.Geolocation
	if g == nil || g.ASN != 6849 || g.Country != "Ukraine" || g.ASO != "Ukrtelecom" || g.Accuracy != 100 || g.Latitude != 50.4 {
		t.Errorf("geolocation %+v", g)
	}
}

func TestTheAsnStillReadsAsANumberAndAMissingOneAsZero(t *testing.T) {
	for doc, want := range map[string]int64{
		`{"asn":6849,"country":"X"}`: 6849, `{"country":"X"}`: 0, `{"asn":null}`: 0, `{"asn":""}`: 0,
	} {
		var g Geolocation
		if err := json.Unmarshal([]byte(doc), &g); err != nil || g.ASN != want {
			t.Errorf("%s: asn %d err %v", doc, g.ASN, err)
		}
	}
}

func TestWhatIsWrittenBackKeepsTheAsnAsANumber(t *testing.T) {
	out, err := json.Marshal(Geolocation{Country: "X", ASN: 6849})
	if err != nil || string(out) != `{"country":"X","asn":6849}` {
		t.Errorf("%s %v", out, err)
	}
}
