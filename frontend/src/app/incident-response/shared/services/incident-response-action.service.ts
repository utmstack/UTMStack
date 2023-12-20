import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {IncidentActionType} from '../type/incident-action.type';

@Injectable({
  providedIn: 'root'
})
export class IncidentResponseActionService {

  public resourceUrl = SERVER_API_URL + 'api/utm-incident-actions';

  constructor(private http: HttpClient) {
  }

  create(action: IncidentActionType): Observable<HttpResponse<IncidentActionType>> {
    return this.http.post<IncidentActionType>(this.resourceUrl, action, {observe: 'response'});
  }

  update(action: IncidentActionType): Observable<HttpResponse<IncidentActionType>> {
    return this.http.put<IncidentActionType>(this.resourceUrl, action, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<IncidentActionType[]>> {
    const options = createRequestOption(req);
    return this.http.get<IncidentActionType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(action: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${action}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<IncidentActionType>> {
    return this.http.get<IncidentActionType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }
}
