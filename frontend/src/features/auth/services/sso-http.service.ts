import { createApiClient } from '@/shared/lib/api-client'

const api = createApiClient()

/** What the login screen needs to draw one button per provider. Read before
 * anyone has signed in, so it carries nothing that is not already public. */
export interface PublicIdentityProvider {
  id: string
  name: string
  providerType: 'saml' | 'oidc' | 'ldap'
  /** Empty for a directory: it has nowhere to send the browser. */
  loginUrl: string
}

export const ssoHttpService = {
  /** Only the providers that redirect. LDAP never appears: its password goes
   * into the ordinary form and the backend binds with it, so there is nothing
   * to click — picking it only says which directory to bind against. */
  list: () => api.get<PublicIdentityProvider[]>('/idp-providers'),
}
