package signatures

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"testing"
)

// With EDR_OFFICIAL_SIGNATURE_DIR, every regex subsignature of the official
// feed must compile here unless it uses a construct this engine leaves
// unknown when reached; the test lists those.
//
// EDR_REGEX_FEED_CASES=<file> also writes oracle cases for each of them:
// subjects built from the pattern (strings it matches, some mutated, inside
// noise). Answered by the reference's library, they are checked through
// EDR_REGEX_ORACLE by TestRegexAgreesWithReferenceLibrary. The cases hold
// feed content, so they stay outside the repository.
func TestFeedRegexSubsignatures(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_SIGNATURE_DIR")
	if dir == "" {
		t.Skip("EDR_OFFICIAL_SIGNATURE_DIR not set")
	}
	d, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	type sub struct{ pattern, flags string }
	seen := map[sub]bool{}
	var subs []sub
	why := map[string]int{}
	compiled := 0
	for i := range d.patterns {
		h := d.patterns[i].holder
		if h == nil || h.reason != UnknownPCRE {
			continue
		}
		if h.regex == nil {
			why[h.regexErr]++
			continue
		}
		compiled++
		s := sub{h.regex.source, h.regex.flags}
		if !seen[s] {
			seen[s] = true
			subs = append(subs, s)
		}
	}
	var reasons []string
	for k := range why {
		reasons = append(reasons, k)
	}
	sort.Strings(reasons)
	for _, k := range reasons {
		t.Logf("%4d unknown when reached: %s", why[k], k)
	}
	t.Logf("%d regex subsignatures compiled, %d distinct patterns", compiled, len(subs))
	out := os.Getenv("EDR_REGEX_FEED_CASES")
	if out == "" {
		return
	}
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	bw := bufio.NewWriter(f)
	defer bw.Flush()
	r := rand.New(rand.NewSource(20260927))
	noise := func() []byte {
		b := make([]byte, r.Intn(24))
		for i := range b {
			b[i] = " \n\x00\x0baAzZ09._-\"'<>=;()\x80\xe9\xff\t"[r.Intn(25)]
		}
		return b
	}
	for _, s := range subs {
		p, err := parseRegex(s.pattern, s.flags)
		if err != nil {
			t.Fatal(err)
		}
		for k := 0; k < 30; k++ {
			var subject []byte
			for n := 1 + r.Intn(3); n > 0; n-- {
				subject = append(subject, noise()...)
				sample := rxSample(p.root, r, map[int][]byte{})
				if len(sample) > 0 && r.Intn(4) == 0 {
					sample[r.Intn(len(sample))] ^= byte(1 + r.Intn(255))
				}
				subject = append(subject, sample...)
			}
			subject = append(subject, noise()...)
			fmt.Fprintf(bw, "%s\t%s\t:0:%s\t0\t0\n", hex.EncodeToString([]byte(s.pattern)), s.flags, hex.EncodeToString(subject))
		}
	}
}

// rxSample builds a string n likely matches: a random member of each set, a
// few iterations of each repeat, one alternative, the group's text for a
// backreference. Lookaround and anchors are not satisfied on purpose.
func rxSample(n *rxNode, r *rand.Rand, caps map[int][]byte) []byte {
	pick := func(s *byteSet) []byte {
		var members []byte
		for c := 0; c < 256; c++ {
			if s.has(byte(c)) {
				members = append(members, byte(c))
			}
		}
		if len(members) == 0 {
			return nil
		}
		// Prefer printable members, as real text holds.
		for tries := 0; tries < 4; tries++ {
			c := members[r.Intn(len(members))]
			if c >= 32 && c < 127 || tries == 3 {
				return []byte{c}
			}
		}
		return nil
	}
	switch n.kind {
	case rxSet:
		return pick(&n.set)
	case rxString:
		var out []byte
		for i := range n.seq {
			out = append(out, pick(&n.seq[i])...)
		}
		return out
	case rxConcat:
		var out []byte
		for _, s := range n.subs {
			out = append(out, rxSample(s, r, caps)...)
		}
		return out
	case rxAlt:
		return rxSample(n.subs[r.Intn(len(n.subs))], r, caps)
	case rxRepeat:
		hi := n.max
		if hi < 0 || hi > n.min+3 {
			hi = n.min + 3
		}
		count := n.min + r.Intn(hi-n.min+1)
		var out []byte
		for i := 0; i < count; i++ {
			out = append(out, rxSample(n.subs[0], r, caps)...)
		}
		return out
	case rxGroup:
		out := rxSample(n.subs[0], r, caps)
		if n.cap > 0 {
			caps[n.cap] = out
		}
		return out
	case rxAtomic:
		return rxSample(n.subs[0], r, caps)
	case rxBackref:
		return append([]byte(nil), caps[n.cap]...)
	}
	return nil
}
