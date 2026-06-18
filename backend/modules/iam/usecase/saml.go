package usecase

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/jwt"
)

// samlUsecase implements the live SP-initiated SAML2 web SSO flow on top of
// github.com/crewjam/saml. ServiceProviders are built per request from the DB
// config (key decrypted with the cipher, IdP metadata fetched live). There is
// no JIT provisioning: the NameID must match the login of a pre-existing,
// activated local user.
type samlUsecase struct {
	idpRepo     connectors.IdentityProviderRepository
	userRepo    connectors.UserRepository
	refreshRepo connectors.RefreshTokenRepository
	signer      *jwt.Signer
	cipher      Cipher
	refreshTTL  time.Duration
	httpClient  *http.Client
}

func NewSAMLUsecase(
	idpRepo connectors.IdentityProviderRepository,
	userRepo connectors.UserRepository,
	refreshRepo connectors.RefreshTokenRepository,
	signer *jwt.Signer,
	cipher Cipher,
	refreshTTL time.Duration,
) connectors.SAMLUsecase {
	return &samlUsecase{
		idpRepo:     idpRepo,
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		signer:      signer,
		cipher:      cipher,
		refreshTTL:  refreshTTL,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (u *samlUsecase) InitiateURL(ctx context.Context, providerName string) (string, error) {
	sp, err := u.buildSP(ctx, providerName)
	if err != nil {
		return "", err
	}
	redirectURL, err := sp.MakeRedirectAuthenticationRequest("")
	if err != nil {
		return "", fmt.Errorf("build authn request: %w", err)
	}
	return redirectURL.String(), nil
}

func (u *samlUsecase) ConsumeACS(ctx context.Context, providerName string, r *http.Request, lc connectors.LoginContext) (string, error) {
	sp, err := u.buildSP(ctx, providerName)
	if err != nil {
		return "", err
	}
	if err := r.ParseForm(); err != nil {
		return "", fmt.Errorf("parse acs form: %w", err)
	}
	raw, err := base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLResponse"))
	if err != nil {
		return "", fmt.Errorf("decode SAMLResponse: %w", err)
	}
	// currentURL is fixed to the configured public ACS URL (not derived from the
	// request) so signature/Destination/Recipient validation works behind nginx.
	assertion, err := sp.ParseXMLResponse(raw, nil, sp.AcsURL)
	if err != nil {
		return "", fmt.Errorf("validate saml response: %w", err)
	}

	login := assertionNameID(assertion)
	if login == "" {
		return "", errors.New("saml assertion has no NameID")
	}
	user, err := u.userRepo.FindByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	if user == nil {
		// No JIT provisioning: the user must already exist locally.
		return "", fmt.Errorf("no local user matches %q", login)
	}
	if !user.Activated {
		return "", errors.New("user is not activated")
	}

	pair, err := IssueTokenPair(ctx, u.userRepo, u.refreshRepo, u.signer, u.refreshTTL, user, lc)
	if err != nil {
		return "", err
	}
	return pair.AccessToken, nil
}

// buildSP loads the IdP config, decrypts the SP key, parses the SP cert and
// fetches the IdP metadata to assemble a crewjam ServiceProvider.
func (u *samlUsecase) buildSP(ctx context.Context, providerName string) (*saml.ServiceProvider, error) {
	cfg, err := u.idpRepo.FindByName(ctx, providerName)
	if err != nil {
		return nil, err
	}
	if cfg == nil || !cfg.Active {
		return nil, domain.ErrIDPNotFound
	}

	keyPem, err := u.cipher.Decrypt(cfg.SpPrivateKeyPem)
	if err != nil {
		return nil, fmt.Errorf("decrypt sp key: %w", err)
	}
	key, err := parsePrivateKey([]byte(keyPem))
	if err != nil {
		return nil, fmt.Errorf("parse sp key: %w", err)
	}
	cert, err := parseCertificate([]byte(cfg.SpCertificatePem))
	if err != nil {
		return nil, fmt.Errorf("parse sp certificate: %w", err)
	}
	acsURL, err := url.Parse(cfg.SpAcsURL)
	if err != nil {
		return nil, fmt.Errorf("parse acs url: %w", err)
	}
	metaURL, err := url.Parse(cfg.MetadataURL)
	if err != nil {
		return nil, fmt.Errorf("parse metadata url: %w", err)
	}
	idpMeta, err := samlsp.FetchMetadata(ctx, u.httpClient, *metaURL)
	if err != nil {
		return nil, fmt.Errorf("fetch idp metadata: %w", err)
	}

	return &saml.ServiceProvider{
		EntityID:          cfg.SpEntityID,
		Key:               key,
		Certificate:       cert,
		AcsURL:            *acsURL,
		IDPMetadata:       idpMeta,
		AllowIDPInitiated: true,
	}, nil
}

func assertionNameID(a *saml.Assertion) string {
	if a == nil || a.Subject == nil || a.Subject.NameID == nil {
		return ""
	}
	return a.Subject.NameID.Value
}

func parsePrivateKey(pemBytes []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no PEM block found in private key")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	signer, ok := keyAny.(crypto.Signer)
	if !ok {
		return nil, errors.New("private key is not a crypto.Signer")
	}
	return signer, nil
}

func parseCertificate(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no PEM block found in certificate")
	}
	return x509.ParseCertificate(block.Bytes)
}
