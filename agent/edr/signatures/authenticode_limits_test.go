package signatures

import (
	"bytes"
	"crypto/x509"
	"errors"
	"testing"
)

func TestAuthenticodePrimaryWorkLimitIsIncomplete(t *testing.T) {
	seed, _ := primaryParserSeed(t)
	packet, err := parseAuthenticodeSignedData(seed, oidSPCIndirectData, 256)
	if err != nil {
		t.Fatal(err)
	}
	d := &Database{}
	verified, err := d.authenticodePrimarySignersChecked(packet)
	if err != nil || len(verified) != 1 {
		t.Fatal("seed did not authenticate", err)
	}
	var candidates []*x509.Certificate
	for i := 0; i < 65; i++ {
		// Valid framing and identical signing identity with distinct issuer
		// signature bytes. This isolates the later primary-proof work budget.
		data := bytes.Clone(verified[0].Raw)
		data[len(data)-2], data[len(data)-1] = byte(i>>8), byte(i)
		cert, err := parseAuthenticodeCertificate(data)
		if err != nil {
			t.Fatal(err)
		}
		candidates = append(candidates, cert)
	}
	packet.certificates = candidates[:64]
	if _, err := d.authenticodePrimarySignersChecked(packet); err != nil {
		t.Fatal("exact candidate window refused", err)
	}
	packet.certificates = candidates
	if got, err := d.authenticodePrimarySignersChecked(packet); len(got) != 0 || !errors.Is(err, errAuthenticodeLimit) {
		t.Fatal("work exhaustion became an invalid signature", err)
	}
	packet.certificates = nil
	if got, err := d.authenticodePrimarySignersChecked(packet); len(got) != 0 || err != nil {
		t.Fatal("known missing signer became unsupported analysis", err)
	}
}
