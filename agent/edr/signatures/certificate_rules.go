package signatures

import (
	"bytes"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
)

// certificateRule retains the identity and policy of one selected CRB entry.
// The subject and serial fields are SHA-1 hashes of the certificate's raw
// ASN.1 subject and serial INTEGER contents, not SHA-1 of the whole cert.
// The RSA modulus is stored as a canonical positive integer byte string.
type certificateRule struct {
	name         string
	blocked      bool
	subject      [20]byte
	serial       [20]byte
	ignoreSerial bool
	modulus      []byte
	codeSign     bool
	timeSign     bool
	certSign     bool
	notBefore    uint64
}

type certificateIdentity struct {
	subject, serial     [20]byte
	modulus             []byte
	codeSign, timeSign  bool
	certSign            bool
	notBefore, notAfter int64
}

// certificateIdentityFromX509 extracts the identifiers the reference uses
// for CRB matching. Go's parsed SerialNumber omits the leading ASN.1 sign
// byte; hashing the original INTEGER content is necessary for parity.
// A parsed certificate alone is not proof that its PE signature is valid.
func certificateIdentityFromX509(cert *x509.Certificate) (certificateIdentity, error) {
	var out certificateIdentity
	if cert == nil {
		return out, errors.New("nil certificate")
	}
	key, ok := authenticodeKey(cert.PublicKey)
	if !ok || key.size() > 513 {
		return out, errors.New("unsupported certificate RSA key")
	}
	subject, rest, err := asn1Object(cert.RawSubject)
	if err != nil || len(rest) != 0 || !derTag(subject, asn1.ClassUniversal, asn1.TagSequence) || len(subject.Bytes) == 0 {
		return out, errors.New("invalid raw certificate subject")
	}
	tbs, rest, err := asn1Object(cert.RawTBSCertificate)
	if err != nil || len(rest) != 0 || tbs.Class != asn1.ClassUniversal || tbs.Tag != asn1.TagSequence {
		return out, errors.New("invalid raw certificate TBS")
	}
	serial, content, err := asn1Object(tbs.Bytes)
	if err != nil {
		return out, err
	}
	if serial.Class == asn1.ClassContextSpecific && serial.Tag == 0 {
		serial, content, err = asn1Object(content)
		if err != nil {
			return out, err
		}
	}
	if serial.Class != asn1.ClassUniversal || serial.Tag != asn1.TagInteger || len(serial.Bytes) == 0 {
		return out, errors.New("invalid raw certificate serial")
	}
	before, after := cert.NotBefore.Unix(), cert.NotAfter.Unix()
	if after <= before {
		return out, errors.New("invalid certificate validity")
	}
	out.subject, out.serial = sha1.Sum(subject.Bytes), sha1.Sum(serial.Bytes)
	out.modulus = key.N.Bytes()
	out.notBefore, out.notAfter = before, after
	policy, err := certificatePolicyFromTBS(cert.RawTBSCertificate)
	if err != nil {
		return out, err
	}
	out.certSign, out.codeSign, out.timeSign = policy.certSign, policy.codeSign, policy.timeSign
	return out, nil
}

func asn1Object(b []byte) (asn1.RawValue, []byte, error) {
	// Local AV framing: the reference uses one tag octet and permits up to
	// four definite length octets, including leading zeroes and long forms
	// for short values. Do not apply this policy to TLS or feed CMS parsing.
	invalid := func() (asn1.RawValue, []byte, error) {
		return asn1.RawValue{}, nil, errInvalidAuthenticode
	}
	if len(b) < 2 {
		return invalid()
	}
	header, size := 2, uint64(b[1])
	if size&0x80 != 0 {
		count := int(size & 0x7f)
		if count == 0 || count > 4 || count > len(b)-header {
			return invalid()
		}
		size = 0
		for _, value := range b[header : header+count] {
			size = size<<8 | uint64(value)
		}
		header += count
	}
	if size > uint64(len(b)-header) {
		return invalid()
	}
	end := header + int(size)
	value := asn1.RawValue{Class: int(b[0] >> 6), Tag: int(b[0] & 0x1f), IsCompound: b[0]&0x20 != 0, Bytes: b[header:end], FullBytes: b[:end]}
	return value, b[end:], nil
}

func crbSHA1(s string) (out [20]byte, err error) {
	if len(s) != 40 {
		return out, errors.New("invalid certificate SHA-1 field")
	}
	_, err = hex.Decode(out[:], []byte(s))
	return out, err
}

func crbModulus(s string) ([]byte, error) {
	if s == "" {
		return nil, errors.New("empty certificate RSA modulus")
	}
	if len(s)%2 != 0 {
		s = "0" + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, errors.New("invalid certificate RSA modulus")
	}
	return bytes.TrimLeft(b, "\x00"), nil
}

func crbFlag(s string) (bool, error) {
	if s == "" {
		return false, errors.New("empty certificate flag")
	}
	switch s[0] {
	case '0':
		return false, nil
	case '1':
		return true, nil
	default:
		return false, errors.New("invalid certificate flag")
	}
}

type certificateSourceState struct {
	ignoreSerial bool
	notBefore    uint64
}

func (d *Database) addCertificateLine(line string) error {
	return d.addCertificateSourceLine(line, &certificateSourceState{})
}

// The legacy CRB loader reuses one certificate record per source. Its serial
// wildcard and omitted not-before fields carry forward after a selected row;
// a new loose file or archive member starts fresh. Keep this format policy
// separate from raw certificate/signer identity comparisons.
func (d *Database) addCertificateSourceLine(line string, state *certificateSourceState) error {
	p := cStrtokenize(line, ';', 14)
	if len(p) < 11 || len(p) > 13 {
		return errors.New("invalid certificate signature fields")
	}
	if len(p) > 11 {
		var maximum *string
		if len(p) == 13 {
			maximum = &p[12]
		}
		if err := d.fieldWindow(p[11], maximum); err != nil {
			return err
		}
	}
	trusted, err := crbFlag(p[1])
	if err != nil {
		return err
	}
	r := certificateRule{name: strings.Clone(p[0]), blocked: !trusted, ignoreSerial: state.ignoreSerial, notBefore: state.notBefore}
	if r.blocked && r.name == "" {
		r.name = "(unnamed CRB rule)" // A matched block must remain a reportable hit.
	}
	if r.subject, err = crbSHA1(p[2]); err != nil {
		return err
	}
	if p[3] == "" {
		r.ignoreSerial = true
		for i := range r.serial {
			r.serial[i] = 0xca
		}
	} else if r.serial, err = crbSHA1(p[3]); err != nil {
		return err
	}
	if r.modulus, err = crbModulus(p[4]); err != nil {
		return err
	}
	if r.codeSign, err = crbFlag(p[6]); err != nil {
		return err
	}
	if r.timeSign, err = crbFlag(p[7]); err != nil {
		return err
	}
	if r.certSign, err = crbFlag(p[8]); err != nil {
		return err
	}
	if p[9] != "" {
		if r.notBefore, err = strconv.ParseUint(p[9], 10, 64); err != nil {
			return errors.New("invalid certificate not-before field")
		}
	}
	state.ignoreSerial, state.notBefore = r.ignoreSerial, r.notBefore
	// Blocklist entries run during PE scanning. Trust entries still need
	// signed-digest and chain verification before they can suppress alerts.
	if !r.blocked {
		d.unknown(UnknownCertificateTrust, 1)
	}
	if !d.duplicateCertificateRule(r) {
		d.certRules = append(d.certRules, r)
	}
	return nil
}

// Duplicate admission follows loaded-list policy, not last-row replacement.
// Blocked records ignore requested roles/validity. Trusted CRB records share
// the loader's issuer/hash/exponent sentinels, and retain distinct wildcard
// modes; an existing record must cover the new record's validity and roles.
func (d *Database) duplicateCertificateRule(candidate certificateRule) bool {
	identity := certificateIdentity{subject: candidate.subject, serial: candidate.serial, modulus: candidate.modulus}
	for i := len(d.certRules) - 1; i >= 0; i-- {
		existing := &d.certRules[i]
		if existing.blocked != candidate.blocked || !existing.matchesIdentity(identity) {
			continue
		}
		if candidate.blocked {
			return true
		}
		if existing.ignoreSerial == candidate.ignoreSerial && candidate.notBefore >= existing.notBefore && (!candidate.codeSign || existing.codeSign) && (!candidate.timeSign || existing.timeSign) && (!candidate.certSign || existing.certSign) {
			return true
		}
	}
	return false
}

func (r *certificateRule) matchesIdentity(cert certificateIdentity) bool {
	return r.subject == cert.subject && (r.ignoreSerial || r.serial == cert.serial) && bytes.Equal(r.modulus, cert.modulus)
}

// blockedCertificate returns the last loaded matching revocation rule, as
// the reference prepends CRB entries to its list. The revocation comparison
// intentionally ignores usage flags and RSA exponent, as the reference does.
func (d *Database) blockedCertificate(cert certificateIdentity) (string, bool) {
	for i := len(d.certRules) - 1; i >= 0; i-- {
		r := &d.certRules[i]
		if r.blocked && r.matchesIdentity(cert) {
			return r.name, true
		}
	}
	return "", false
}

// trustedCertificate compares the CRB trust fields. A matching identity is
// only a trust anchor candidate; the caller must verify the Authenticode
// signed digest and the certificate chain before exempting any detection.
func (d *Database) trustedCertificate(cert certificateIdentity) bool {
	for i := len(d.certRules) - 1; i >= 0; i-- {
		r := &d.certRules[i]
		if r.blocked || !r.matchesIdentity(cert) || cert.notBefore < 0 || uint64(cert.notBefore) < r.notBefore {
			continue
		}
		if (!cert.codeSign || r.codeSign) && (!cert.timeSign || r.timeSign) && (!cert.certSign || r.certSign) {
			return true
		}
	}
	return false
}
