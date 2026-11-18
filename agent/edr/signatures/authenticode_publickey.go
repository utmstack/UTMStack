package signatures

import (
	"bytes"
	"encoding/asn1"
	"math/big"
)

// certificateKeyObject reads the pinned reference's bounded definite-length
// objects. In particular, the RSA structure inside the BIT STRING can use
// nonminimal definite lengths; it is not passed to Go's DER integer parser.
func certificateKeyObject(data []byte, tag byte) (content, rest []byte, err error) {
	if len(data) < 2 || data[0] != tag {
		return nil, nil, errRejectedAuthenticodeCertificate
	}
	length := uint64(data[1])
	header := 2
	if length&0x80 != 0 {
		count := int(length & 0x7f)
		if count == 0 || count > 4 || count > len(data)-header {
			return nil, nil, errRejectedAuthenticodeCertificate
		}
		length = 0
		for _, b := range data[header : header+count] {
			length = length<<8 | uint64(b)
		}
		header += count
	}
	if length > uint64(len(data)-header) {
		return nil, nil, errRejectedAuthenticodeCertificate
	}
	end := header + int(length)
	return data[header:end], data[end:], nil
}

// normalizeCertificatePublicKey validates reference SPKI framing and produces
// only a metadata copy. N and E are unsigned byte integers in that format.
// Signed TBS bytes and the original SPKI stay unchanged on the returned cert.
func decodeAuthenticodePublicKey(data []byte) (*authenticodeRSAKey, error) {
	spki, tail, err := certificateKeyObject(data, 0x30)
	if err != nil || len(tail) != 0 {
		return nil, errRejectedAuthenticodeCertificate
	}
	algorithm, rest, err := certificateKeyObject(spki, 0x30)
	if err != nil {
		return nil, err
	}
	oid, parameters, err := certificateKeyObject(algorithm, 6)
	if err != nil {
		return nil, err
	}
	// rsaEncryption is the only public-key algorithm admitted by the reference.
	if !bytes.Equal(oid, []byte{0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 1, 1, 1}) {
		return nil, errRejectedAuthenticodeCertificate
	}
	if len(parameters) != 0 {
		value, tail, err := certificateKeyObject(parameters, 5)
		if err != nil || len(value) != 0 || len(tail) != 0 {
			return nil, errRejectedAuthenticodeCertificate
		}
	}
	bits, tail, err := certificateKeyObject(rest, 3)
	if err != nil || len(tail) != 0 || len(bits) == 0 || bits[0] != 0 {
		return nil, errRejectedAuthenticodeCertificate
	}
	key, tail, err := certificateKeyObject(bits[1:], 0x30)
	if err != nil || len(tail) != 0 {
		return nil, errRejectedAuthenticodeCertificate
	}
	modulus, rest, err := certificateKeyObject(key, 2)
	if err != nil || len(modulus) < 128 || len(modulus) > 513 {
		return nil, errRejectedAuthenticodeCertificate
	}
	exponent, tail, err := certificateKeyObject(rest, 2)
	if err != nil || len(tail) != 0 || len(exponent) == 0 || len(exponent) > len(modulus) {
		return nil, errRejectedAuthenticodeCertificate
	}
	return &authenticodeRSAKey{N: new(big.Int).SetBytes(modulus), E: new(big.Int).SetBytes(exponent)}, nil
}

func encodeAuthenticodeMetadataKey(key *authenticodeRSAKey) ([]byte, error) {
	encodedKey, err := asn1.Marshal(struct {
		N, E *big.Int
	}{key.N, key.E})
	if err != nil {
		return nil, err
	}
	encodedOID, err := asn1.Marshal(oidAuthenticodeRSA)
	if err != nil {
		return nil, err
	}
	encodedAlgorithm, err := certificateSequence(encodedOID, []byte{5, 0})
	if err != nil {
		return nil, err
	}
	encodedBits, err := asn1.Marshal(asn1.BitString{Bytes: encodedKey, BitLength: len(encodedKey) * 8})
	if err != nil {
		return nil, err
	}
	return certificateSequence(encodedAlgorithm, encodedBits)
}

func normalizeCertificatePublicKey(data []byte) ([]byte, error) {
	key, err := decodeAuthenticodePublicKey(data)
	if err != nil {
		return nil, err
	}
	return encodeAuthenticodeMetadataKey(key)
}
