import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';


@Injectable({
  providedIn: 'root'
})
export class FederationConnectionService {
  public resourceUrl = SERVER_API_URL + 'api/federation-service';

  constructor(private http: HttpClient) {
  }

  activate(req: { activate: boolean, name?: string, domain?: string }): Observable<HttpResponse<string>> {
    return this.http.post(this.resourceUrl + '/activate', req, {observe: 'response', responseType: 'text'});
  }

  getToken(): Observable<HttpResponse<string>> {
    return this.http.get(this.resourceUrl + '/token', {observe: 'response', responseType: 'text'});
  }


  regenerateToken(): Observable<HttpResponse<string>> {
    return this.http.get(this.resourceUrl + '/generateApiToken', {observe: 'response', responseType: 'text'});
  }

}
