package usecase

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInternalAddressesAreForbidden(t *testing.T) {
	for _, s := range []string{"127.0.0.1", "10.1.2.3", "172.16.0.9", "192.168.1.1", "169.254.169.254", "0.0.0.0", "100.64.0.1", "::1", "fe80::1", "fd00::1", "224.0.0.1"} {
		if !idpAddressForbidden(net.ParseIP(s)) {
			t.Errorf("%s was allowed", s)
		}
	}
	for _, s := range []string{"8.8.8.8", "93.184.216.34", "2606:4700:4700::1111"} {
		if idpAddressForbidden(net.ParseIP(s)) {
			t.Errorf("%s was refused", s)
		}
	}
}

// httptest listens on loopback, which is exactly the address the client must not
// reach.
func TestTheMetadataClientDoesNotReachLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("secret")) }))
	defer srv.Close()

	resp, err := newIDPMetadataClient().Get(srv.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatal("the request reached loopback")
	}
	if !errors.Is(err, errIDPHostForbidden) {
		t.Errorf("err = %v", err)
	}
}
