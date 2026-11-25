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
