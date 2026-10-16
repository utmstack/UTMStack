import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';

@Injectable({providedIn: 'root'})
export class FederationTfaService {
  private readonly completeEndpoint = SERVER_API_URL + 'api/v1/tfa/complete';
  private readonly disableEndpoint = SERVER_API_URL + 'api/v1/tfa/disable';
  private readonly currentUserInfo = SERVER_API_URL + 'api/v1/auth/me';

  constructor(private http: HttpClient) {}

  me(): Observable<HttpResponse<any>> {
    return this.http.get(this.currentUserInfo, {observe: 'response'});
  }

  completeConfiguration(): Observable<HttpResponse<any>> {
    return this.http.post(this.completeEndpoint, null, {observe: 'response'});
  }

  disableTfa(password: string): Observable<HttpResponse<any>> {
    const body = JSON.stringify({password: password});
    return this.http.post(this.disableEndpoint, body,
      {headers: {'Content-Type': 'application/json'}, observe: 'response'});
  }
}
