import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {SERVER_API_URL} from '../../app.constants';
import {createRequestOption} from '../util/request-util';
import {Observable} from "rxjs";

export enum ProviderType {
  GOOGLE = 'GOOGLE',
  AZURE = 'AZURE',
  OKTA = 'OKTA',
  GITHUB = 'GITHUB',
  AUTH0 = 'AUTH0',
  FACEBOOK = 'FACEBOOK',
  LINKEDIN = 'LINKEDIN'
}

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
