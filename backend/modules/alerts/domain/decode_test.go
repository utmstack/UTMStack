package domain

import (
	"encoding/json"
	"testing"
)

// What the event processor stores: 64-bit integers written as strings. An alert like this used to fail to
// decode, and every search that decodes alerts left it out without a word.
const storedAlert = `{
  "id": "a1", "name": "Known Malicious IP Detected", "severity": "low", "status": "Open", "tenantId": "t1",
  "echoes": "3",
  "adversary": {"ip": "1.2.3.4", "geolocation": {"country": "Ukraine", "asn": "6849"}},
  "target": {"ip": "10.0.0.5", "host": "srv", "port": 443, "packagesSent": "17", "packagesReceived": "12",
             "connections": "4", "totalMem": "8589934592"}
}`

func TestAnAlertWith64BitIntegersWrittenAsTextDecodes(t *testing.T) {
	var a UtmAlert
	if err := json.Unmarshal([]byte(storedAlert), &a); err != nil {
		t.Fatalf("the alert did not decode: %v", err)
	}

	if a.Echoes != 3 || a.Adversary.Geolocation.ASN != 6849 || a.Target.PackagesReceived != 12 ||
		a.Target.PackagesSent != 17 || a.Target.Connections != 4 || a.Target.TotalMem != 8589934592 {
		t.Errorf("numbers: %+v / %+v", a, a.Target)
	}

	if a.ID != "a1" || a.Name != "Known Malicious IP Detected" || a.Target.Host != "srv" || a.Target.Port != 443 {
		t.Errorf("the rest of the alert was lost: %+v", a)
	}
}

func TestAnAlertThatIsWrittenWithNumbersStillDecodes(t *testing.T) {
	var a UtmAlert
	doc := `{"id":"a1","echoes":3,"target":{"packagesReceived":12,"geolocation":{"asn":6849}}}`

	if err := json.Unmarshal([]byte(doc), &a); err != nil || a.Echoes != 3 || a.Target.PackagesReceived != 12 || a.Target.Geolocation.ASN != 6849 {
		t.Errorf("err %v, %+v", err, a)
	}
}

func TestAnAlertThatIsReallyBrokenStillFails(t *testing.T) {
	var a UtmAlert
	if err := json.Unmarshal([]byte(`{"id":"a1","severity":["not","a","string"]}`), &a); err == nil {
		t.Error("a document that cannot be an alert must not decode")
	}
}
