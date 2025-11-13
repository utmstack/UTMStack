export enum ProviderType {
  GOOGLE = 'GOOGLE',
  MICROSOFT = 'MICROSOFT'
}

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
  clientId: string;
  clientSecret: string;
  redirectUri: string;
  authUri: string;
  tokenUri: string;
  userInfoUri: string;
  jwksUri?: string;
  clientAuthMethod?: ClientAuthMethod;
  scopes: string;
  allowedDomains?: string;
  active: boolean;
  createdAt?: Date;
  updatedAt?: Date;
}
