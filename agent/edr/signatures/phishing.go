package signatures

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"regexp/syntax"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

// ErrHTMLIncomplete means a relevant link could not be analyzed with the
// implemented URL semantics. Callers must not convert it to a clean verdict.
var ErrHTMLIncomplete = errors.New("UTMStack EDR HTML link analysis incomplete")

const maxPhishingURL = 8192
const maxPhishingRegex = 2048

type phishingRegex struct {
	pattern *regexp.Regexp
	cost    int
}
type phishingAllow struct {
	real, display string
	// Static M patterns with a terminal percent escape or trailing dot must
	// compare against the reference's cleaned, still-literal host spelling.
	realRaw, displayRaw bool
	expression          *phishingRegex
}
type phishingRules struct {
	hosts          map[string]bool
	expressions    []phishingRegex
	allowed        []phishingAllow
	hashes         map[[32]byte]byte
	hashDetections int
	hostPrefixes   map[[4]byte]struct{}
	// unsupported are allow rules whose URL semantics this engine does not
	// implement. Each keeps literals that any URL pair it allows must hold.
	unsupported []approxAllow
}

// errPhishingSemantics marks allow-rule syntax the reference accepts and
// this engine cannot interpret exactly.
var errPhishingSemantics = errors.New("phishing rule semantics not implemented")

func (p phishingRules) count() int { return len(p.hosts) + len(p.expressions) + p.hashDetections }

// PDB/WDB rules are POSIX extended regexes, not PCRE. The regexp POSIX parser
// rejects lookaround, backreferences and other unsupported constructs. ASCII
// inputs avoid Unicode versus byte-oriented POSIX matching discrepancies.
func compilePhishingRegex(body string) (phishingRegex, error) {
	return compilePhishingRegexMode(body, false)
}

// X rules match the cleaned real:display URL pair, including its schemes.
// ClamAV appends '/' to every regex and removes a final optional path form
// before compiling X rules (regex_list.c regex_list_add_pattern). A regexp
// without a fixed start is allowed to match later in that pair.
func compilePhishingAllowRegex(body string) (phishingRegex, error) {
	return compilePhishingRegexMode(body, true)
}

func compilePhishingRegexMode(body string, allowURL bool) (phishingRegex, error) {
	result := phishingRegex{}
	if body == "" || len(body) > maxPhishingRegex {
		return result, errors.New("empty or oversized phishing regular expression")
	}
	inClass := false
	for i := 0; i < len(body); i++ {
		b := body[i]
		if b == '[' {
			inClass = true
		}
		if b == ']' {
			inClass = false
		}
		if ((b == '/' && !inClass) || b == '@') && !allowURL {
			return result, fmt.Errorf("explicit URL paths, schemes and userinfo in phishing regexes are unsupported: %w", errPhishingSemantics)
		}
		if b < 32 || b > 126 {
			return result, errors.New("phishing regex must use printable ASCII")
		}
		if b == '\\' {
			i++
			if i >= len(body) {
				return result, errors.New("truncated phishing regex escape")
			}
			if !strings.ContainsRune(`\.^$|?*+()[]{}-/:`, rune(body[i])) {
				return result, errors.New("unsupported phishing regex escape")
			}
		}
	}
	// The R-rule subset below is kept under its earlier conservative anchors;
	// relocating an explicit R anchor into that wrapper changes its meaning.
	// X rules instead use the reference's unwrapped URL-pair expression.
	bodyTree, parseErr := syntax.Parse(body, syntax.POSIX)
	if parseErr != nil {
		return result, fmt.Errorf("unsupported POSIX phishing regex: %w", parseErr)
	}
	var hasAnchor func(*syntax.Regexp) bool
	hasAnchor = func(r *syntax.Regexp) bool {
		if r.Op == syntax.OpBeginLine || r.Op == syntax.OpEndLine || r.Op == syntax.OpBeginText || r.Op == syntax.OpEndText {
			return true
		}
		for _, sub := range r.Sub {
			if hasAnchor(sub) {
				return true
			}
		}
		return false
	}
	if hasAnchor(bodyTree) && !allowURL {
		return result, fmt.Errorf("explicit phishing regex anchors are unsupported; anchoring is implicit: %w", errPhishingSemantics)
	}
	expression := "^(" + body + "/)$"
	if allowURL {
		expression = body + "/"
		for _, end := range []string{"([/?].*)?/", "([/?].*)/"} {
			if strings.HasSuffix(expression, end) {
				expression = strings.TrimSuffix(expression, end) + "/"
				break
			}
		}
	}
	parsed, err := syntax.Parse(expression, syntax.POSIX)
	if err != nil {
		return result, fmt.Errorf("unsupported POSIX phishing regex: %w", err)
	}
	program, err := syntax.Compile(parsed.Simplify())
	if err != nil {
		return result, err
	}
	if len(program.Inst) > 4096 {
		return result, errors.New("phishing regex instruction limit exceeded")
	}
	result.pattern, err = regexp.CompilePOSIX(expression)
	result.cost = len(program.Inst)
	return result, err
}
func (p phishingRegex) matches(s string, w *matchWork) (bool, error) {
	if err := w.ctx.Err(); err != nil {
		return false, err
	}
	if err := w.step(int64((len(s) + 1) * p.cost)); err != nil {
		return false, err
	}
	matched := p.pattern.MatchString(s)
	if err := w.ctx.Err(); err != nil {
		return false, err
	}
	return matched, nil
}

// addPhishing compiles one .pdb or .wdb line. A level suffix is removed from
// the whole line first (phishingLevel); ignore lists never apply here. The
// pinned 1.5.4 loader uses only the first type byte and ignores any flag-mask
// text before the first ':'. Its Y allow rules take the same path as X rules.
func (d *Database) addPhishing(ext, line string) error {
	line, err := phishingLevel(line, d.functionality())
	if err != nil {
		return err
	}
	colon := strings.IndexByte(line, ':')
	if colon < 1 || colon+1 >= len(line) {
		return errors.New("unsupported phishing signature type or flag mask")
	}
	kind := line[0]
	body := line[colon+1:]
	if ext == ".pdb" {
		switch kind {
		case 'H':
			host, err := signatureHost(body)
			if err != nil {
				return err
			}
			if d.phishing.hosts == nil {
				d.phishing.hosts = make(map[string]bool)
			}
			d.phishing.hosts[host] = true
		case 'R':
			r, err := compilePhishingRegex(body)
			if err != nil {
				return err
			}
			d.phishing.expressions = append(d.phishing.expressions, r)
		default:
			return errors.New("unsupported protected-domain signature type")
		}
	} else {
		switch kind {
		case 'M':
			fields := strings.Split(body, ":")
			if len(fields) != 2 {
				return errors.New("phishing host allowlist requires real and displayed host")
			}
			// M patterns are static strings: the reference never interprets
			// backslash escapes in them. Every phishing link enters with
			// CLEANUP_URL, which changes every backslash to a slash before M
			// matching, so a pattern containing '\\' cannot match.
			if strings.Contains(fields[0], "\\") || strings.Contains(fields[1], "\\") {
				return nil
			}
			real, realRaw, err := staticAllowHost(fields[0])
			if err == nil {
				var display string
				var displayRaw bool
				if display, displayRaw, err = staticAllowHost(fields[1]); err == nil {
					d.phishing.allowed = append(d.phishing.allowed, phishingAllow{real: real, display: display, realRaw: realRaw, displayRaw: displayRaw})
					return nil
				}
			}
			if !errors.Is(err, errPhishingSemantics) {
				return err
			}
			d.phishing.unsupported = append(d.phishing.unsupported, approxAllow{
				real: []string{looseHost(fields[0])}, display: []string{looseHost(fields[1])}})
			d.unknown(UnknownPhishingAllowlist, 1)
		case 'X', 'Y':
			if !strings.Contains(body, ":") {
				return errors.New("phishing regex allowlist requires real:displayed expression")
			}
			r, err := compilePhishingAllowRegex(body)
			if err == nil {
				d.phishing.allowed = append(d.phishing.allowed, phishingAllow{expression: &r})
				return nil
			}
			if !errors.Is(err, errPhishingSemantics) {
				return err
			}
			approx, err := approximateAllow(body)
			if err != nil {
				return err
			}
			d.phishing.unsupported = append(d.phishing.unsupported, approx)
			d.unknown(UnknownPhishingAllowlist, 1)
		default:
			return errors.New("unsupported phishing allowlist signature type")
		}
	}
	return nil
}
func signatureHost(raw string) (string, error) {
	if raw != strings.ToLower(raw) || strings.ContainsAny(raw, "/:?#@%\\ \t\r\n") || strings.HasSuffix(raw, ".") {
		return "", fmt.Errorf("phishing hostname rules require canonical lowercase ASCII hostnames: %w", errPhishingSemantics)
	}
	for _, r := range raw {
		if r > 127 {
			return "", errors.New("phishing hostname rules require ASCII (use IDNA A-labels)")
		}
	}
	host, err := canonicalHost(raw)
	if err != nil || host != raw {
		return "", errors.New("invalid phishing signature hostname")
	}
	if net.ParseIP(host) != nil {
		return "", errors.New("IP-address protected-domain rules are unsupported")
	}
	if !strings.Contains(host, ".") {
		return "", errors.New("phishing hostname must have at least two labels")
	}
	return host, nil
}

// ClamAV keeps a final %xx in a cleaned hostname because its percent decoder
// stops before the last two characters (phishcheck.c str_hex_to_char). It also
// preserves a trailing dot. Accept those exact static spellings when the base
// hostname is otherwise canonical; all other noncanonical M forms remain
// conditional unknowns.
func staticAllowHost(raw string) (string, bool, error) {
	host, originalErr := signatureHost(raw)
	if originalErr == nil {
		return host, false, nil
	}
	base := ""
	switch {
	case strings.HasSuffix(raw, "."):
		base = strings.TrimSuffix(raw, ".")
	case len(raw) >= 3 && raw[len(raw)-3] == '%' && hexNibble(raw[len(raw)-2]) >= 0 && hexNibble(raw[len(raw)-1]) >= 0:
		base = raw[:len(raw)-3]
	}
	if base != "" {
		if _, err := signatureHost(base); err == nil {
			return raw, true, nil
		}
	}
	return "", false, originalErr
}
func suffixHost(host, suffix string) bool {
	return host == suffix || strings.HasSuffix(host, "."+suffix)
}
func sameSite(a, b string) bool {
	if a == b {
		return true
	}
	// ClamAV keeps a terminal root dot in its cleaned host and treats it as
	// different from the same spelling without that dot.
	if strings.HasSuffix(a, ".") != strings.HasSuffix(b, ".") {
		return false
	}
	if net.ParseIP(a) != nil || net.ParseIP(b) != nil {
		return false
	}
	first, e1 := publicsuffix.EffectiveTLDPlusOne(a)
	second, e2 := publicsuffix.EffectiveTLDPlusOne(b)
	return e1 == nil && e2 == nil && first == second
}

type phishingURL struct {
	host, allowHost, full, scheme string
	parsed                        *url.URL
}

func canonicalHost(raw string) (string, error) {
	if raw == "" || len(raw) > 253 {
		return "", errors.New("invalid URL hostname length")
	}
	if strings.Contains(raw, "%") {
		decoded, err := url.PathUnescape(raw)
		if err != nil || strings.ContainsAny(decoded, "%/@?#\\:\x00") {
			return "", errors.New("ambiguous encoded URL hostname")
		}
		raw = decoded
	}
	raw = strings.TrimSuffix(raw, ".")
	if address := net.ParseIP(raw); address != nil {
		return address.String(), nil
	}
	host, err := idna.Lookup.ToASCII(raw)
	if err != nil {
		return "", errors.New("invalid IDNA hostname")
	}
	host = strings.ToLower(host)
	if len(host) > 253 || strings.HasSuffix(host, ".") {
		return "", errors.New("invalid URL hostname")
	}
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", errors.New("invalid hostname label")
		}
		for _, b := range []byte(label) {
			if b != '-' && (b < 'a' || b > 'z') && (b < '0' || b > '9') {
				return "", errors.New("invalid hostname character")
			}
		}
	}
	// URL parsers disagree on legacy integer, octal, hexadecimal and shortened
	// IPv4 forms. Do not treat such hosts as ordinary DNS names.
	last := labels[len(labels)-1]
	if _, err := strconv.ParseUint(last, 0, 64); err == nil {
		return "", errors.New("unsupported legacy numeric IP representation")
	}
	numeric := true
	for _, b := range []byte(last) {
		numeric = numeric && b >= '0' && b <= '9'
	}
	if numeric {
		return "", errors.New("invalid numeric URL host")
	}
	return host, nil
}
func compactURL(raw string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\r', '\n', '\f':
			return -1
		}
		return r
	}, raw)
}

// ClamAV's isURL accepts a bare displayed hostname ending in an unknown TLD
// when an earlier label is a known TLD. It terminates the displayed URL at
// the last such label before matching protected domains. Keep this behavior
// scoped to plain host text; an explicit URL scheme takes a different path.
func truncateBareDisplayHost(raw string) string {
	if strings.Contains(raw, "://") || strings.HasPrefix(raw, "//") || strings.Contains(raw, "@") {
		return raw
	}
	host := raw
	if i := strings.IndexAny(host, "/:?#"); i >= 0 {
		host = host[:i]
	}
	if !strings.Contains(host, ".") {
		return raw
	}
	for _, c := range host {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '-') {
			return raw
		}
	}
	labels := strings.Split(host, ".")
	if len(labels) < 3 {
		return raw
	}
	if _, icann := publicsuffix.PublicSuffix(strings.ToLower(labels[len(labels)-1])); icann {
		return raw
	}
	end, last := 0, 0
	for i := 0; i < len(labels)-1; i++ {
		end += len(labels[i])
		if _, icann := publicsuffix.PublicSuffix(strings.ToLower(labels[i])); icann && i > 0 {
			last = end
		}
		end++
	}
	if last != 0 {
		return host[:last]
	}
	return raw
}
func normalizePhishingURL(raw string, base *url.URL, display bool) (phishingURL, error) {
	result := phishingURL{}
	if len(raw) > maxPhishingURL || !utf8.ValidString(raw) {
		return result, errors.New("oversized or invalid UTF-8 URL")
	}
	raw = compactURL(raw)
	if raw == "" {
		return result, errors.New("empty URL")
	}
	for _, r := range raw {
		if r < 32 || r == 127 {
			return result, errors.New("control character in URL")
		}
	}
	raw = strings.ReplaceAll(raw, "\\", "/")
	if display {
		raw = truncateBareDisplayHost(raw)
	}
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "http:") && !strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https:") && !strings.HasPrefix(lower, "https://") {
		return result, errors.New("ambiguous HTTP URL form")
	}
	// Decode only the authority before net/url, which otherwise rejects encoded
	// host bytes. Delimiter escapes remain errors instead of changing authority.
	prefix := 0
	if i := strings.Index(raw, "://"); i >= 0 {
		prefix = i + 3
	} else if strings.HasPrefix(raw, "//") {
		prefix = 2
	}
	rawAuthorityHost := ""
	if prefix > 0 {
		end := len(raw)
		if i := strings.IndexAny(raw[prefix:], "/?#"); i >= 0 {
			end = prefix + i
		}
		authority := raw[prefix:end]
		rawAuthorityHost = authority
		if i := strings.LastIndexByte(rawAuthorityHost, '@'); i >= 0 {
			rawAuthorityHost = rawAuthorityHost[i+1:]
		}
		if !strings.HasPrefix(rawAuthorityHost, "[") {
			if i := strings.LastIndexByte(rawAuthorityHost, ':'); i >= 0 && isNumber(rawAuthorityHost[i+1:]) {
				rawAuthorityHost = rawAuthorityHost[:i]
			}
		}
		rawAuthorityHost = strings.ToLower(rawAuthorityHost)
		if strings.Contains(authority, "%") {
			decoded, err := url.PathUnescape(authority)
			if err != nil || strings.ContainsAny(decoded, "%/?#\\\x00") {
				return result, errors.New("ambiguous URL authority escape")
			}
			if strings.Count(decoded, "@") != strings.Count(authority, "@") || strings.Count(decoded, ":") != strings.Count(authority, ":") {
				return result, errors.New("encoded URL authority delimiter")
			}
			raw = raw[:prefix] + decoded + raw[end:]
		}
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return result, errors.New("invalid URL")
	}
	if parsed.Scheme == "" && parsed.Host == "" {
		if base != nil && !display {
			parsed = base.ResolveReference(parsed)
		} else {
			// Bare displayed hostnames are valid; relative real links need an origin.
			first := raw
			if i := strings.IndexAny(first, "/?#"); i >= 0 {
				first = first[:i]
			}
			if strings.HasPrefix(raw, "/") || !strings.Contains(first, ".") {
				return result, errors.New("relative URL without an origin")
			}
			parsed, err = url.Parse("http://" + raw)
			if err != nil {
				return result, errors.New("invalid bare hostname URL")
			}
		}
	} else if parsed.Scheme == "" && parsed.Host != "" {
		parsed.Scheme = "http"
		if base != nil {
			parsed.Scheme = base.Scheme
		}
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" && scheme != "ftp" {
		return result, errors.New("unsupported URL scheme")
	}
	originalParsedHost := parsed.Hostname()
	host, err := canonicalHost(originalParsedHost)
	if err != nil {
		return result, err
	}
	if display && !strings.Contains(host, ".") && net.ParseIP(host) == nil {
		return result, errors.New("display text is not a public hostname")
	}
	port := parsed.Port()
	if port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return result, errors.New("invalid URL port")
		}
	}
	hostname := host
	if strings.Contains(host, ":") {
		hostname = "[" + host + "]"
	}
	if port != "" {
		hostname += ":" + port
	}
	parsed.Scheme = scheme
	parsed.Host = hostname
	// Paths and queries retain escaped delimiters. The POSIX matcher receives an
	// ASCII URI, avoiding byte/rune semantics changing a regex match.
	parsed.Fragment = ""
	parsed.Path = ""
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	// X-type allow rules inspect the whole cleaned URL pair before the
	// reference extracts hostnames. Keep userinfo in that representation;
	// parsed.Hostname still gives the post-@ hostname for domain checks.
	full := parsed.String()
	for _, r := range full {
		if r > 127 {
			return result, errors.New("non-ASCII URL path requires unsupported normalization")
		}
	}
	allowHost := host
	if strings.HasSuffix(originalParsedHost, ".") {
		allowHost = strings.ToLower(originalParsedHost)
	}
	if len(rawAuthorityHost) >= 3 && rawAuthorityHost[len(rawAuthorityHost)-3] == '%' && hexNibble(rawAuthorityHost[len(rawAuthorityHost)-2]) >= 0 && hexNibble(rawAuthorityHost[len(rawAuthorityHost)-1]) >= 0 {
		allowHost = rawAuthorityHost
	}
	result = phishingURL{host: host, allowHost: allowHost, scheme: scheme, full: full, parsed: parsed}
	return result, nil
}
func (d *Database) protected(real, display phishingURL, w *matchWork) (bool, error) {
	for host := display.host; host != ""; {
		if d.phishing.hosts[host] {
			return true, nil
		}
		i := strings.IndexByte(host, '.')
		if i < 0 {
			break
		}
		host = host[i+1:]
	}
	for _, r := range d.phishing.expressions {
		match, err := r.matches(real.full+":"+display.full+"/", w)
		if err != nil || match {
			return match, err
		}
		match, err = r.matches(display.host+"/", w)
		if err != nil || match {
			return match, err
		}
	}
	return false, nil
}
func (d *Database) phishingAllowed(real, display phishingURL, w *matchWork) (bool, error) {
	for _, rule := range d.phishing.allowed {
		if err := w.step(1); err != nil {
			return false, err
		}
		if rule.expression != nil {
			match, err := rule.expression.matches(real.full+":"+display.full+"/", w)
			if err != nil || match {
				return match, err
			}
		} else {
			realHost, displayHost := real.host, display.host
			if rule.realRaw {
				realHost = real.allowHost
			}
			if rule.displayRaw {
				displayHost = display.allowHost
			}
			if suffixHost(realHost, rule.real) && suffixHost(displayHost, rule.display) {
				return true, nil
			}
		}
	}
	return false, nil
}

// MatchHTML examines HTML link destination/display pairs using PDB/WDB/GDB rules.
// Invoke it on the original HTML before entity/text normalization and after the
// whole-object Allowed check. No page is fetched and no script is executed.
// Raw byte matching remains a separate call to Match.
func (d *Database) MatchHTML(ctx context.Context, data []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if d.phishing.count() == 0 {
		return "", nil
	}
	pairs, baseRaw, err := extractHTMLLinks(ctx, data, len(d.phishing.hashes) != 0)
	if err != nil {
		return "", err
	}
	w := &matchWork{ctx: ctx, limit: d.limits.matchLimit(len(data))}
	var base *url.URL
	var incomplete error
	if baseRaw != "" {
		normalized, e := normalizePhishingURL(baseRaw, nil, false)
		if e == nil {
			base = normalized.parsed
		}
		// A malformed base is irrelevant to absolute links and to a page
		// without links. Relative links that need it remain incomplete below
		// when their destination cannot be resolved.
	}
	for _, pair := range pairs {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if len(d.phishing.hashes) != 0 {
			verdict, hashErr := d.phishing.matchSafeBrowsingURL(pair.real, w)
			if hashErr != nil {
				if errors.Is(hashErr, ErrHTMLIncomplete) {
					incomplete = hashErr
				} else {
					return "", hashErr
				}
			} else if verdict != "" {
				return verdict, nil
			}
		}
		if pair.display == "" {
			continue
		}
		display, e := normalizePhishingURL(pair.display, nil, true)
		if e != nil {
			candidate := strings.ToLower(compactURL(pair.display))
			if strings.Contains(candidate, "://") || strings.HasPrefix(candidate, "www.") {
				incomplete = fmt.Errorf("%w: displayed URL: %v", ErrHTMLIncomplete, e)
			}
			continue
		} // ordinary prose is not a claimed URL
		real, e := normalizePhishingURL(pair.real, base, false)
		if e != nil {
			// Check the displayed host before marking irrelevant links incomplete.
			relevant, checkErr := d.protected(phishingURL{}, display, w)
			if checkErr != nil {
				return "", checkErr
			}
			if relevant {
				incomplete = fmt.Errorf("%w: %v", ErrHTMLIncomplete, e)
			}
			continue
		}
		allowed, e := d.phishingAllowed(real, display, w)
		if e != nil {
			return "", e
		}
		if allowed {
			continue
		}
		protected, e := d.protected(real, display, w)
		if e != nil {
			return "", e
		}
		if !protected {
			continue
		}
		if d.phishing.mayAllow(real, display) {
			// An allow rule this engine cannot interpret could apply here.
			incomplete = fmt.Errorf("%w: %w", ErrRuleIncomplete, unknownBecause(UnknownPhishingAllowlist))
			continue
		}
		if pair.ssl && display.scheme == "https" && real.scheme == "http" {
			return "Heuristics.Phishing.Email.SSL-Spoof", nil
		}
		realSite, displaySite := real.host, display.host
		if strings.HasSuffix(real.allowHost, ".") {
			realSite = real.allowHost
		}
		if strings.HasSuffix(display.allowHost, ".") {
			displaySite = display.allowHost
		}
		if !sameSite(realSite, displaySite) {
			return "Heuristics.Phishing.Email.SpoofedDomain", nil
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return "", incomplete
}

// approxAllow is an allow rule reduced to literals every URL pair it allows
// must contain, compared case-insensitively on the real and the displayed
// URL. Missing any of them proves the rule cannot allow the pair.
type approxAllow struct {
	real, display []string
}

func (p phishingRules) mayAllow(real, display phishingURL) bool {
	holds := func(url string, literals []string) bool {
		url = strings.ToLower(url)
		for _, l := range literals {
			if !strings.Contains(url, l) {
				return false
			}
		}
		return true
	}
	for _, a := range p.unsupported {
		// The rules match "real:display/", so the displayed URL ends in '/'.
		if holds(real.full, a.real) && holds(display.full+"/", a.display) {
			return true
		}
	}
	return false
}

// looseHost is the literal a non-canonical allowlist host still requires:
// percent escapes and a trailing dot removed, lowercased.
func looseHost(raw string) string {
	raw = strings.ToLower(raw)
	if i := strings.IndexByte(raw, '%'); i >= 0 {
		raw = raw[:i]
	}
	return strings.TrimRight(raw, ".")
}

// approximateAllow splits an allow regex at its top-level ':' into the real
// and displayed URL expressions and keeps the literals each requires.
func approximateAllow(body string) (approxAllow, error) {
	depth, inClass, split := 0, false, -1
	for i := 0; i < len(body); i++ {
		switch c := body[i]; {
		case c == '\\':
			i++
		case inClass:
			inClass = c != ']'
		case c == '[':
			inClass = true
		case c == '(':
			depth++
		case c == ')':
			depth--
		case c == ':' && depth == 0 && split < 0:
			split = i
		}
	}
	if split < 0 {
		return approxAllow{}, errors.New("phishing regex allowlist requires real:displayed expression")
	}
	real, err := requiredLiterals(body[:split])
	if err != nil {
		return approxAllow{}, err
	}
	display, err := requiredLiterals(body[split+1:])
	if err != nil {
		return approxAllow{}, err
	}
	return approxAllow{real: real, display: display}, nil
}

// requiredLiterals returns the literal runs of an expression's top-level
// concatenation, lowercased: text every match contains.
func requiredLiterals(expr string) ([]string, error) {
	tree, err := syntax.Parse(expr, syntax.Perl)
	if err != nil {
		return nil, fmt.Errorf("unsupported phishing regex: %w", err)
	}
	var out []string
	add := func(r *syntax.Regexp) {
		if r.Op == syntax.OpLiteral && len(r.Rune) > 0 {
			out = append(out, strings.ToLower(string(r.Rune)))
		}
	}
	if tree.Op == syntax.OpConcat {
		for _, sub := range tree.Sub {
			add(sub)
		}
	} else {
		add(tree)
	}
	return out, nil
}
