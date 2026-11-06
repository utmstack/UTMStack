package usecase

import (
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"

	"github.com/utmstack/utmstack/backend/internal/mail/domain"
)

// smtpAuth must agree with toEmailConfig (backend/modules/appconfig/usecase/config.go):
// the settings page stores the chosen encryption ("TLS"/"SSL"/"NONE") under the
// SmtpAuth field, not a literal "true"/"false". ACC-F1: the real send path used to
// require the literal string "true", which never happened, so production mail was
// never authenticated even though the config-test button (a different code path)
// reported success.
func TestSmtpAuth(t *testing.T) {
	cases := []struct {
		name     string
		smtpAuth string
		username string
		wantAuth bool
	}{
		{"TLS enables auth", "TLS", "user", true},
		{"SSL enables auth", "SSL", "user", true},
		{"lowercase tls enables auth", "tls", "user", true},
		{"literal true still enables auth", "true", "user", true},
		{"NONE disables auth", "NONE", "user", false},
		{"false disables auth", "false", "user", false},
		{"empty disables auth", "", "user", false},
		{"no username disables auth even with TLS", "TLS", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := &domain.EmailConfig{SmtpAuth: c.smtpAuth, Username: c.username, Password: "pw", Host: "smtp.example.com"}
			got := smtpAuth(cfg) != nil
			if got != c.wantAuth {
				t.Fatalf("smtpAuth(%q, username=%q) auth=%v, want %v", c.smtpAuth, c.username, got, c.wantAuth)
			}
		})
	}
}

// ACC-F29: invitation/reset links carry a literal "=" (the query separator,
// plus any base64 padding in the token) inside an HTML body long enough to
// cross quoted-printable's 76-column soft-wrap. Decoding used to lose "="
// characters that landed on a wrap boundary; base64 has no such escape
// character, so the link must come back byte-for-byte regardless of length.
func TestBuildMessagePreservesLinkEqualsSigns(t *testing.T) {
	link := "https://siem.example.com/reset/finish?key=abcDEF123%3D%3D"
	// Pad the body well past 76 columns before and after the link, the way a
	// real invitation/reset email does, so a quoted-printable soft line break
	// would have landed inside or right next to the "=" characters.
	body := "<!DOCTYPE html><html><body>" +
		"<p>" + strings.Repeat("Dear User, this is a long paragraph that pushes the link past the seventy six column soft wrap boundary. ", 2) + "</p>" +
		`<p><a href="` + link + `">` + link + `</a></p>` +
		"</body></html>"

	raw, err := buildMessage(&domain.EmailConfig{From: "alerts@example.com"}, []string{"user@example.com"}, nil, "subject", body, nil)
	if err != nil {
		t.Fatalf("buildMessage: %v", err)
	}

	msg, err := mail.ReadMessage(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("parse message: %v", err)
	}
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		t.Fatalf("Content-Type = %q (%v)", msg.Header.Get("Content-Type"), err)
	}
	mr := multipart.NewReader(msg.Body, params["boundary"])
	part, err := mr.NextPart()
	if err != nil {
		t.Fatalf("read body part: %v", err)
	}
	if cte := part.Header.Get("Content-Transfer-Encoding"); !strings.EqualFold(cte, "base64") {
		t.Fatalf("Content-Transfer-Encoding = %q, want base64", cte)
	}
	raw64, err := io.ReadAll(part)
	if err != nil {
		t.Fatalf("read part: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(string(raw64), "\r\n", ""))
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	if !strings.Contains(string(decoded), link) {
		t.Fatalf("decoded body does not contain the link intact.\nwant substring: %s\ngot: %s", link, decoded)
	}
}
