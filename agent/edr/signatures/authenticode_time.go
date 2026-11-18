package signatures

import (
	"encoding/asn1"
	"time"
)

// referenceAuthenticodeTime interprets only the local Authenticode wire policy.
// limit bounds the ASN.1 object in its logical parent; mapped retains the
// physically available bytes after it. The reference requires a 13/15-byte
// mapped window even for minute precision, then advances by the declared TLV
// length. This is independent of the time bytes used in the original signature.
func referenceAuthenticodeTime(mapped []byte, limit int) (at time.Time, consumed int, err error) {
	reject := func() (time.Time, int, error) { return time.Time{}, 0, errRejectedAuthenticodeCertificate }
	if limit < 0 || limit > len(mapped) {
		return reject()
	}
	value, _, e := asn1Object(mapped[:limit])
	if e != nil {
		return reject()
	}
	length, yearDigits := 13, 2
	if value.Class != asn1.ClassUniversal || value.IsCompound {
		return reject()
	}
	if value.Tag == asn1.TagGeneralizedTime {
		length, yearDigits = 15, 4
	} else if value.Tag != asn1.TagUTCTime {
		return reject()
	}
	header := len(value.FullBytes) - len(value.Bytes)
	if len(mapped)-header < length {
		return reject()
	}
	body := mapped[header : header+length]
	pair := func(at int) int {
		a, b := body[at], body[at+1]
		if a < '0' || a > '9' || b < '0' || b > '9' {
			return -1
		}
		return int(a-'0')*10 + int(b-'0')
	}
	year := pair(0)
	if year < 0 {
		return reject()
	}
	if yearDigits == 4 {
		// The pinned reader obtains both pairs from the same pointer. Its
		// local generalized-time policy maps 20xx to 2020, 21xx to 2121, etc.
		// The final two year bytes are skipped without being interpreted.
		year = year*100 + year
	} else if year >= 50 {
		year += 1900
	} else {
		year += 2000
	}
	month, day, hour, minute := pair(yearDigits), pair(yearDigits+2), pair(yearDigits+4), pair(yearDigits+6)
	if month < 1 || month > 12 || day < 1 || day > 31 || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return reject()
	}
	second := 0
	zone := yearDigits + 8
	if body[zone] != 'Z' {
		second = pair(zone)
		zone += 2
		if second < 0 || second > 59 || body[zone] != 'Z' {
			return reject()
		}
	}
	// The pinned UTC reference normalizes calendar overflow (for example Feb 31),
	// but does not accept numeric zones, fractions or leap seconds. Bytes after
	// the first expected Z are ignored; original TBS/attribute bytes remain signed.
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), len(value.FullBytes), nil
}

func normalizeCertificateValidity(wire *authenticodeCertificateWire, fields [][]byte) error {
	index := wire.algorithm + 2
	if index >= len(wire.fields) {
		return errRejectedAuthenticodeCertificate
	}
	validity := wire.fields[index]
	if !derTag(validity, 0, asn1.TagSequence) {
		return errRejectedAuthenticodeCertificate
	}
	offset := len(validity.FullBytes) - len(validity.Bytes)
	for i := 0; i < index; i++ {
		offset += len(wire.fields[i].FullBytes)
	}
	if offset > len(wire.tbs.Bytes) {
		return errRejectedAuthenticodeCertificate
	}
	mapped := wire.tbs.Bytes[offset:]
	before, n, err := referenceAuthenticodeTime(mapped, len(validity.Bytes))
	if err != nil {
		return err
	}
	after, m, err := referenceAuthenticodeTime(mapped[n:], len(validity.Bytes)-n)
	if err != nil || n+m != len(validity.Bytes) || !before.Before(after) {
		return errRejectedAuthenticodeCertificate
	}
	first, err := asn1.Marshal(before)
	if err != nil {
		return errRejectedAuthenticodeCertificate
	}
	last, err := asn1.Marshal(after)
	if err != nil {
		return errRejectedAuthenticodeCertificate
	}
	fields[index], err = certificateSequence(first, last)
	return err
}
