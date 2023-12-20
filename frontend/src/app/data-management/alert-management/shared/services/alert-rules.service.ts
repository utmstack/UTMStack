import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {createRequestOption} from '../../../../shared/util/request-util';
import {AlertRuleType} from '../../alert-rules/alert-rule.type';


@Injectable({
  providedIn: 'root'
})
export class AlertRulesService {

  public resourceUrl = SERVER_API_URL + 'api/alert-tag-rules';

  constructor(private http: HttpClient) {
  }

  /**
   * Find alert-rules by filters
   */
  query(query: any): Observable<HttpResponse<AlertRuleType[]>> {
    const options = createRequestOption(query);
    return this.http.get<AlertRuleType[]>(this.resourceUrl , {params: options, observe: 'response'});
  }

  /**
   * Delete a rule
   */
  delete(id: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  create(rule: AlertRuleType): Observable<HttpResponse<AlertRuleType>> {
    return this.http.post<AlertRuleType>(this.resourceUrl, rule, {observe: 'response'});
  }

  update(rule: AlertRuleType): Observable<HttpResponse<AlertRuleType>> {
    return this.http.put<AlertRuleType>(this.resourceUrl, rule, {observe: 'response'});
  }

  getByIds(ids: number[]): Observable<HttpResponse<AlertRuleType[]>> {
    const options = createRequestOption({ids});
    return this.http.get<AlertRuleType[]>(this.resourceUrl + '/get-by-ids', {params: options, observe: 'response'});
  }

}
