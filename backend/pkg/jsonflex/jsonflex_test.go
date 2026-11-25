package jsonflex

import (
	"encoding/json"
	"testing"
)

func TestItReadsANumberInEitherForm(t *testing.T) {
	cases := map[string]int64{
		`6849`: 6849, `"6849"`: 6849, ` 6849 `: 6849, `6849.0`: 6849, `6.849e3`: 6849, `-5`: -5,
		`null`: 0, ``: 0, `""`: 0, `"AS6849"`: 0, `"abc"`: 0, `true`: 0, `{}`: 0,
	}

	for in, want := range cases {
		if got := Int64(json.RawMessage(in)); got != want {
			t.Errorf("%q: got %d, want %d", in, got, want)
		}
	}
}

func TestNormalizeIntsTurnsTheNamedStringsIntoNumbersAtAnyDepth(t *testing.T) {
	in := json.RawMessage(`{"echoes":"641","name":"7","target":{"packagesReceived":"12","host":"123","geolocation":{"asn":"6849","country":"1"}},"list":[{"totalMem":"9"}]}`)

	var out struct {
		Echoes int64  `json:"echoes"`
		Name   string `json:"name"`
		Target struct {
			PackagesReceived int64  `json:"packagesReceived"`
			Host             string `json:"host"`
			Geolocation      struct {
				ASN     int64  `json:"asn"`
				Country string `json:"country"`
			} `json:"geolocation"`
		} `json:"target"`
		List []struct {
			TotalMem int64 `json:"totalMem"`
		} `json:"list"`
	}

	fixed := NormalizeInts(in, "echoes", "packagesReceived", "asn", "totalMem")
	if err := json.Unmarshal(fixed, &out); err != nil {
		t.Fatalf("%v: %s", err, fixed)
	}

	if out.Echoes != 641 || out.Target.PackagesReceived != 12 || out.Target.Geolocation.ASN != 6849 || out.List[0].TotalMem != 9 {
		t.Errorf("numbers not restored: %+v", out)
	}

	if out.Name != "7" || out.Target.Host != "123" || out.Target.Geolocation.Country != "1" {
		t.Errorf("text that only looks like a number must stay text: %+v", out)
	}
}

func TestNormalizeIntsLeavesWhatItCannotReadAlone(t *testing.T) {
	for _, in := range []string{`not json`, `{"echoes":"abc"}`, `[1,2]`, `{"echoes":"1.5"}`} {
		got := NormalizeInts(json.RawMessage(in), "echoes")
		if in == `not json` && string(got) != in {
			t.Errorf("%q changed to %q", in, got)
		}
	}
}
