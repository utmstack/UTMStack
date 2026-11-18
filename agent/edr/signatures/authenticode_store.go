package signatures

import (
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"math/big"
)

// A store entry keeps authenticated purposes separate from the purposes in
// the original signed certificate. A competing parent cannot contribute an
// additional purpose after a different parent has already admitted the child.
type authenticodeStoreEntry struct {
	cert        *x509.Certificate
	rule        *certificateRule
	identity    certificateIdentity
	issuer      [sha1.Size]byte
	key         *authenticodeRSAKey
	hash        crypto.Hash
	path        []*x509.Certificate
	fingerprint [sha256.Size]byte
	modulus     string
}

type authenticodeTrustStore struct {
	entries    []*authenticodeStoreEntry // Lookup priority, newest admission first.
	learned    []catalogCertificate      // Actual admission order, oldest first.
	work       int
	proofs     int
	rsaWork    uint64
	uncertain  bool
	checks     map[authenticodeStoreProof]bool
	duplicates map[authenticodeStoreDuplicate][]*authenticodeStoreEntry
	parents    map[[sha1.Size]byte][]*authenticodeStoreEntry
}

type authenticodeStoreDuplicate struct {
	subject, issuer, serial [sha1.Size]byte
	hash                    crypto.Hash
	exponent                string
	modulus                 string
	crb, ignoreSerial       bool
}

func (e *authenticodeStoreEntry) duplicateKey() authenticodeStoreDuplicate {
	key := authenticodeStoreDuplicate{subject: e.identity.subject, issuer: e.issuer, serial: e.identity.serial, hash: e.hash, exponent: string(e.key.E.Bytes()), modulus: e.modulus, crb: e.rule != nil}
	if e.rule != nil {
		key.ignoreSerial = e.rule.ignoreSerial
	}
	return key
}

type authenticodeStoreProof struct {
	child   [sha256.Size]byte
	modulus string
	exp     string
}

func storeCertificateEntry(cert *x509.Certificate) (*authenticodeStoreEntry, error) {
	identity, err := certificateIdentityFromX509(cert)
	if err != nil {
		return nil, err
	}
	issuer, tail, err := asn1Object(cert.RawIssuer)
	if err != nil || len(tail) != 0 {
		return nil, errInvalidAuthenticode
	}
	key, _ := authenticodeKey(cert.PublicKey)
	return &authenticodeStoreEntry{cert: cert, identity: identity, issuer: sha1.Sum(issuer.Bytes), key: key, hash: authenticodeCertificateHash(cert.SignatureAlgorithm), fingerprint: sha256.Sum256(cert.Raw), modulus: string(identity.modulus)}, nil
}

func (s *authenticodeTrustStore) step() error {
	s.work++
	if s.work > 1<<20 {
		return errAuthenticodeLimit
	}
	return nil
}

func (s *authenticodeTrustStore) duplicate(candidate *authenticodeStoreEntry) (bool, error) {
	if err := s.step(); err != nil {
		return false, err
	}
	for _, prior := range s.duplicates[candidate.duplicateKey()] {
		if err := s.step(); err != nil {
			return false, err
		}
		a, b := prior.identity, candidate.identity
		if a.subject == b.subject && a.serial == b.serial && b.notBefore >= a.notBefore && b.notAfter <= a.notAfter && (!b.codeSign || a.codeSign) && (!b.timeSign || a.timeSign) && (!b.certSign || a.certSign) {
			return true, nil
		}
	}
	return false, nil
}

func (s *authenticodeTrustStore) admit(entry *authenticodeStoreEntry, learn bool) error {
	duplicate, err := s.duplicate(entry)
	if err != nil || duplicate {
		return err
	}
	if len(s.entries) >= 8192 {
		return errAuthenticodeLimit
	}
	s.entries = append(s.entries, nil)
	copy(s.entries[1:], s.entries)
	s.entries[0] = entry
	key := entry.duplicateKey()
	s.duplicates[key] = append(s.duplicates[key], entry)
	if entry.identity.certSign {
		parents := s.parents[entry.identity.subject]
		parents = append(parents, nil)
		copy(parents[1:], parents)
		parents[0] = entry
		s.parents[entry.identity.subject] = parents
	}
	if learn && entry.cert != nil {
		s.learned = append(s.learned, catalogCertificate{cert: entry.cert, codeSign: entry.identity.codeSign, timeSign: entry.identity.timeSign})
	}
	return nil
}

func (s *authenticodeTrustStore) parent(child *authenticodeStoreEntry) (*authenticodeStoreEntry, error) {
	var selected *authenticodeStoreEntry
	best := 0
	for _, parent := range s.parents[child.issuer] {
		if err := s.step(); err != nil {
			return nil, err
		}
		if !parent.identity.certSign || parent.identity.subject != child.issuer {
			continue
		}
		proof := authenticodeStoreProof{child: child.fingerprint, modulus: parent.modulus, exp: string(parent.key.E.Bytes())}
		valid, checked := s.checks[proof]
		if !checked {
			s.proofs++
			if s.proofs > 4096 {
				return nil, errAuthenticodeLimit
			}
			if err := s.chargeRSA(parent.key); err != nil {
				return nil, err
			}
			valid = verifyAuthenticodeCertificateSignature(child.cert, parent.key)
			s.checks[proof] = valid
		}
		if !valid {
			continue
		}
		a, b := parent.identity, child.identity
		if (!b.codeSign || a.codeSign) && (!b.timeSign || a.timeSign) {
			return parent, nil
		}
		score := 0
		if a.codeSign && b.codeSign {
			score++
		}
		if a.timeSign && b.timeSign {
			score++
		}
		if score > best {
			selected, best = parent, score
		}
	}
	return selected, nil
}

// Catalog processing mutates the global trust order. PE scanning instead
// copies that store by insertion, reversing lookup priority. Both then admit
// the reverse-wire pending bag and restart after each successful promotion.
func (d *Database) resolveAuthenticodeStore(certs []*x509.Certificate, catalog bool) (*authenticodeTrustStore, error) {
	if len(certs)+len(d.catalogCerts) > 4096 {
		return nil, errAuthenticodeLimit
	}
	s := &authenticodeTrustStore{checks: make(map[authenticodeStoreProof]bool), duplicates: make(map[authenticodeStoreDuplicate][]*authenticodeStoreEntry), parents: make(map[[sha1.Size]byte][]*authenticodeStoreEntry)}
	for i := range d.certRules {
		rule := &d.certRules[i]
		if rule.blocked {
			continue
		}
		entry := &authenticodeStoreEntry{rule: rule, identity: certificateIdentity{subject: rule.subject, serial: rule.serial, modulus: rule.modulus, codeSign: rule.codeSign, timeSign: rule.timeSign, certSign: rule.certSign, notBefore: int64(rule.notBefore), notAfter: 1<<63 - 1}, key: &authenticodeRSAKey{N: new(big.Int).SetBytes(rule.modulus), E: big.NewInt(65537)}}
		entry.modulus = string(rule.modulus)
		// CRB issuer/hash/exponent sentinels are a separate identity class.
		if err := s.admit(entry, false); err != nil {
			return nil, err
		}
	}
	for _, trusted := range d.catalogCerts {
		entry, err := storeCertificateEntry(trusted.cert)
		if err != nil {
			return nil, errInvalidAuthenticode
		}
		entry.identity.codeSign, entry.identity.timeSign = trusted.codeSign, trusted.timeSign
		entry.path = []*x509.Certificate{entry.cert}
		if err := s.admit(entry, false); err != nil {
			return nil, err
		}
	}
	if !catalog {
		global := s.entries
		s.entries = nil
		s.duplicates = make(map[authenticodeStoreDuplicate][]*authenticodeStoreEntry)
		s.parents = make(map[[sha1.Size]byte][]*authenticodeStoreEntry)
		for _, entry := range global {
			if err := s.admit(entry, false); err != nil {
				return nil, err
			}
		}
	}
	var pending []*authenticodeStoreEntry
	duplicates := make(authenticodeDuplicateIndex)
	for _, cert := range certs {
		if cert == nil {
			continue
		}
		entry, err := storeCertificateEntry(cert)
		if err != nil {
			s.uncertain = true
			continue
		}
		if duplicates.discard(cert) {
			continue
		}
		pending = append(pending, entry)
	}
	for l, r := 0, len(pending)-1; l < r; l, r = l+1, r-1 {
		pending[l], pending[r] = pending[r], pending[l]
	}
	remaining := pending[:0]
	blockedCatalog := false
	for _, entry := range pending {
		if _, blocked := d.blockedCertificate(entry.identity); blocked {
			if !catalog {
				return s, errRejectedAuthenticodeCertificate
			}
			// Catalog admission completes the independent direct-approval
			// pass before rejecting this bag. Never learn the blocked entry
			// itself, and do not run chained promotions after this pass.
			blockedCatalog = true
			continue
		}
		// CRB defines the anchor exponent as 65537. Modulus-only identity
		// matching must not authorize substituting an embedded exponent (in
		// particular E=1). A different exponent needs an actual certificate
		// signature under an already authorized key through the normal path.
		if !d.trustedCertificate(entry.identity) || entry.key.E.Cmp(big.NewInt(65537)) != 0 {
			remaining = append(remaining, entry)
			continue
		}
		entry.path = []*x509.Certificate{entry.cert}
		if err := s.admit(entry, true); err != nil {
			return nil, err
		}
	}
	if blockedCatalog {
		return s, errRejectedAuthenticodeCertificate
	}
	pending = remaining
	for i := 0; i < len(pending); {
		entry := pending[i]
		duplicate, err := s.duplicate(entry)
		if err != nil {
			return nil, err
		}
		if duplicate {
			pending = append(pending[:i], pending[i+1:]...)
			continue
		}
		parent, err := s.parent(entry)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			i++
			continue
		}
		if len(parent.path) >= 64 {
			return nil, errAuthenticodeLimit
		}
		entry.identity.codeSign = entry.identity.codeSign && parent.identity.codeSign
		entry.identity.timeSign = entry.identity.timeSign && parent.identity.timeSign
		entry.path = append([]*x509.Certificate{entry.cert}, parent.path...)
		if err := s.admit(entry, true); err != nil {
			return nil, err
		}
		pending = append(pending[:i], pending[i+1:]...)
		i = 0
	}
	return s, nil
}

func (s *authenticodeTrustStore) primarySigner(packet *authenticodeSignedData) (*x509.Certificate, error) {
	if packet == nil || packet.primary.hash == 0 || len(packet.primary.attributes) < 2 {
		return nil, errInvalidAuthenticode
	}
	checked := 0
	digest := authenticodeSignatureDigest(packet.primary.hash, packet.primary.attributes)
	for _, entry := range s.entries {
		if entry.cert == nil || !entry.identity.codeSign || entry.issuer != packet.primary.issuer || entry.identity.serial != packet.primary.serial {
			continue
		}
		checked++
		if checked > 64 {
			return nil, errAuthenticodeLimit
		}
		if err := s.chargeRSA(entry.key); err != nil {
			return nil, err
		}
		if verifyAuthenticodeRSA(entry.key, packet.primary.hash, digest, packet.primary.signature) {
			return entry.cert, nil
		}
	}
	if s.uncertain {
		return nil, errInvalidAuthenticode
	}
	return nil, nil
}
