/* Mirrors backend iam domain.IdentityProviderConfig + dto.IdentityProviderRequest.
   SAML-only. The SP private key is write-only (json:"-" on the response). */

export interface IdentityProvider {
  id: number
  name: string
  providerType: string
  metadataUrl: string
  spEntityId: string
  spAcsUrl: string
  spCertificatePem: string
  active: boolean
  createdAt: string
  updatedAt: string
}

export interface IdentityProviderRequest {
  id?: number
  name: string
  providerType: string
  metadataUrl: string
  spEntityId: string
  spAcsUrl: string
  // Required on create; on update leave empty/undefined to keep the stored key.
  spPrivateKeyPem?: string
  spCertificatePem: string
  active: boolean
}
