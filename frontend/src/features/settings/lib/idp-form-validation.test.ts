import { describe, expect, it } from 'vitest'
import { validateIdpForm, type IdpFormInput } from './idp-form-validation'
import type { LdapSettings, OidcSettings, SamlSettings } from '../types/idp.types'

const CERT = '-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----'
const KEY = '-----BEGIN PRIVATE KEY-----\nMIIB\n-----END PRIVATE KEY-----'

const samlSettings: SamlSettings = {
  metadataUrl: 'https://idp.example.com/metadata',
  spEntityId: 'https://utmstack.example.com/sp',
  spAcsUrl: 'https://utmstack.example.com/api/v1/sso/saml/acme-entra/login',
  spCertificatePem: CERT,
}
const oidcSettings: OidcSettings = {
  issuer: 'https://login.microsoftonline.com/tenant/v2.0',
  clientId: 'client-123',
  redirectUrl: 'https://utmstack.example.com/api/v1/sso/oidc/acme/login',
}
const ldapSettings: LdapSettings = {
  host: 'ldap.example.com',
  port: 389,
  startTls: true,
  bindDn: 'cn=svc,cn=users,dc=example,dc=com',
  baseDn: 'dc=example,dc=com',
  userFilter: '(mail=%s)',
  emailAttribute: 'mail',
  nameAttribute: 'displayName',
  groupAttribute: 'memberOf',
}

interface CommonOverrides {
  name?: string
  editing?: boolean
  secret?: string
}

function samlBase(o: CommonOverrides = {}, s: Partial<SamlSettings> = {}): IdpFormInput {
  return {
    name: o.name ?? 'acme-entra',
    editing: o.editing ?? false,
    providerType: 'saml',
    secret: o.secret ?? KEY,
    settings: { ...samlSettings, ...s },
  }
}
function oidcBase(o: CommonOverrides = {}, s: Partial<OidcSettings> = {}): IdpFormInput {
  return {
    name: o.name ?? 'acme-entra',
    editing: o.editing ?? false,
    providerType: 'oidc',
    secret: o.secret ?? 'client-secret',
    settings: { ...oidcSettings, ...s },
  }
}
function ldapBase(o: CommonOverrides = {}, s: Partial<LdapSettings> = {}): IdpFormInput {
  return {
    name: o.name ?? 'acme-ad',
    editing: o.editing ?? false,
    providerType: 'ldap',
    secret: o.secret ?? 'bind-password',
    settings: { ...ldapSettings, ...s },
  }
}

describe('validateIdpForm — SAML', () => {
  it('accepts a fully valid create form', () => {
    expect(validateIdpForm(samlBase())).toEqual({})
  })

  it('accepts an edit form with the secret left blank', () => {
    expect(validateIdpForm(samlBase({ editing: true, secret: '' }))).toEqual({})
  })

  it('flags a metadataUrl that is not a URL', () => {
    expect(validateIdpForm(samlBase({}, { metadataUrl: 'not a url' })).metadataUrl).toBe('idp.form.errors.url')
  })

  it('flags an spAcsUrl that is not a URL', () => {
    expect(validateIdpForm(samlBase({}, { spAcsUrl: 'ftp://x' })).spAcsUrl).toBe('idp.form.errors.url')
  })

  it('flags a certificate that has no PEM armour', () => {
    expect(validateIdpForm(samlBase({}, { spCertificatePem: 'MIIBnotarmoured' })).spCertificatePem).toBe(
      'idp.form.errors.pemCertificate',
    )
  })

  it('requires the private key on create', () => {
    expect(validateIdpForm(samlBase({ secret: '' })).spPrivateKeyPem).toBe('idp.form.errors.required')
  })

  it('flags a private key that has no PEM armour on create', () => {
    expect(validateIdpForm(samlBase({ secret: 'just some text' })).spPrivateKeyPem).toBe('idp.form.errors.pemKey')
  })

  it('does not flag a blank key on edit (keep the stored one)', () => {
    expect(validateIdpForm(samlBase({ editing: true, secret: '' })).spPrivateKeyPem).toBeUndefined()
  })

  it('flags a blank required setting', () => {
    expect(validateIdpForm(samlBase({}, { spEntityId: '' })).spEntityId).toBe('idp.form.errors.required')
  })

  it('accepts an https entity ID on create', () => {
    expect(validateIdpForm(samlBase({}, { spEntityId: 'https://utmstack.example.com/sp' })).spEntityId).toBeUndefined()
  })

  it('accepts a urn: entity ID on create', () => {
    expect(validateIdpForm(samlBase({}, { spEntityId: 'urn:utmstack:sp' })).spEntityId).toBeUndefined()
  })

  it('flags an entity ID that is not a URI (has a space) on create', () => {
    expect(validateIdpForm(samlBase({}, { spEntityId: 'https://foo com/sp' })).spEntityId).toBe(
      'idp.form.errors.entityId',
    )
  })

  it('does not flag a non-URI legacy entity ID on edit', () => {
    // The field is not disabled on edit, so a value inherited before this rule
    // must not lock the form out of saving any other change.
    expect(validateIdpForm(samlBase({ editing: true }, { spEntityId: 'not a uri' })).spEntityId).toBeUndefined()
  })
})

describe('validateIdpForm — OIDC', () => {
  it('accepts a fully valid create form', () => {
    expect(validateIdpForm(oidcBase())).toEqual({})
  })

  it('flags an http issuer (must be https)', () => {
    expect(validateIdpForm(oidcBase({}, { issuer: 'http://login.example.com' })).issuer).toBe(
      'idp.form.errors.httpsUrl',
    )
  })

  it('flags a redirectUrl that is not a URL', () => {
    expect(validateIdpForm(oidcBase({}, { redirectUrl: 'not a url' })).redirectUrl).toBe(
      'idp.form.errors.redirectUrl',
    )
  })

  it('accepts an https redirectUrl', () => {
    expect(validateIdpForm(oidcBase({}, { redirectUrl: 'https://utmstack.example.com/cb' })).redirectUrl).toBeUndefined()
  })

  it.each(['http://localhost:3000/cb', 'http://127.0.0.1/cb', 'http://[::1]:8080/cb'])(
    'accepts a loopback redirectUrl over http (RFC 8252): %s',
    (url) => {
      expect(validateIdpForm(oidcBase({}, { redirectUrl: url })).redirectUrl).toBeUndefined()
    },
  )

  it('rejects a private-range redirect over http (loopback is the only carve-out)', () => {
    expect(validateIdpForm(oidcBase({}, { redirectUrl: 'http://10.0.0.5/cb' })).redirectUrl).toBe(
      'idp.form.errors.redirectUrl',
    )
  })

  it('requires the client secret on create', () => {
    expect(validateIdpForm(oidcBase({ secret: '' })).clientSecret).toBe('idp.form.errors.required')
  })

  it('does not require the client secret on edit', () => {
    expect(validateIdpForm(oidcBase({ editing: true, secret: '' })).clientSecret).toBeUndefined()
  })

  it('flags a blank client id', () => {
    expect(validateIdpForm(oidcBase({}, { clientId: '' })).clientId).toBe('idp.form.errors.required')
  })
})

describe('validateIdpForm — LDAP', () => {
  it('accepts a fully valid create form', () => {
    expect(validateIdpForm(ldapBase())).toEqual({})
  })

  it('flags a host that carries a scheme', () => {
    expect(validateIdpForm(ldapBase({}, { host: 'ldap://ldap.example.com' })).host).toBe('idp.form.errors.hostname')
  })

  it('flags a host with a path', () => {
    expect(validateIdpForm(ldapBase({}, { host: 'ldap.example.com/dc=foo' })).host).toBe('idp.form.errors.hostname')
  })

  it('accepts a plain IP host', () => {
    expect(validateIdpForm(ldapBase({}, { host: '10.0.0.5' })).host).toBeUndefined()
  })

  it('flags a port above 65535', () => {
    expect(validateIdpForm(ldapBase({}, { port: 99999 })).port).toBe('idp.form.errors.port')
  })

  it('accepts a zero port (the backend dials it as 389)', () => {
    // Clearing the field yields 0; that is the default, not a mistype, so it must
    // not lock the form out of saving.
    expect(validateIdpForm(ldapBase({}, { port: 0 })).port).toBeUndefined()
  })

  it('flags a filter missing the %s placeholder', () => {
    expect(validateIdpForm(ldapBase({}, { userFilter: '(mail=jane@example.com)' })).userFilter).toBe(
      'idp.form.errors.userFilterPlaceholder',
    )
  })

  it('flags a filter with unbalanced parentheses', () => {
    expect(validateIdpForm(ldapBase({}, { userFilter: '(&|(mail=%s)(userPrincipalName=%s)' })).userFilter).toBe(
      'idp.form.errors.userFilterUnbalanced',
    )
  })

  it('accepts a balanced filter with the placeholder', () => {
    expect(validateIdpForm(ldapBase({}, { userFilter: '(&(|(mail=%s)(upn=%s)))' })).userFilter).toBeUndefined()
  })

  it('requires the bind password on create', () => {
    expect(validateIdpForm(ldapBase({ secret: '' })).bindPassword).toBe('idp.form.errors.required')
  })

  it('does not require the bind password on edit', () => {
    expect(validateIdpForm(ldapBase({ editing: true, secret: '' })).bindPassword).toBeUndefined()
  })

  it('flags a blank baseDn', () => {
    expect(validateIdpForm(ldapBase({}, { baseDn: '' })).baseDn).toBe('idp.form.errors.required')
  })
})

describe('validateIdpForm — name', () => {
  it('requires a name', () => {
    expect(validateIdpForm(samlBase({ name: '  ' })).name).toBe('idp.form.errors.required')
  })

  it('rejects a name with a space on create', () => {
    expect(validateIdpForm(samlBase({ name: 'acme entra' })).name).toBe('idp.form.errors.nameFormat')
  })

  it('rejects a name longer than 64 on create', () => {
    expect(validateIdpForm(samlBase({ name: 'a'.repeat(65) })).name).toBe('idp.form.errors.nameFormat')
  })

  it('allows a dot in the name', () => {
    expect(validateIdpForm(samlBase({ name: 'acme.entra' })).name).toBeUndefined()
  })

  it('does not block saving an edit over a legacy name with a space', () => {
    // The name input is disabled on edit, so its format must not lock the form out.
    expect(validateIdpForm(samlBase({ editing: true, name: 'legacy idp name' })).name).toBeUndefined()
  })
})
