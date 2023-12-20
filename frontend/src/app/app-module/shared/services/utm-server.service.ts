import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {UtmServerType} from '../type/utm-server.type';

@Injectable({
  providedIn: 'root'
})
export class UtmServerService {
  public resourceUrl = SERVER_API_URL + 'api/utm-servers';

  constructor(private http: HttpClient) {
  }

  create(server: any): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl, server, {observe: 'response'});
  }

  update(server: UtmServerType): Observable<HttpResponse<UtmServerType>> {
    return this.http.put<UtmServerType>(this.resourceUrl, server, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<UtmServerType[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmServerType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(server: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${server}`, {observe: 'response'});
  }

  find(id: number): Observable<HttpResponse<UtmServerType>> {
    return this.http.get<UtmServerType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
