//go:build linux

package service

import (
	"os"
	"testing"
)

func TestBoundedLinuxResolverAddresses(t *testing.T) {
	b := []byte("#comment\nnameserver 127.0.0.53\nnameserver 2001:db8::53\nnameserver 127.0.0.53 #duplicate\nnameserver garbage\nnameserver 0.0.0.0\nnameserver ff02::1\nsearch example.invalid\n")
	a := parseLinuxResolvers(b)
	if len(a) != 2 || a[0].String() != "127.0.0.53" || a[1].String() != "2001:db8::53" {
		t.Fatal(a)
	}
	if parseLinuxResolvers(make([]byte, (64<<10)+1)) != nil {
		t.Fatal("oversize accepted")
	}
}

func TestDNSStubRequiresActualUpstreams(t *testing.T) {
	for _, c := range []struct {
		name, main, upstream string
		missing, wantError   bool
		want                 int
	}{
		{"systemd stub", "nameserver 127.0.0.53\n", "nameserver 192.0.2.53\nnameserver 2001:db8::53\n", false, false, 3},
		{"unknown upstream", "nameserver 127.0.0.53\n", "", true, true, 1},
		{"recursive loopback", "nameserver 127.0.0.53\n", "nameserver 127.0.0.1\n", false, true, 1},
		{"direct resolver", "nameserver 192.0.2.53\n", "", true, false, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := discoverLinuxResolvers(func(path string) ([]byte, error) {
				if path == "/etc/resolv.conf" {
					return []byte(c.main), nil
				}
				if c.missing {
					return nil, os.ErrNotExist
				}
				return []byte(c.upstream), nil
			})
			if (err != nil) != c.wantError || len(got) != c.want {
				t.Fatalf("%v %v", got, err)
			}
		})
	}
}
