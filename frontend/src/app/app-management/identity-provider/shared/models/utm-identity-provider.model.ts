export enum ProviderType {
  GOOGLE = 'GOOGLE',
  MICROSOFT = 'MICROSOFT',
  OKTA = 'OKTA',
  KEYCLOAK = 'KEYCLOAK',
  GENERIC = 'GENERIC'
}

export const PROVIDER_ICONS: Record<ProviderType, string> = {
  [ProviderType.GOOGLE]: 'fa-brands fa-google',
  [ProviderType.MICROSOFT]: 'fa-brands fa-microsoft',
  [ProviderType.OKTA]: 'fa-solid fa-shield',
  [ProviderType.KEYCLOAK]: 'fa-solid fa-key',
  [ProviderType.GENERIC]: 'fa-solid fa-lock'
};


export enum ClientAuthMethod {
  CLIENT_SECRET_BASIC = 'CLIENT_SECRET_BASIC',
  CLIENT_SECRET_POST = 'CLIENT_SECRET_POST',
  CLIENT_SECRET_JWT = 'CLIENT_SECRET_JWT',
  PRIVATE_KEY_JWT = 'PRIVATE_KEY_JWT'
}

export interface UtmIdentityProvider {
  id?: number;
  name: string;
  providerType: ProviderType;
  metadataUrl: string;
  spPrivateKeyPem: string;
  spCertificatePem: string;
  active: boolean;
  createdAt?: Date;
  updatedAt?: Date;
}

