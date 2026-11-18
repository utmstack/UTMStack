package event

import "strings"

// vendorToken is the upstream feed vendor's name as it appears in a few of the
// feed's own detection names; its test signature is "Clamav.Test.File-6".
// vendorReplacement is what an event carries instead.
const (
	vendorToken       = "clamav"
	vendorReplacement = "UTMStack"
)

// SanitizeName keeps the upstream feed vendor's name out of a detection or
// signature name placed in an event (branding rule). Names come from the feed
// as written, and must otherwise reach the platform unchanged, so the rule is
// deliberately narrow:
//
//   - the token "clamav" is matched ignoring ASCII case;
//   - it must begin a name component: at the start of the name or right after
//     a '.' (the feed's component separator) or a ':' (the separator of a
//     "file_watcher:" style source prefix);
//   - it must end the component or be followed by punctuation such as '-' or
//     '_' ("Clamav-Test-File"). A following letter, digit or non-ASCII byte
//     means a different word, which is left alone.
//
// So "Clamav.Test.File-6" becomes "UTMStack.Test.File-6", while a family name
// that merely contains the letters, such as "Win.Trojan.Clampi-43",
// "Win.Trojan.Clamavir-1" or "Win.Trojan.MyClamav-1", is kept exactly.
func SanitizeName(name string) string {
	n := len(vendorToken)
	var b strings.Builder
	last := 0
	for i := 0; i+n <= len(name); i++ {
		if i > 0 && name[i-1] != '.' && name[i-1] != ':' {
			continue
		}
		end := i + n
		if !equalFoldASCII(name[i:end], vendorToken) {
			continue
		}
		if end < len(name) && isWordByte(name[end]) {
			continue
		}
		if b.Len() == 0 {
			b.Grow(len(name) + len(vendorReplacement))
		}
		b.WriteString(name[last:i])
		b.WriteString(vendorReplacement)
		last = end
		i = end - 1
	}
	if last == 0 {
		return name
	}
	b.WriteString(name[last:])
	return b.String()
}

// equalFoldASCII reports whether s equals the lower-case ASCII word lower,
// ignoring ASCII case only. Non-ASCII bytes never match.
func equalFoldASCII(s, lower string) bool {
	if len(s) != len(lower) {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != lower[i] {
			return false
		}
	}
	return true
}

// isWordByte reports whether c continues a word: an ASCII letter or digit, or
// any byte of a multi-byte UTF-8 character.
func isWordByte(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c >= 0x80
}
