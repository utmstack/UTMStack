import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';
import {createRequestOption} from '../../../../../shared/util/request-util';
import {ITarget} from '../../../../shared/model/target.model';

@Injectable({
  providedIn: 'root'
})
export class TargetService {
  private resourceUrl = SERVER_API_URL + 'api/openvas/targets';

  constructor(private http: HttpClient) {
  }

  create(target: any): Observable<HttpResponse<ITarget>> {
    return this.http.post<any>(this.resourceUrl, target, {observe: 'response'});
  }

  update(target: any): Observable<HttpResponse<any>> {
    return this.http.put<any>(this.resourceUrl, target, {observe: 'response'});
  }

  find(target: string): Observable<HttpResponse<ITarget>> {
    return this.http.get<ITarget>(`${this.resourceUrl}/${target}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + '/get-by-filters', {params: options, observe: 'response'});
  }

  delete(targetId: string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${targetId}`, {observe: 'response'});
  }
}
