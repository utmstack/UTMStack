import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {LogAnalyzerQueryType} from '../type/log-analyzer-query.type';

@Injectable({
  providedIn: 'root'
})
export class LogAnalyzerQueryService {

  public resourceUrl = SERVER_API_URL + 'api/log-analyzer/queries';

  constructor(private http: HttpClient) {
  }

  create(category: LogAnalyzerQueryType): Observable<HttpResponse<LogAnalyzerQueryType>> {
    return this.http.post<LogAnalyzerQueryType>(this.resourceUrl, category, {observe: 'response'});
  }

  update(category: LogAnalyzerQueryType): Observable<HttpResponse<LogAnalyzerQueryType>> {
    return this.http.put<LogAnalyzerQueryType>(this.resourceUrl, category, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<LogAnalyzerQueryType[]>> {
    const options = createRequestOption(req);
    return this.http.get<LogAnalyzerQueryType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(category: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${category}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<LogAnalyzerQueryType>> {
    return this.http.get<LogAnalyzerQueryType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }
}
