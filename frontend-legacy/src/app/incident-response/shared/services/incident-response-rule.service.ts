import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {IncidentRuleSelectType, IncidentRuleType} from '../type/incident-rule.type';

@Injectable({
  providedIn: 'root'
})
export class IncidentResponseRuleService {

  public resourceUrl = SERVER_API_URL + 'api/utm-alert-response-rules';

  constructor(private http: HttpClient) {
  }

  create(action: IncidentRuleType): Observable<HttpResponse<IncidentRuleType>> {
    return this.http.post<IncidentRuleType>(this.resourceUrl, action, {observe: 'response'});
  }

  update(action: IncidentRuleType): Observable<HttpResponse<IncidentRuleType>> {
    return this.http.put<IncidentRuleType>(this.resourceUrl, action, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<IncidentRuleType[]>> {
    const options = createRequestOption(req);
    return this.http.get<IncidentRuleType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(action: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${action}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<IncidentRuleType>> {
    return this.http.get<IncidentRuleType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  getSelectOptions(): Observable<HttpResponse<IncidentRuleSelectType>> {
    return this.http.get<IncidentRuleSelectType>(`${this.resourceUrl}/resolve-filter-values`, {observe: 'response'});
  }

}
