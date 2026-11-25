package coordination

import (
	"regexp"
	"testing"
)

var natsSafe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func TestNatsTokenCharset(t *testing.T) {
	for _, name := range []string{"", ".", "..", ".hidden", "a b", "a/b", "a=b", "ñandú", "Prod AWS", "a.b.c", "*", ">"} {
		got := NatsToken(name)
		if !natsSafe.MatchString(got) {
			t.Errorf("NatsToken(%q) = %q, outside the NATS-safe charset", name, got)
		}
	}
}

func TestNatsTokenInjective(t *testing.T) {
	names := []string{"", "_", "grp.1", "grp_1", "grp 1", "Prod AWS", "Prod_AWS", "Tenant A", "Tenant_A", ".foo", "foo", "foo."}
	seen := map[string]string{}
	for _, n := range names {
		tok := NatsToken(n)
		if prev, dup := seen[tok]; dup {
			t.Fatalf("%q and %q both encode to %q", prev, n, tok)
		}
		seen[tok] = n
	}
}
