package usecase

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

var errIDPHostForbidden = errors.New("the identity provider address is not allowed: it resolves to a loopback, private, link-local or otherwise internal address")

var cgnat = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

func idpAddressForbidden(ip net.IP) bool {
	return ip == nil ||
		ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsMulticast() ||
		cgnat.Contains(ip)
}

// newIDPMetadataClient fetches an identity provider's metadata on behalf of a
// tenant administrator, so it must not become a way to reach the platform's own
// network. The address is checked when the connection is made, after DNS has
// answered, which is what stops a name that points inward; redirects are not
// followed and no proxy is consulted, because either would send the request
// somewhere the check never saw.
func newIDPMetadataClient() *http.Client {
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			if idpAddressForbidden(net.ParseIP(host)) {
				return fmt.Errorf("%w: %s", errIDPHostForbidden, host)
			}
			return nil
		},
	}
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DialContext:         dialer.DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}
