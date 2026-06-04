import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {UtmIdentityProvider} from '../../app-management/identity-provider/shared/models/utm-identity-provider.model';
import {SERVER_API_URL} from '../../app.constants';
import {createRequestOption} from '../util/request-util';

@Injectable({providedIn: 'root'})
export class LoginProviderService {
  serverApiUrl = SERVER_API_URL + 'api/utm-providers';

  constructor(private http: HttpClient) {}

  getProviders(request: any) {
    const params = createRequestOption(request);
    return this.http.get<UtmIdentityProvider[]>(this.serverApiUrl, {params, observe: 'response'});
  }

  loginWithProvider(provider: string): void {
    window.location.href = `${SERVER_API_URL}saml2/authenticate/${provider}`;
  }
}
