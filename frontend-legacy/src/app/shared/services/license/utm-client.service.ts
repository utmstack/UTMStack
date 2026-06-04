import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {UtmClientType} from '../../types/license/utm-client.type';
import {createRequestOption} from '../../util/request-util';


@Injectable({
  providedIn: 'root'
})
export class UtmClientService {
  public resourceUrl = SERVER_API_URL + 'api/utm-clients';

  constructor(private http: HttpClient) {
  }

  create(client: any): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl, client, {observe: 'response'});
  }

  update(client: UtmClientType): Observable<HttpResponse<UtmClientType>> {
    return this.http.put<UtmClientType>(this.resourceUrl, client, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<UtmClientType[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmClientType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(client: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${client}`, {observe: 'response'});
  }

  find(id: number): Observable<HttpResponse<UtmClientType>> {
    return this.http.get<UtmClientType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
