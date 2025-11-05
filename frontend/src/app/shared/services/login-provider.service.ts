import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {SERVER_API_URL} from '../../app.constants';
import {createRequestOption} from '../util/request-util';

export enum ProviderType {
  GOOGLE = 'GOOGLE',
  AZURE = 'AZURE',
  OKTA = 'OKTA',
  GITHUB = 'GITHUB',
  AUTH0 = 'AUTH0',
  FACEBOOK = 'FACEBOOK',
  LINKEDIN = 'LINKEDIN'
}

export const PROVIDER_ICONS: Record<ProviderType, string> = {
  [ProviderType.GOOGLE]: 'fa-brands fa-google',
  [ProviderType.AZURE]: 'fa-brands fa-microsoft',
  [ProviderType.OKTA]: 'fa-solid fa-shield-halved', // No hay icono específico de Okta
  [ProviderType.GITHUB]: 'fa-brands fa-github',
  [ProviderType.AUTH0]: 'fa-solid fa-lock', // No hay icono específico de Auth0
  [ProviderType.FACEBOOK]: 'fa-brands fa-facebook',
  [ProviderType.LINKEDIN]: 'fa-brands fa-linkedin'
};

export interface IdentityProviderDto {
  id: number;
  name: string;
  providerType: ProviderType;
  authUri: string;
  tokenUri: string;
  redirectUri: string;
  scopes: string;
  allowedDomains?: string;
  active: boolean;
}


@Injectable({providedIn: 'root'})
export class LoginProviderService {
  serverApiUrl = SERVER_API_URL + 'api/utm-providers';

  constructor(private http: HttpClient) {}

  getAllProviders(request: any) {
    const params = createRequestOption(request);
    return this.http.get<IdentityProviderDto[]>(this.serverApiUrl, {params, observe: 'response'});
  }

  loginWithProvider(provider: string): void {
    console.log(`${SERVER_API_URL}oauth2/authorization/${provider}`);
    window.location.href = `${SERVER_API_URL}oauth2/authorization/${provider}`;
  }
}
