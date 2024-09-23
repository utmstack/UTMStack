import {HttpClient, HttpHeaders, HttpParams, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, of} from 'rxjs';
import {catchError, delay} from 'rxjs/operators';
import {SERVER_API_URL} from '../../app.constants';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {RefreshDataService} from '../../shared/services/util/refresh-data.service';
import {createRequestOption} from '../../shared/util/request-util';
import {Mode, Rule} from '../models/rule.model';

const resourceUrl = `${SERVER_API_URL}api/correlation-rule`;

@Injectable()
export class RuleService extends RefreshDataService<boolean, HttpResponse<Rule[]>> {
  rules: Rule[] = [];

  constructor(private http: HttpClient,
              private utmToast: UtmToastService) {
    super();
  }

  fetchData(req: any): Observable<HttpResponse<Rule[]>> {
    return this.getRules(createRequestOption(req));
  }

  private getRules(options: HttpParams): Observable<HttpResponse<Rule[]>> {
    return this.http.get<Rule[]>(`${resourceUrl}/search-by-filters`, {params: options, observe: 'response'})
      .pipe(
        catchError((error: any): Observable<HttpResponse<Rule[]>> => {
          console.error('Error loading rules', error);
          this.utmToast.showError('Failed to fetch notifications',
            'An error occurred while fetching rules data.');
          return of(new HttpResponse({
            headers: new HttpHeaders({'X-Total-Count': '0'}),
            body: null
          }));
        })
      );
  }

  getRuleById(id: number): Observable<HttpResponse<Rule>> {
    return this.http.get<Rule>(`${resourceUrl}/${id}`, {observe: 'response'})
      .pipe(
        catchError((error: any): Observable<HttpResponse<Rule>> => {
          console.error('Error loading rules', error);
          return of(new HttpResponse({
            body: undefined
          }));
        })
      );
  }

  save(rule: Partial<Rule>): Observable<any> {
    return this.http.post(resourceUrl, rule, {observe: 'response'});
  }

  update(rule: Partial<Rule>): Observable<any> {
    return this.http.put(resourceUrl, rule, {observe: 'response'});
  }

  delete(id: number): Observable<HttpResponse<any>> {
    return this.http.delete<any>(`${resourceUrl}/${id}`, {observe: 'response'});
  }

  activeRule(req: any): Observable<any> {
    const options = createRequestOption(req);
    return this.http.put<any>(`${resourceUrl}/activate-deactivate`, {}, { params: options, observe: 'response'});
  }

  public saveRule(mode: Mode, rule: Partial<Rule>) {
    if (mode === 'ADD') {
      return this.save(rule);
    } else {
      return this.update(rule);
    }
  }
}
