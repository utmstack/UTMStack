package signatures

import (
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"time"
)

var (
	oidCatalogCTL     = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 10, 1}
	oidCatalogList    = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 12, 1, 1}
	oidCatalogMember  = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 12, 1, 2}
	oidCatalogMember2 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 12, 1, 3}
	oidCatalogCAB     = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 25}
)

type catalogEntry struct {
	digest []byte
	kind   int64 // reference sentinel: CAB 1, PE 2
}

type verifiedCatalog struct {
	entries []catalogEntry
	// Only members of CRB-rooted verified chains are admitted.
	certs []catalogCertificate
}

// Keep authenticated roles separately from the raw certificate's requested
// EKUs. A code-only CRB chain cannot confer timestamp authority through a
// catalog simply because a child certificate requests both purposes.
type catalogCertificate struct {
	cert               *x509.Certificate
	codeSign, timeSign bool
}

// compileCatalog validates the SignedData and its CRB-rooted signer before
// admitting any member hash. The caller defers .cat processing until all
// certificate rules have been compiled, irrespective of archive order.
func (d *Database) compileCatalog(src source, addError func(string, int, error)) {
	verified, err := d.verifyCatalog(src.data, time.Now().UTC())
	if err != nil {
		if errors.Is(err, errAuthenticodeLimit) {
			addError(src.name, 0, err)
			return
		}
		// A bad catalog contributes no member hashes. Its independently
		// CRB-authenticated certificate paths remain usable, as in the reference.
		newCerts, certErr := d.prepareCatalogCertificates(verified.certs)
		if certErr != nil {
			addError(src.name, 0, certErr)
			return
		}
		d.catalogCerts = newCerts
		d.skipped(SkipInvalidCatalog, 1)
		return
	}
	if d.report.Loaded >= d.limits.MaxSignatures {
		addError(src.name, 0, errors.New("signature count limit exceeded"))
		return
	}
	var arena20, arena32 uint64
	for _, entry := range verified.entries {
		switch len(entry.digest) {
		case 20:
			arena20 += hashNameEncodedSize(len(src.name))
		case 32:
			arena32 += hashNameEncodedSize(len(src.name))
		default:
			addError(src.name, 0, errors.New("invalid catalog digest width"))
			return
		}
	}
	if d.catalog.frozen || uint64(len(d.catalog.groups[1].arena))+arena20 > uint64(^uint32(0)) ||
		uint64(len(d.catalog.groups[2].arena))+arena32 > uint64(^uint32(0)) {
		addError(src.name, 0, errors.New("catalog hash index limit exceeded"))
		return
	}
	newCerts, err := d.prepareCatalogCertificates(verified.certs)
	if err != nil {
		addError(src.name, 0, err)
		return
	}
	for _, entry := range verified.entries {
		if !d.catalog.add(string(entry.digest), entry.kind, src.name) {
			addError(src.name, 0, errors.New("catalog hash index limit exceeded"))
			return
		}
	}
	d.catalogCerts = newCerts
	d.report.Files = append(d.report.Files, src.name)
	d.report.Loaded++ // one authenticated catalog source, many member hashes
}

func (d *Database) prepareCatalogCertificates(certs []catalogCertificate) ([]catalogCertificate, error) {
	prepared := append([]catalogCertificate(nil), d.catalogCerts...)
	seen := make(map[[sha256.Size]byte][]int)
	duplicates := make(authenticodeDuplicateIndex)
	covered := func(entry catalogCertificate) bool {
		identity, err := certificateIdentityFromX509(entry.cert)
		if err != nil {
			return false
		}
		identity.codeSign, identity.timeSign = entry.codeSign, entry.timeSign
		return duplicates.discardIdentity(entry.cert, identity)
	}
	certBytes := 0
	for i, entry := range prepared {
		fingerprint := sha256.Sum256(entry.cert.Raw)
		seen[fingerprint] = append(seen[fingerprint], i)
		covered(entry)
		certBytes += len(entry.cert.Raw)
	}
	for _, entry := range certs {
		if entry.cert == nil {
			continue
		}
		fingerprint := sha256.Sum256(entry.cert.Raw)
		duplicate := false
		for _, i := range seen[fingerprint] {
			prior := prepared[i]
			if (!entry.codeSign || prior.codeSign) && (!entry.timeSign || prior.timeSign) {
				duplicate = true
				break
			}
		}
		if duplicate || covered(entry) {
			continue
		}
		if len(prepared) >= 4096 || len(entry.cert.Raw) > (8<<20)-certBytes {
			return nil, errors.New("catalog certificate index limit exceeded")
		}
		seen[fingerprint] = append(seen[fingerprint], len(prepared))
		certBytes += len(entry.cert.Raw)
		prepared = append(prepared, entry)
	}
	return prepared, nil
}

func (d *Database) verifiedCatalogEntries(der []byte, at time.Time) ([]catalogEntry, error) {
	verified, err := d.verifyCatalog(der, at)
	return verified.entries, err
}

func (d *Database) verifyCatalog(der []byte, at time.Time) (verifiedCatalog, error) {
	var independentlyTrusted []catalogCertificate
	invalid := func() (verifiedCatalog, error) {
		return verifiedCatalog{certs: independentlyTrusted}, errors.New("invalid or untrusted signed catalog")
	}
	if len(der) == 0 || len(der) > 64<<20 {
		return invalid()
	}
	packet, parseErr := parseAuthenticodeSignedData(der, oidCatalogCTL, 4096)
	if packet == nil {
		return invalid()
	}
	store, storeErr := d.resolveAuthenticodeStore(packet.certificates, true)
	if store != nil {
		independentlyTrusted = store.learned
	}
	if storeErr != nil {
		if errors.Is(storeErr, errAuthenticodeLimit) {
			return verifiedCatalog{}, storeErr
		}
		return invalid()
	}
	if parseErr != nil {
		return invalid()
	}
	signer, signerErr := store.primarySigner(packet)
	if errors.Is(signerErr, errAuthenticodeLimit) {
		return verifiedCatalog{}, signerErr
	}
	if signerErr != nil || signer == nil {
		return invalid()
	}
	verifyAt, err := d.legacyAuthenticodeSigningTime(packet, signer, store, at)
	if errors.Is(err, errAuthenticodeLimit) {
		return verifiedCatalog{}, err
	}
	if err != nil || verifyAt.Before(signer.NotBefore) || verifyAt.After(signer.NotAfter) {
		return invalid()
	}
	entries, err := parseCatalogEntries(packet.content)
	if err != nil || len(entries) == 0 {
		return invalid()
	}
	return verifiedCatalog{entries: entries, certs: independentlyTrusted}, nil
}

// parseCatalogEntries reads the CTL member hash records supported by the
// reference: PE and CAB SpcIndirectDataContent with SHA-1 or SHA-256.
func parseCatalogEntries(content []byte) ([]catalogEntry, error) {
	invalid := func() ([]catalogEntry, error) { return nil, errors.New("invalid catalog member list") }
	purpose, rest, err := asn1Object(content)
	if err != nil || !derTag(purpose, asn1.ClassUniversal, asn1.TagSequence) {
		return invalid()
	}
	listOID, tail, err := asn1Object(purpose.Bytes)
	if err != nil || len(tail) != 0 || !derOID(listOID, oidCatalogList) {
		return invalid()
	}
	listID, rest, err := asn1Object(rest)
	if err != nil || listID.Class != asn1.ClassUniversal || listID.Tag != asn1.TagOctetString {
		return invalid()
	}
	effective, rest, err := asn1Object(rest)
	if err != nil || effective.Class != asn1.ClassUniversal || effective.Tag != asn1.TagUTCTime {
		return invalid()
	}
	memberType, rest, err := asn1Object(rest)
	if err != nil || !derTag(memberType, asn1.ClassUniversal, asn1.TagSequence) {
		return invalid()
	}
	memberOID, params, err := asn1Object(memberType.Bytes)
	if err != nil || !(derOID(memberOID, oidCatalogMember) || derOID(memberOID, oidCatalogMember2)) {
		return invalid()
	}
	if len(params) > 0 {
		value, remaining, err := asn1Object(params)
		if err != nil || len(remaining) != 0 || value.Class != asn1.ClassUniversal || value.Tag != asn1.TagNull || value.IsCompound || len(value.Bytes) != 0 {
			return invalid()
		}
	}
	members, tail, err := asn1Object(rest)
	if err != nil || !derTag(members, asn1.ClassUniversal, asn1.TagSequence) {
		return invalid()
	}
	if len(tail) > 0 {
		descriptive, remaining, err := asn1Object(tail)
		if err != nil || len(remaining) != 0 || !derTag(descriptive, asn1.ClassContextSpecific, 0) {
			return invalid()
		}
	}
	var entries []catalogEntry
	for len(members.Bytes) > 0 {
		member, remaining, err := asn1Object(members.Bytes)
		if err != nil || !derTag(member, asn1.ClassUniversal, asn1.TagSequence) {
			return invalid()
		}
		members.Bytes = remaining
		name, fields, err := asn1Object(member.Bytes)
		if err != nil || name.Class != asn1.ClassUniversal || name.Tag != asn1.TagOctetString {
			return invalid()
		}
		attributes, fields, err := asn1Object(fields)
		if err != nil || len(fields) != 0 || !derTag(attributes, asn1.ClassUniversal, asn1.TagSet) {
			return invalid()
		}
		for len(attributes.Bytes) > 0 {
			attribute, remaining, err := asn1Object(attributes.Bytes)
			if err != nil || !derTag(attribute, asn1.ClassUniversal, asn1.TagSequence) {
				return invalid()
			}
			attributes.Bytes = remaining
			oid, values, err := asn1Object(attribute.Bytes)
			if err != nil {
				return invalid()
			}
			if !derOID(oid, oidSPCIndirectData) {
				continue
			}
			set, fields, err := asn1Object(values)
			if err != nil || len(fields) != 0 || !derTag(set, asn1.ClassUniversal, asn1.TagSet) {
				return invalid()
			}
			indirect, fields, err := asn1Object(set.Bytes)
			if err != nil || len(fields) != 0 || !derTag(indirect, asn1.ClassUniversal, asn1.TagSequence) {
				return invalid()
			}
			entry, err := parseCatalogDigest(indirect.Bytes)
			if err != nil {
				return invalid()
			}
			if len(entries) >= 100000 {
				return invalid()
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func parseCatalogDigest(content []byte) (catalogEntry, error) {
	hash, digest, kind, err := authenticodeIndirectDigest(content)
	if err != nil || hash != crypto.SHA1 && hash != crypto.SHA256 {
		return catalogEntry{}, errInvalidAuthenticode
	}
	return catalogEntry{digest: append([]byte(nil), digest...), kind: kind}, nil
}

func (d *Database) HasCatalogHashes() bool {
	return d != nil && d.catalog != nil && d.catalog.len() != 0
}

// CatalogAllows compares an authenticated catalog digest in the reference's
// separate PE/CAB namespaces. Never feed a whole-file PE digest here.
func (d *Database) CatalogAllows(kind int64, digest []byte) bool {
	return d != nil && d.catalog != nil && (kind == 1 || kind == 2) && d.catalog.lookup(string(digest), kind) != ""
}
