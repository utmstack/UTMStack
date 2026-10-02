import type { ProviderSettings, ProviderType } from '../types/idp.types'

/**
 * Pure, dependency-free validation for the identity provider form. Returns an
 * i18n key per field that is wrong (an empty object means the form is good), so
 * the component owns the strings and this file owns the rules. The backend only
 * checks presence on save and format at login, so a mistyped URL or key is what
 * this catches before it is written to the directory.
 */

export interface IdpFormInput {
  name: string
  providerType: ProviderType
  settings: ProviderSettings
  /** The protocol secret, blank on edit to mean "keep it". */
  secret: string
  editing: boolean
}

export type IdpFieldErrors = Partial<Record<string, string>>

const NAME_MAX = 64
// The name goes into the SSO login URL, so a space would break it. Dots are
// allowed (acme.entra) — they are legal path segments.
const NAME_RE = /^[A-Za-z0-9][A-Za-z0-9._-]*$/
// Host or hostname the LDAP dial builds into ldap://host:port — no scheme.
const HOST_RE = /^(?!.*[\/\s])[A-Za-z0-9.-]+$/
// Port 1..65535 as a string, no leading zeros.
const PORT_RE = /^(0|[1-9][0-9]*)$/

function isHttpUrl(value: string): boolean {
  try {
    const u = new URL(value)
    return u.protocol === 'http:' || u.protocol === 'https:'
  } catch {
    return false
  }
}

function isHttpsUrl(value: string): boolean {
  try {
    return new URL(value).protocol === 'https:'
  } catch {
    return false
  }
}

// RFC 8252 §7.3: native clients may register loopback redirect URIs over http.
// The token exchange from the loopback host is what keeps them off the wire, so
// http on 127.0.0.1/::1/localhost is the standard's own carve-out and a form
// that rejects it blocks a legitimate mobile app configuration.
const LOOPBACK_HOSTS = new Set(['localhost', '127.0.0.1', '::1'])

function isLoopbackRedirect(value: string): boolean {
  try {
    const u = new URL(value)
    // URL keeps the IPv6 literal bracketed; drop it so [::1] matches ::1.
    const host = u.hostname.replace(/^\[|\]$/g, '')
    return u.protocol === 'http:' && LOOPBACK_HOSTS.has(host)
  } catch {
    return false
  }
}

// A SAML entity ID is a URI, not necessarily an http(s) one — urn: is legal.
// We catch "https://foo com/sp" (the space) and "foo bar" here, which is the
// typo an IdP would reject at login; a pure presence check would not.
function isEntityIdUri(value: string): boolean {
  try {
    const proto = new URL(value).protocol
    return proto === 'http:' || proto === 'https:' || proto === 'urn:'
  } catch {
    return false
  }
}

// A PEM document is its armour lines; the base64 body is never checked because
// an operator pasting a real key always has the markers, and a missing marker
// is exactly the typo we are here to catch.
const CERT_PEM_RE = /-----BEGIN CERTIFICATE-----[\s\S]*?-----END CERTIFICATE-----/
const KEY_PEM_RE =
  /-----BEGIN (RSA |EC |ENCRYPTED |DSA )?PRIVATE KEY-----[\s\S]*?-----END (RSA |EC |ENCRYPTED |DSA )?PRIVATE KEY-----/

function samlErrors(input: IdpFormInput, errors: IdpFieldErrors): void {
  const s = input.settings
  if (s && 'metadataUrl' in s) {
    const metadataUrl = s.metadataUrl.trim()
    if (metadataUrl && !isHttpUrl(metadataUrl)) errors.metadataUrl = 'idp.form.errors.url'
    // Only checked on create: on edit the form carries the stored value, and a
    // legacy non-URI entity ID would otherwise block saving any other field.
    if (!input.editing) {
      const spEntityId = s.spEntityId.trim()
      if (spEntityId && !isEntityIdUri(spEntityId)) errors.spEntityId = 'idp.form.errors.entityId'
    }
    const spAcsUrl = s.spAcsUrl.trim()
    if (spAcsUrl && !isHttpUrl(spAcsUrl)) errors.spAcsUrl = 'idp.form.errors.url'
    if (s.spCertificatePem.trim() && !CERT_PEM_RE.test(s.spCertificatePem)) {
      errors.spCertificatePem = 'idp.form.errors.pemCertificate'
    }
  }
}

function oidcErrors(input: IdpFormInput, errors: IdpFieldErrors): void {
  const s = input.settings
  if (s && 'issuer' in s) {
    // Discovery fetches the issuer, so it must be https to keep the secret safe.
    const issuer = s.issuer.trim()
    if (issuer && !isHttpsUrl(issuer)) errors.issuer = 'idp.form.errors.httpsUrl'
    const redirectUrl = s.redirectUrl.trim()
    // The one carve-out from the https rule is a loopback redirect for a native
    // client, where the token exchange never leaves the machine.
    if (redirectUrl && !isLoopbackRedirect(redirectUrl) && !isHttpsUrl(redirectUrl)) {
      errors.redirectUrl = 'idp.form.errors.redirectUrl'
    }
  }
}

function ldapErrors(input: IdpFormInput, errors: IdpFieldErrors): void {
  const s = input.settings
  if (!s || !('host' in s)) return
  const host = s.host.trim()
  if (host && !HOST_RE.test(host)) errors.host = 'idp.form.errors.hostname'
  const port = String(s.port ?? '').trim()
  if (port !== '' && (!PORT_RE.test(port) || Number(port) > 65535)) errors.port = 'idp.form.errors.port'
  const userFilter = s.userFilter.trim()
  if (userFilter) {
    // A filter is a chain of (attr=value); it needs the placeholder and balanced
    // parens. Both are the common ways a search silently returns nothing.
    if (!userFilter.includes('%s')) errors.userFilter = 'idp.form.errors.userFilterPlaceholder'
    else if ((userFilter.match(/\(/g)?.length ?? 0) !== (userFilter.match(/\)/g)?.length ?? 0)) {
      errors.userFilter = 'idp.form.errors.userFilterUnbalanced'
    }
  }
}

function commonErrors(input: IdpFormInput, errors: IdpFieldErrors): void {
  const name = input.name.trim()
  if (!name) {
    errors.name = 'idp.form.errors.required'
    return
  }
  // On edit the name input is disabled, so a name with a space or a dot in it
  // (legal at save time, but odd in a URL) would otherwise lock the whole form
  // out of saving any other field.
  if (!input.editing && (name.length > NAME_MAX || !NAME_RE.test(name))) {
    errors.name = 'idp.form.errors.nameFormat'
  }
}

export function validateIdpForm(input: IdpFormInput): IdpFieldErrors {
  const errors: IdpFieldErrors = {}
  commonErrors(input, errors)

  // Presence of the protocol's required settings is the form's existing
  // job; the format checks above only fire on values that are present.
  const s = input.settings
  if (input.providerType === 'saml' && s && 'metadataUrl' in s) {
    const missing = ['metadataUrl', 'spEntityId', 'spAcsUrl', 'spCertificatePem'].filter(
      (k) => String(s[k as keyof typeof s] ?? '').trim() === '',
    )
    for (const k of missing) errors[k] = errors[k] ?? 'idp.form.errors.required'
    // The secret is required on create (the backend keeps nothing else) and
    // optional on edit (blank means keep it).
    if (!input.editing && !input.secret.trim()) {
      errors.spPrivateKeyPem = 'idp.form.errors.required'
    } else if (input.secret.trim() && !KEY_PEM_RE.test(input.secret)) {
      errors.spPrivateKeyPem = 'idp.form.errors.pemKey'
    }
  } else if (input.providerType === 'oidc' && s && 'issuer' in s) {
    for (const k of ['issuer', 'clientId', 'redirectUrl']) {
      if (String(s[k as keyof typeof s] ?? '').trim() === '') errors[k] = errors[k] ?? 'idp.form.errors.required'
    }
    if (!input.editing && !input.secret.trim()) errors.clientSecret = 'idp.form.errors.required'
  } else if (input.providerType === 'ldap' && s && 'host' in s) {
    for (const k of ['host', 'bindDn', 'baseDn', 'userFilter']) {
      if (String(s[k as keyof typeof s] ?? '').trim() === '') errors[k] = errors[k] ?? 'idp.form.errors.required'
    }
    if (!input.editing && !input.secret.trim()) errors.bindPassword = 'idp.form.errors.required'
  }

  if (input.providerType === 'saml') samlErrors(input, errors)
  else if (input.providerType === 'oidc') oidcErrors(input, errors)
  else ldapErrors(input, errors)

  return errors
}
