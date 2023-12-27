import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';
import {createRequestOption} from '../../../../../shared/util/request-util';
import {ICredential} from '../../../../shared/model/credential.model';

@Injectable({
  providedIn: 'root'
})
export class CredentialService {
  // /api/openvas/credentials
  // GET /api/openvas/credentials/get-by-filters
  private resourceUrl = SERVER_API_URL + 'api/openvas/credentials';

  constructor(private http: HttpClient) {
  }

  create(credential: any): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl, credential, {observe: 'response'});
  }

  update(credential: ICredential): Observable<HttpResponse<ICredential>> {
    return this.http.put<ICredential>(this.resourceUrl, credential, {observe: 'response'});
  }

  find(port: string): Observable<HttpResponse<ICredential>> {
    return this.http.get<ICredential>(`${this.resourceUrl}/${port}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + '/get-by-filters', {params: options, observe: 'response'});
  }

  delete(credentialId: any): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${credentialId}`, {observe: 'response'});
  }
}
