export enum ProviderType {
  GOOGLE = 'GOOGLE',
  MICROSOFT = 'MICROSOFT',
  OKTA = 'OKTA',
  KEYCLOAK = 'KEYCLOAK',
  GENERIC = 'GENERIC'
}

export const PROVIDER_ICONS: Record<ProviderType, string> = {
  [ProviderType.GOOGLE]: 'fa fa-brands fa-google',
  [ProviderType.MICROSOFT]: 'fa fa-brands fa-microsoft',
  [ProviderType.OKTA]: 'fa fa-solid fa-shield',
  [ProviderType.KEYCLOAK]: 'fa fa-solid fa-key',
  [ProviderType.GENERIC]: 'fa fa-solid fa-lock'
};

export interface UtmIdentityProvider {
  id?: number;
  name: string;
  spIdentityId: string;
  spAcsUrl: string;
  providerType: ProviderType;
  metadataUrl: string;
  spPrivateKeyPem: string;
  spCertificatePem: string;
  active: boolean;
  createdAt?: Date;
  updatedAt?: Date;
}

