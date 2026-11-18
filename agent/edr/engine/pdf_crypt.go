package engine

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rc4"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"math/big"
)

// PDF encryption with the standard security handler (ISO 32000-1 7.6.3 and
// ISO 32000-2 7.6.4), for files a reader opens without a password: the
// file key is derived for an empty user password and checked against /U.
// Like the reference (ClamAV 1.4.3, checked on qpdf's encryptions with RC4
// 40 and 128 bits, AES 128 bits and AES 256 bits of revisions 5 and 6),
// streams are decrypted before their filters, and strings are not: a
// marker in an encrypted string is not found, unless the string is inside
// an object stream, which is decrypted as a whole.

// pdfPadding is the standard security handler's password padding.
var pdfPadding = []byte{
	0x28, 0xbf, 0x4e, 0x5e, 0x4e, 0x75, 0x8a, 0x41, 0x64, 0x00, 0x4e, 0x56, 0xff, 0xfa, 0x01, 0x08,
	0x2e, 0x2e, 0x00, 0xb6, 0xd0, 0x68, 0x3e, 0x80, 0x2f, 0x0c, 0xa9, 0xfe, 0x64, 0x53, 0x69, 0x7a,
}

type pdfCipher uint8

const (
	pdfIdentity pdfCipher = iota
	pdfRC4
	pdfAES128
	pdfAES256
)

// pdfCrypt is what decrypting a file's streams needs.
type pdfCrypt struct {
	key             []byte
	streams, files  pdfCipher // the stream and embedded-file crypt filters
	encryptMetadata bool
}

// newPDFCrypt reads an encryption dictionary and the first file identifier
// and derives the file key for an empty user password. ErrUnsupported means
// the file cannot be decrypted that way: another security handler, a form
// this engine does not decode, or a user password.
func newPDFCrypt(ps *pdfScan, dict map[string]pdfValue, id []byte) (*pdfCrypt, error) {
	integer := func(key string, missing int64) (int64, bool) {
		v, exists := dict[key]
		if !exists {
			return missing, true
		}
		v, err := ps.resolve(v)
		return v.number, err == nil && v.kind == 'i'
	}
	str := func(key string) []byte {
		v, err := ps.resolve(dict[key])
		if err != nil || v.kind != 's' {
			return nil
		}
		return v.str
	}
	filter, err := ps.resolve(dict["Filter"])
	if err != nil || filter.kind != 'n' || filter.name != "Standard" {
		return nil, ErrUnsupported
	}
	v, ok1 := integer("V", 0)
	r, ok2 := integer("R", 0)
	length, ok3 := integer("Length", 40)
	p, ok4 := integer("P", 0)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return nil, ErrMalformed
	}
	c := &pdfCrypt{encryptMetadata: true}
	if m, exists := dict["EncryptMetadata"]; exists {
		m, err := ps.resolve(m)
		if err != nil || m.kind != 'k' || m.name != "true" && m.name != "false" {
			return nil, ErrMalformed
		}
		c.encryptMetadata = m.name == "true"
	}
	o, u := str("O"), str("U")
	switch {
	case v == 1 || v == 2:
		if v == 1 {
			length = 40
		}
		if length < 40 || length > 128 || length%8 != 0 || r < 2 || r > 3 {
			return nil, ErrUnsupported
		}
		c.streams, c.files = pdfRC4, pdfRC4
	case v == 4:
		if r != 4 {
			return nil, ErrUnsupported
		}
		if _, exists := dict["Length"]; !exists {
			length = 128
		}
		if length != 128 {
			return nil, ErrUnsupported
		}
		if c.streams, err = ps.cryptFilter(dict, "StmF", "Identity"); err != nil {
			return nil, err
		}
		if c.files, err = ps.cryptFilter(dict, "EFF", "StmF"); err != nil {
			return nil, err
		}
	case v == 5:
		if r != 5 && r != 6 {
			return nil, ErrUnsupported
		}
		if c.streams, err = ps.cryptFilter(dict, "StmF", "Identity"); err != nil {
			return nil, err
		}
		if c.files, err = ps.cryptFilter(dict, "EFF", "StmF"); err != nil {
			return nil, err
		}
		key, ok := pdfKeyAES256(int(r), u, str("UE"))
		if !ok {
			return nil, ErrUnsupported
		}
		c.key = key
		return c, nil
	default:
		return nil, ErrUnsupported
	}
	if len(o) < 32 || len(u) < 32 {
		return nil, ErrMalformed
	}
	n := int(length / 8)
	h := md5.New()
	h.Write(pdfPadding)
	h.Write(o[:32])
	binary.Write(h, binary.LittleEndian, uint32(int32(p)))
	h.Write(id)
	if r >= 4 && !c.encryptMetadata {
		h.Write([]byte{0xff, 0xff, 0xff, 0xff})
	}
	key := h.Sum(nil)
	if r >= 3 {
		for i := 0; i < 50; i++ {
			sum := md5.Sum(key[:n])
			key = sum[:]
		}
	}
	key = key[:n]
	// The user password check (algorithms 4 and 5).
	var check []byte
	if r == 2 {
		check = make([]byte, 32)
		pdfRC4XOR(key, pdfPadding, check)
		if !bytes.Equal(check, u[:32]) {
			return nil, ErrUnsupported
		}
	} else {
		sum := md5.Sum(append(append([]byte(nil), pdfPadding...), id...))
		check = sum[:]
		k := make([]byte, n)
		for i := 0; i < 20; i++ {
			for j := range k {
				k[j] = key[j] ^ byte(i)
			}
			pdfRC4XOR(k, check, check)
		}
		if !bytes.Equal(check, u[:16]) {
			return nil, ErrUnsupported
		}
	}
	c.key = key
	return c, nil
}

// cryptFilter is the cipher of the crypt filter an encryption dictionary
// names under key (StmF, EFF), or under fallback when the key is absent.
func (ps *pdfScan) cryptFilter(dict map[string]pdfValue, key, fallback string) (pdfCipher, error) {
	v, exists := dict[key]
	if !exists {
		if fallback == "Identity" {
			return pdfIdentity, nil
		}
		return ps.cryptFilter(dict, fallback, "Identity")
	}
	name, err := ps.resolve(v)
	if err != nil || name.kind != 'n' {
		return 0, ErrMalformed
	}
	if name.name == "Identity" {
		return pdfIdentity, nil
	}
	filters, err := ps.resolve(dict["CF"])
	if err != nil || filters.kind != 'd' {
		return 0, ErrMalformed
	}
	filter, err := ps.resolve(filters.dict[name.name])
	if err != nil || filter.kind != 'd' {
		return 0, ErrMalformed
	}
	method, err := ps.resolve(filter.dict["CFM"])
	if err != nil || method.kind != 'n' {
		return 0, ErrUnsupported
	}
	switch method.name {
	case "None":
		return pdfIdentity, nil
	case "V2":
		return pdfRC4, nil
	case "AESV2":
		return pdfAES128, nil
	case "AESV3":
		return pdfAES256, nil
	}
	return 0, ErrUnsupported
}

// pdfKeyAES256 checks the empty user password against /U and decrypts the
// file key from /UE: revision 5 hashes with one SHA-256, revision 6 with
// the iterative hash of ISO 32000-2 (algorithm 2.B).
func pdfKeyAES256(r int, u, ue []byte) ([]byte, bool) {
	if len(u) < 48 || len(ue) < 32 {
		return nil, false
	}
	hash := func(salt []byte) []byte {
		if r == 5 {
			sum := sha256.Sum256(salt)
			return sum[:]
		}
		return pdfHash2B(nil, salt)
	}
	if !bytes.Equal(hash(u[32:40]), u[:32]) {
		return nil, false
	}
	block, err := aes.NewCipher(hash(u[40:48]))
	if err != nil {
		return nil, false
	}
	key := make([]byte, 32)
	cipher.NewCBCDecrypter(block, make([]byte, 16)).CryptBlocks(key, ue[:32])
	return key, true
}

// pdfHash2B is ISO 32000-2's revision 6 password hash for a user password
// (no user data).
func pdfHash2B(password, salt []byte) []byte {
	sum := sha256.Sum256(append(append([]byte(nil), password...), salt...))
	k := sum[:]
	for i := 0; ; i++ {
		round := append(append([]byte(nil), password...), k...)
		k1 := bytes.Repeat(round, 64)
		block, _ := aes.NewCipher(k[:16])
		e := make([]byte, len(k1))
		cipher.NewCBCEncrypter(block, k[16:32]).CryptBlocks(e, k1)
		switch new(big.Int).Mod(new(big.Int).SetBytes(e[:16]), big.NewInt(3)).Int64() {
		case 0:
			s := sha256.Sum256(e)
			k = s[:]
		case 1:
			s := sha512.Sum384(e)
			k = s[:]
		default:
			s := sha512.Sum512(e)
			k = s[:]
		}
		if i >= 63 && int(e[len(e)-1]) <= i-31 {
			return k[:32]
		}
	}
}

func pdfRC4XOR(key, src, dst []byte) {
	c, _ := rc4.NewCipher(key)
	c.XORKeyStream(dst, src)
}

// decrypt decrypts a stream of the object ref with the cipher.
func (c *pdfCrypt) decrypt(ref pdfRef, with pdfCipher, data []byte) ([]byte, error) {
	key := c.key
	if with == pdfRC4 || with == pdfAES128 {
		h := md5.New()
		h.Write(c.key)
		h.Write([]byte{byte(ref.object), byte(ref.object >> 8), byte(ref.object >> 16), byte(ref.generation), byte(ref.generation >> 8)})
		if with == pdfAES128 {
			h.Write([]byte("sAlT"))
		}
		key = h.Sum(nil)[:min(len(c.key)+5, 16)]
	}
	switch with {
	case pdfIdentity:
		return data, nil
	case pdfRC4:
		out := make([]byte, len(data))
		pdfRC4XOR(key, data, out)
		return out, nil
	}
	if len(data) == 0 {
		return data, nil
	}
	// AES: a 16-byte initialization vector, whole blocks, and PKCS#5
	// padding of 1 to 16 bytes.
	if len(data) < 32 || len(data)%16 != 0 {
		return nil, ErrMalformed
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrMalformed
	}
	out := make([]byte, len(data)-16)
	cipher.NewCBCDecrypter(block, data[:16]).CryptBlocks(out, data[16:])
	pad := int(out[len(out)-1])
	if pad < 1 || pad > 16 || !bytes.Equal(out[len(out)-pad:], bytes.Repeat([]byte{byte(pad)}, pad)) {
		return nil, ErrMalformed
	}
	return out[:len(out)-pad], nil
}
