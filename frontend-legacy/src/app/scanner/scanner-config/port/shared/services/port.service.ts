import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';
import {createRequestOption} from '../../../../../shared/util/request-util';
import {IPort} from '../../../../shared/model/port.model';

@Injectable({
  providedIn: 'root'
})
export class PortService {
  private resourceUrl = SERVER_API_URL + 'api/openvas/port-lists';

  constructor(private http: HttpClient) {
  }

  create(port: any): Observable<HttpResponse<IPort>> {
    return this.http.post<any>(this.resourceUrl, port, {observe: 'response'});
  }

  update(port: IPort): Observable<HttpResponse<IPort>> {
    return this.http.put<any>(this.resourceUrl, port, {observe: 'response'});
  }

  find(port: string): Observable<HttpResponse<IPort>> {
    return this.http.get<any>(`${this.resourceUrl}/${port}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(SERVER_API_URL + 'api/openvas/port-list/get-by-filters', {
      params: options,
      observe: 'response'
    });
  }

  delete(login: string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${login}`, {observe: 'response'});
  }
}
