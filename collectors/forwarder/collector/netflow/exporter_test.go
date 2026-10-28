package netflow

import "testing"

func TestTheExporterIsIdentifiedByHostOnly(t *testing.T) {
	for in, want := range map[string]string{
		"10.0.0.5:51234":        "10.0.0.5",
		"[2001:db8::1]:2055":    "2001:db8::1",
		"router.example.com:99": "router.example.com",
		"no-port":               "no-port",
	} {
		if got := netflowExporter(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}
