package usecase

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	oidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/go-ldap/ldap/v3"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gorm.io/datatypes"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/jwt"
)

// Cipher is what turns the one secret each protocol carries into something a
// database read cannot use.
type Cipher interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

type identityProviderUsecase struct {
	repo   connectors.IdentityProviderRepository
	rbac   connectors.RBACRepository
	cipher Cipher
}

func NewIdentityProviderUsecase(
	repo connectors.IdentityProviderRepository,
	rbac connectors.RBACRepository,
	cipher Cipher,
) connectors.IdentityProviderUsecase {
	return &identityProviderUsecase{repo: repo, rbac: rbac, cipher: cipher}
}

func settingsOf[T any](p *domain.IdentityProviderConfig) (T, error) {
	var out T
	if len(p.Settings) == 0 {
		return out, domain.ErrIDPSettingsInvalid
	}
	if err := json.Unmarshal(p.Settings, &out); err != nil {
		return out, domain.ErrIDPSettingsInvalid
	}
	return out, nil
}

func samlSettings(p *domain.IdentityProviderConfig) (domain.SAMLSettings, error) {
	return settingsOf[domain.SAMLSettings](p)
}

func oidcSettings(p *domain.IdentityProviderConfig) (domain.OIDCSettings, error) {
	return settingsOf[domain.OIDCSettings](p)
}

func ldapSettings(p *domain.IdentityProviderConfig) (domain.LDAPSettings, error) {
	return settingsOf[domain.LDAPSettings](p)
}

// prepareSettings validates what the chosen protocol needs and encrypts the one
// secret each of them carries. A secret is never stored as it arrived and never
// leaves again: readSettings blanks it on the way out.
func (u *identityProviderUsecase) prepareSettings(
	kind domain.ProviderType, raw json.RawMessage, previous *domain.IdentityProviderConfig,
) (datatypes.JSON, error) {
	switch kind {
	case domain.ProviderSAML:
		var s domain.SAMLSettings
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, domain.ErrIDPSettingsInvalid
		}
		if blank(s.MetadataURL, s.SpEntityID, s.SpACSURL, s.SpCertificatePem) {
			return nil, domain.ErrIDPSettingsInvalid
		}
		kept, err := u.keepOrEncrypt(s.SpPrivateKeyPem, previous, func(p *domain.IdentityProviderConfig) string {
			old, _ := samlSettings(p)
			return old.SpPrivateKeyPem
		})
		if err != nil {
			return nil, err
		}
		if kept == "" {
			return nil, domain.ErrIDPKeyRequired
		}
		s.SpPrivateKeyPem = kept
		return json.Marshal(s)

	case domain.ProviderOIDC:
		var s domain.OIDCSettings
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, domain.ErrIDPSettingsInvalid
		}
		if blank(s.Issuer, s.ClientID, s.RedirectURL) {
			return nil, domain.ErrIDPSettingsInvalid
		}
		kept, err := u.keepOrEncrypt(s.ClientSecret, previous, func(p *domain.IdentityProviderConfig) string {
			old, _ := oidcSettings(p)
			return old.ClientSecret
		})
		if err != nil {
			return nil, err
		}
		if kept == "" {
			return nil, domain.ErrIDPSettingsInvalid
		}
		s.ClientSecret = kept
		return json.Marshal(s)

	case domain.ProviderLDAP:
		var s domain.LDAPSettings
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, domain.ErrIDPSettingsInvalid
		}
		// The filter is where the address is substituted; without the placeholder
		// every login would search for the same literal string.
		if blank(s.Host, s.BindDN, s.BaseDN, s.UserFilter) || !strings.Contains(s.UserFilter, "%s") {
			return nil, domain.ErrIDPSettingsInvalid
		}
		kept, err := u.keepOrEncrypt(s.BindPassword, previous, func(p *domain.IdentityProviderConfig) string {
			old, _ := ldapSettings(p)
			return old.BindPassword
		})
		if err != nil {
			return nil, err
		}
		if kept == "" {
			return nil, domain.ErrIDPSettingsInvalid
		}
		s.BindPassword = kept
		return json.Marshal(s)
	}
	return nil, domain.ErrIDPTypeUnsupported
}

// keepOrEncrypt lets an update omit the secret to keep the stored one, so an
// operator editing a filter does not have to retype a private key they cannot
// read back.
func (u *identityProviderUsecase) keepOrEncrypt(
	incoming string, previous *domain.IdentityProviderConfig, from func(*domain.IdentityProviderConfig) string,
) (string, error) {
	if strings.TrimSpace(incoming) == "" {
		if previous == nil {
			return "", nil
		}
		return from(previous), nil
	}
	return u.cipher.Encrypt(incoming)
}

// readSettings is what a caller may see: everything except the secret.
func readSettings(p domain.IdentityProviderConfig) datatypes.JSON {
	var generic map[string]any
	if err := json.Unmarshal(p.Settings, &generic); err != nil {
		return nil
	}
	for _, secret := range []string{"spPrivateKeyPem", "clientSecret", "bindPassword"} {
		delete(generic, secret)
	}
	out, err := json.Marshal(generic)
	if err != nil {
		return nil
	}
	return out
}

func (u *identityProviderUsecase) Create(ctx context.Context, req dto.IdentityProviderRequest) (*domain.IdentityProviderConfig, error) {
	if req.ID != uuid.Nil {
		return nil, domain.ErrIDPIDForbidden
	}
	cfg, err := u.build(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	if err := u.repo.Save(ctx, cfg); err != nil {
		return nil, err
	}
	if err := u.saveMappings(ctx, cfg, req.GroupMappings); err != nil {
		return nil, err
	}
	return u.GetByID(ctx, cfg.ID)
}

func (u *identityProviderUsecase) Update(ctx context.Context, req dto.IdentityProviderRequest) (*domain.IdentityProviderConfig, error) {
	if req.ID == uuid.Nil {
		return nil, domain.ErrIDPIDRequired
	}
	existing, err := u.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrIDPNotFound
	}
	cfg, err := u.build(ctx, req, existing)
	if err != nil {
		return nil, err
	}
	cfg.ID = existing.ID
	cfg.TenantID = existing.TenantID
	cfg.CreatedAt = existing.CreatedAt
	if err := u.repo.Save(ctx, cfg); err != nil {
		return nil, err
	}
	if err := u.saveMappings(ctx, cfg, req.GroupMappings); err != nil {
		return nil, err
	}
	return u.GetByID(ctx, cfg.ID)
}

func (u *identityProviderUsecase) build(
	ctx context.Context, req dto.IdentityProviderRequest, previous *domain.IdentityProviderConfig,
) (*domain.IdentityProviderConfig, error) {
	kind := domain.ProviderType(strings.ToLower(strings.TrimSpace(req.ProviderType)))
	if !kind.Valid() {
		return nil, domain.ErrIDPTypeUnsupported
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, domain.ErrIDPInvalidInput
	}
	settings, err := u.prepareSettings(kind, req.Settings, previous)
	if err != nil {
		return nil, err
	}
	if err := u.assertRoleBelongs(ctx, req.DefaultRoleID); err != nil {
		return nil, err
	}

	return &domain.IdentityProviderConfig{
		Name:             strings.TrimSpace(req.Name),
		ProviderType:     kind,
		Active:           req.Active,
		Settings:         settings,
		JITProvisioning:  req.JITProvisioning,
		DefaultRoleID:    req.DefaultRoleID,
		GroupsAttribute:  strings.TrimSpace(req.GroupsAttribute),
		SyncRolesOnLogin: req.SyncRolesOnLogin,
	}, nil
}

func (u *identityProviderUsecase) assertRoleBelongs(ctx context.Context, roleID *uuid.UUID) error {
	if roleID == nil {
		return nil
	}
	role, err := u.rbac.FindRoleByID(ctx, *roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return domain.ErrRoleNotFound
	}
	return nil
}

func (u *identityProviderUsecase) saveMappings(
	ctx context.Context, cfg *domain.IdentityProviderConfig, in []dto.GroupMapping,
) error {
	rows := make([]domain.IdentityProviderGroupMapping, 0, len(in))
	for _, m := range in {
		if strings.TrimSpace(m.Group) == "" {
			continue
		}
		if err := u.assertRoleBelongs(ctx, &m.RoleID); err != nil {
			return err
		}
		rows = append(rows, domain.IdentityProviderGroupMapping{
			IdentityProviderID: cfg.ID,
			TenantID:           cfg.TenantID,
			Group:              strings.TrimSpace(m.Group),
			RoleID:             m.RoleID,
		})
	}
	return u.repo.ReplaceMappings(ctx, cfg.ID, rows)
}

func (u *identityProviderUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.IdentityProviderConfig, error) {
	cfg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, domain.ErrIDPNotFound
	}
	cfg.Settings = readSettings(*cfg)
	return cfg, nil
}

func (u *identityProviderUsecase) List(ctx context.Context, f dto.IdentityProviderFilter) ([]domain.IdentityProviderConfig, int64, error) {
	items, total, err := u.repo.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	for i := range items {
		items[i].Settings = readSettings(items[i])
	}
	return items, total, nil
}

func (u *identityProviderUsecase) ListActive(ctx context.Context) ([]dto.IdentityProviderPublic, error) {
	items, err := u.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.IdentityProviderPublic, 0, len(items))
	for _, c := range items {
		p := dto.IdentityProviderPublic{
			ID:           c.ID,
			Name:         c.Name,
			ProviderType: string(c.ProviderType),
		}
		if c.ProviderType.Redirecting() {
			p.LoginURL = "/api/v1/sso/" + c.Name + "/login"
		}
		out = append(out, p)
	}
	return out, nil
}

func (u *identityProviderUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	cfg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if cfg == nil {
		return domain.ErrIDPNotFound
	}
	return u.repo.Delete(ctx, id)
}

func (u *identityProviderUsecase) ListMappings(ctx context.Context, id uuid.UUID) ([]domain.IdentityProviderGroupMapping, error) {
	return u.repo.ListMappings(ctx, id)
}

func blank(values ...string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			return true
		}
	}
	return false
}

type federationUsecase struct {
	idpRepo     connectors.IdentityProviderRepository
	userRepo    connectors.UserRepository
	rbacRepo    connectors.RBACRepository
	refreshRepo connectors.RefreshTokenRepository
	signer      *jwt.Signer
	cipher      Cipher
	refreshTTL  time.Duration
	httpClient  *http.Client
}

func NewFederationUsecase(
	idpRepo connectors.IdentityProviderRepository,
	userRepo connectors.UserRepository,
	rbacRepo connectors.RBACRepository,
	refreshRepo connectors.RefreshTokenRepository,
	signer *jwt.Signer,
	cipher Cipher,
	refreshTTL time.Duration,
) connectors.FederationUsecase {
	return &federationUsecase{
		idpRepo:     idpRepo,
		userRepo:    userRepo,
		rbacRepo:    rbacRepo,
		refreshRepo: refreshRepo,
		signer:      signer,
		cipher:      cipher,
		refreshTTL:  refreshTTL,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (u *federationUsecase) provider(ctx context.Context, name string) (*domain.IdentityProviderConfig, error) {
	cfg, err := u.idpRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if cfg == nil || !cfg.Active {
		return nil, domain.ErrIDPNotFound
	}
	return cfg, nil
}

func (u *federationUsecase) complete(
	ctx context.Context,
	p *domain.IdentityProviderConfig,
	id domain.FederatedIdentity,
	lc connectors.LoginContext,
) (*dto.TokenPair, error) {
	email := strings.ToLower(strings.TrimSpace(id.Email))
	if email == "" {
		return nil, domain.ErrFederatedUnknownUser
	}

	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	roles, err := u.rolesFor(ctx, p, id.Groups)
	if err != nil {
		return nil, err
	}

	switch {
	case user == nil:
		if !p.JITProvisioning {
			return nil, domain.ErrFederatedUnknownUser
		}
		user = &domain.User{
			Email:              email,
			Name:               strings.TrimSpace(id.Name),
			TenantID:           p.TenantID,
			Status:             domain.UserStatusActive,
			IdentityProviderID: &p.ID,
		}
		if err := u.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
		if err := u.rbacRepo.ReplaceUserRoles(ctx, user.ID, roles); err != nil {
			return nil, err
		}

	default:
		if err := statusError(user.Status); err != nil {
			return nil, err
		}
		if p.SyncRolesOnLogin {
			if err := u.rbacRepo.ReplaceUserRoles(ctx, user.ID, roles); err != nil {
				return nil, err
			}
		}
	}

	return IssueTokenPair(ctx, u.userRepo, u.refreshRepo, u.signer, u.refreshTTL, user, lc)
}

func (u *federationUsecase) rolesFor(
	ctx context.Context,
	p *domain.IdentityProviderConfig,
	groups []string,
) ([]uuid.UUID, error) {
	seen := make(map[uuid.UUID]struct{})
	out := make([]uuid.UUID, 0, len(groups))

	if len(groups) > 0 {
		mappings, err := u.idpRepo.ListMappings(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		byGroup := make(map[string]uuid.UUID, len(mappings))
		for _, m := range mappings {
			byGroup[strings.ToLower(m.Group)] = m.RoleID
		}
		for _, g := range groups {
			roleID, ok := byGroup[strings.ToLower(strings.TrimSpace(g))]
			if !ok {
				continue
			}
			if _, dup := seen[roleID]; dup {
				continue
			}
			seen[roleID] = struct{}{}
			out = append(out, roleID)
		}
	}

	if len(out) == 0 {
		if p.DefaultRoleID == nil {
			return nil, domain.ErrFederatedNoRoles
		}
		out = append(out, *p.DefaultRoleID)
	}
	return out, nil
}

func (u *federationUsecase) StartURL(ctx context.Context, providerName string) (string, string, error) {
	p, err := u.provider(ctx, providerName)
	if err != nil {
		return "", "", err
	}
	state, err := newState()
	if err != nil {
		return "", "", err
	}
	switch p.ProviderType {
	case domain.ProviderSAML:
		return u.samlRedirect(ctx, p)
	case domain.ProviderOIDC:
		url, err := u.oidcRedirect(ctx, p, state)
		return url, state, err
	}
	return "", "", domain.ErrIDPTypeUnsupported
}

func (u *federationUsecase) samlRedirect(ctx context.Context, p *domain.IdentityProviderConfig) (string, string, error) {
	sp, err := u.buildSP(ctx, p)
	if err != nil {
		return "", "", err
	}
	req, err := sp.MakeAuthenticationRequest(
		sp.GetSSOBindingLocation(saml.HTTPRedirectBinding),
		saml.HTTPRedirectBinding, saml.HTTPPostBinding)
	if err != nil {
		return "", "", fmt.Errorf("build authn request: %w", err)
	}
	redirect, err := req.Redirect("", sp)
	if err != nil {
		return "", "", fmt.Errorf("build authn redirect: %w", err)
	}
	return redirect.String(), req.ID, nil
}

func (u *federationUsecase) ConsumeSAML(
	ctx context.Context, providerName string, r *http.Request, requestID string, lc connectors.LoginContext,
) (*dto.TokenPair, error) {
	p, err := u.provider(ctx, providerName)
	if err != nil {
		return nil, err
	}
	if p.ProviderType != domain.ProviderSAML {
		return nil, domain.ErrIDPTypeUnsupported
	}
	sp, err := u.buildSP(ctx, p)
	if err != nil {
		return nil, err
	}
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("parse acs form: %w", err)
	}
	raw, err := base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLResponse"))
	if err != nil {
		return nil, fmt.Errorf("decode SAMLResponse: %w", err)
	}

	var expected []string
	if requestID != "" {
		expected = []string{requestID}
	}
	assertion, err := sp.ParseXMLResponse(raw, expected, sp.AcsURL)
	if err != nil {

		var invalid *saml.InvalidResponseError
		if errors.As(err, &invalid) && invalid.PrivateErr != nil {
			return nil, fmt.Errorf("validate saml response: %w", invalid.PrivateErr)
		}
		return nil, fmt.Errorf("validate saml response: %w", err)
	}

	email := assertionAttr(assertion, "email", "mail", "emailaddress")
	if email == "" {
		email = assertionNameID(assertion)
	}

	return u.complete(ctx, p, domain.FederatedIdentity{
		Email:  email,
		Name:   assertionAttr(assertion, "displayName", "name", "cn"),
		Groups: assertionValues(assertion, p.GroupsAttribute),
	}, lc)
}

func (u *federationUsecase) buildSP(ctx context.Context, p *domain.IdentityProviderConfig) (*saml.ServiceProvider, error) {
	cfg, err := samlSettings(p)
	if err != nil {
		return nil, err
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
	metaURL, err := url.Parse(cfg.MetadataURL)
	if err != nil {
		return nil, fmt.Errorf("parse metadata url: %w", err)
	}
	idpMeta, err := samlsp.FetchMetadata(ctx, u.httpClient, *metaURL)
	if err != nil {
		return nil, fmt.Errorf("fetch idp metadata: %w", err)
	}
	acs, err := url.Parse(cfg.SpACSURL)
	if err != nil {
		return nil, fmt.Errorf("parse acs url: %w", err)
	}

	return &saml.ServiceProvider{
		EntityID:    cfg.SpEntityID,
		Key:         key.(crypto.Signer),
		Certificate: cert,
		AcsURL:      *acs,
		IDPMetadata: idpMeta,
	}, nil
}

func assertionNameID(a *saml.Assertion) string {
	if a == nil || a.Subject == nil || a.Subject.NameID == nil {
		return ""
	}
	return strings.TrimSpace(a.Subject.NameID.Value)
}

func assertionValues(a *saml.Assertion, name string) []string {
	if a == nil || name == "" {
		return nil
	}
	var out []string
	for _, stmt := range a.AttributeStatements {
		for _, attr := range stmt.Attributes {
			if !attrMatches(attr, name) {
				continue
			}
			for _, v := range attr.Values {
				if s := strings.TrimSpace(v.Value); s != "" {
					out = append(out, s)
				}
			}
		}
	}
	return out
}

func assertionAttr(a *saml.Assertion, names ...string) string {
	for _, n := range names {
		if v := assertionValues(a, n); len(v) > 0 {
			return v[0]
		}
	}
	return ""
}

func attrMatches(attr saml.Attribute, want string) bool {
	if strings.EqualFold(attr.Name, want) || strings.EqualFold(attr.FriendlyName, want) {
		return true
	}
	if i := strings.LastIndex(attr.Name, "/"); i >= 0 {
		return strings.EqualFold(attr.Name[i+1:], want)
	}
	return false
}

func parsePrivateKey(pemBytes []byte) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no PEM block found")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	if k, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	return nil, errors.New("unsupported private key format")
}

func parseCertificate(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		// Some IdP exports carry the base64 body with no PEM armour.
		der, err := base64.StdEncoding.DecodeString(
			strings.NewReplacer("\n", "", "\r", "", " ", "").Replace(string(pemBytes)))
		if err != nil {
			return nil, errors.New("no PEM block found")
		}
		return x509.ParseCertificate(der)
	}
	return x509.ParseCertificate(block.Bytes)
}

func newState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (u *federationUsecase) oidcConfig(ctx context.Context, p *domain.IdentityProviderConfig) (*oauth2.Config, *oidc.Provider, error) {
	cfg, err := oidcSettings(p)
	if err != nil {
		return nil, nil, err
	}
	secret, err := u.cipher.Decrypt(cfg.ClientSecret)
	if err != nil {
		return nil, nil, fmt.Errorf("decrypt client secret: %w", err)
	}

	ctx = oidc.ClientContext(ctx, u.httpClient)
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, nil, fmt.Errorf("oidc discovery: %w", err)
	}

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: secret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURL,
		Scopes:       scopes,
	}, provider, nil
}

func (u *federationUsecase) oidcRedirect(ctx context.Context, p *domain.IdentityProviderConfig, state string) (string, error) {
	conf, _, err := u.oidcConfig(ctx, p)
	if err != nil {
		return "", err
	}
	return conf.AuthCodeURL(state), nil
}

func (u *federationUsecase) ConsumeOIDC(
	ctx context.Context, providerName, code, state, wantState string, lc connectors.LoginContext,
) (*dto.TokenPair, error) {
	p, err := u.provider(ctx, providerName)
	if err != nil {
		return nil, err
	}
	if p.ProviderType != domain.ProviderOIDC {
		return nil, domain.ErrIDPTypeUnsupported
	}
	// The state is what ties this callback to the redirect that started it;
	// without the check anyone could hand the browser a code of their own.
	if wantState == "" || state != wantState {
		return nil, domain.ErrIDPStateInvalid
	}

	conf, provider, err := u.oidcConfig(ctx, p)
	if err != nil {
		return nil, err
	}
	ctx = oidc.ClientContext(ctx, u.httpClient)

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange authorization code: %w", err)
	}
	rawID, ok := token.Extra("id_token").(string)
	if !ok || rawID == "" {
		return nil, fmt.Errorf("the provider returned no id_token")
	}
	idToken, err := provider.Verifier(&oidc.Config{ClientID: conf.ClientID}).Verify(ctx, rawID)
	if err != nil {
		return nil, fmt.Errorf("verify id_token: %w", err)
	}

	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("read id_token claims: %w", err)
	}

	groupsClaim := p.GroupsAttribute
	if groupsClaim == "" {
		groupsClaim = "groups"
	}

	return u.complete(ctx, p, domain.FederatedIdentity{
		Email:  claimString(claims, "email", "preferred_username", "upn"),
		Name:   claimString(claims, "name", "given_name"),
		Groups: claimStrings(claims, groupsClaim),
	}, lc)
}

func claimString(claims map[string]any, names ...string) string {
	for _, n := range names {
		if s, ok := claims[n].(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func claimStrings(claims map[string]any, name string) []string {
	switch v := claims[name].(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	case []string:
		return v
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		return []string{strings.TrimSpace(v)}
	}
	return nil
}

func (u *federationUsecase) AuthenticateLDAP(
	ctx context.Context, email, password string, providerID *uuid.UUID, lc connectors.LoginContext,
) (*dto.TokenPair, error) {
	if strings.TrimSpace(email) == "" || password == "" {
		return nil, nil
	}
	providers, err := u.idpRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	for i := range providers {
		p := &providers[i]
		if p.ProviderType != domain.ProviderLDAP {
			continue
		}
		if providerID != nil && p.ID != *providerID {
			continue
		}
		id, err := u.ldapBind(p, email, password)
		if err != nil || id == nil {
			continue
		}
		return u.complete(authz.WithTenantID(ctx, p.TenantID.String()), p, *id, lc)
	}
	return nil, nil
}

func (u *federationUsecase) ldapBind(p *domain.IdentityProviderConfig, email, password string) (*domain.FederatedIdentity, error) {
	cfg, err := ldapSettings(p)
	if err != nil {
		return nil, err
	}
	bindPassword, err := u.cipher.Decrypt(cfg.BindPassword)
	if err != nil {
		return nil, fmt.Errorf("decrypt bind password: %w", err)
	}

	port := cfg.Port
	if port == 0 {
		port = 389
	}
	conn, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", cfg.Host, port))
	if err != nil {
		return nil, fmt.Errorf("dial ldap: %w", err)
	}
	defer conn.Close()

	if cfg.StartTLS {
		if err := conn.StartTLS(&tls.Config{
			ServerName:         cfg.Host,
			InsecureSkipVerify: cfg.InsecureTLS,
		}); err != nil {
			return nil, fmt.Errorf("starttls: %w", err)
		}
	}

	if err := conn.Bind(cfg.BindDN, bindPassword); err != nil {
		return nil, fmt.Errorf("service bind: %w", err)
	}

	filter := cfg.UserFilter
	if !strings.Contains(filter, "%s") {
		filter = "(mail=%s)"
	}
	attrs := compact([]string{cfg.EmailAttribute, cfg.NameAttribute, cfg.GroupAttribute})

	res, err := conn.Search(ldap.NewSearchRequest(
		cfg.BaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 2, 10, false,
		fmt.Sprintf(filter, ldap.EscapeFilter(email)), attrs, nil,
	))
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	// More than one match means the filter is ambiguous, and binding to whichever
	// came first would be a coin toss over whose account this is.
	if len(res.Entries) != 1 {
		return nil, fmt.Errorf("the filter matched %d entries", len(res.Entries))
	}
	entry := res.Entries[0]

	if err := conn.Bind(entry.DN, password); err != nil {
		return nil, fmt.Errorf("user bind: %w", err)
	}

	mail := entry.GetAttributeValue(orDefault(cfg.EmailAttribute, "mail"))
	if strings.TrimSpace(mail) == "" {
		mail = email
	}
	return &domain.FederatedIdentity{
		Email:  mail,
		Name:   entry.GetAttributeValue(orDefault(cfg.NameAttribute, "displayName")),
		Groups: groupNames(entry.GetAttributeValues(orDefault(cfg.GroupAttribute, "memberOf"))),
	}, nil
}

func groupNames(dns []string) []string {
	out := make([]string, 0, len(dns))
	for _, dn := range dns {
		name := dn
		if parsed, err := ldap.ParseDN(dn); err == nil && len(parsed.RDNs) > 0 {
			for _, attr := range parsed.RDNs[0].Attributes {
				if strings.EqualFold(attr.Type, "cn") {
					name = attr.Value
					break
				}
			}
		}
		if s := strings.TrimSpace(name); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func compact(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

func orDefault(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
