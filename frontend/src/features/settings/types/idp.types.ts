/* Mirrors backend iam domain.IdentityProviderConfig + dto.IdentityProviderRequest. */

export const PROVIDER_TYPES = ['saml', 'oidc', 'ldap'] as const
export type ProviderType = (typeof PROVIDER_TYPES)[number]

/** Only these redirect the browser. LDAP has no button on the login page: its
 * password goes into the ordinary form and the backend binds with it. */
export const REDIRECTING_PROVIDER_TYPES: readonly ProviderType[] = ['saml', 'oidc']

export interface SamlSettings {
  metadataUrl: string
  spEntityId: string
  spAcsUrl: string
  spCertificatePem: string
  /** Write-only. Omit on update to keep the stored key. */
  spPrivateKeyPem?: string
}

export interface OidcSettings {
  issuer: string
  clientId: string
  redirectUrl: string
  scopes?: string[]
  /** Write-only. Omit on update to keep the stored secret. */
  clientSecret?: string
}

export interface LdapSettings {
  host: string
  port: number
  startTls: boolean
  insecureTls?: boolean
  bindDn: string
  baseDn: string
  /** Must contain %s, where the address typed at login is placed. */
  userFilter: string
  emailAttribute: string
  nameAttribute: string
  groupAttribute: string
  /** Write-only. Omit on update to keep the stored password. */
  bindPassword?: string
}

export type ProviderSettings = SamlSettings | OidcSettings | LdapSettings

/** One line of "this directory group grants this role here". */
export interface GroupMapping {
  group: string
  roleId: string
}

interface Provisioning {
  jitProvisioning: boolean
  defaultRoleId?: string | null
  groupsAttribute?: string
  syncRolesOnLogin: boolean
}

export interface IdentityProvider extends Provisioning {
  id: string
  name: string
  providerType: ProviderType
  active: boolean
  /** The secret of whichever protocol this is never comes back. */
  settings: Record<string, unknown>
  createdAt: string
  updatedAt: string
}

export interface IdentityProviderRequest extends Provisioning {
  id?: string
  name: string
  providerType: ProviderType
  active: boolean
  settings: ProviderSettings
  groupMappings?: GroupMapping[]
}

export const EMPTY_SETTINGS: Record<ProviderType, ProviderSettings> = {
  saml: { metadataUrl: '', spEntityId: '', spAcsUrl: '', spCertificatePem: '' },
  oidc: { issuer: '', clientId: '', redirectUrl: '' },
  ldap: {
    host: '',
    port: 389,
    startTls: true,
    bindDn: '',
    baseDn: '',
    userFilter: '(mail=%s)',
    emailAttribute: 'mail',
    nameAttribute: 'displayName',
    groupAttribute: 'memberOf',
  },
}
