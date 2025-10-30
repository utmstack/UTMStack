import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {SERVER_API_URL} from '../../app.constants';
import {createRequestOption} from '../util/request-util';

@Injectable({providedIn: 'root'})
export class LoginProviderService {
  serverApiUrl = SERVER_API_URL + 'api/utm-providers';

  constructor(private http: HttpClient) {}

  getAllProviders(request: any) {
    const params = createRequestOption(request);
    return this.http.get(this.serverApiUrl, {params, observe: 'response'});
  }
}
