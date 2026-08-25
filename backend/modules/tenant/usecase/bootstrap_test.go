package usecase

import "testing"

func TestNormalizeHealHost(t *testing.T) {
	cases := map[string]string{
		"UTM.example.com":                            "utm.example.com",
		"utm.example.com:8443":                       "utm.example.com",
		"utm.customer.com, proxy.internal":           "utm.customer.com",
		"  utm.example.com  ":                        "utm.example.com",
		"":                                           "",
		"[2001:db8::1]:443":                          "2001:db8::1",
	}
	for in, want := range cases {
		if got := normalizeHealHost(in); got != want {
			t.Errorf("normalizeHealHost(%q) = %q, want %q", in, got, want)
		}
	}
}
