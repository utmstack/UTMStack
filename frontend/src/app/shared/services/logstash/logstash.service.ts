import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {LogstashFilterType} from '../../../logstash/shared/types/logstash-filter.type';
import {UtmLogstashPipeline} from '../../../logstash/shared/types/logstash-stats.type';
import {createRequestOption} from '../../util/request-util';


@Injectable({
  providedIn: 'root'
})
export class LogstashService {
  public logstashFilterResourceUrl = SERVER_API_URL + 'api/logstash-filters';

  constructor(private http: HttpClient) {
  }

  create(filter: any, params?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(params);
    return this.http.post<any>(this.logstashFilterResourceUrl, filter, {params: options, observe: 'response'});
  }

  update(filter: LogstashFilterType): Observable<HttpResponse<LogstashFilterType>> {
    return this.http.put<LogstashFilterType>(this.logstashFilterResourceUrl, filter, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<LogstashFilterType[]>> {
    const options = createRequestOption(req);
    return this.http.get<LogstashFilterType[]>(this.logstashFilterResourceUrl + '/by-pipelineid', {params: options, observe: 'response'});
  }

  delete(filter: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.logstashFilterResourceUrl}/${filter}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<LogstashFilterType>> {
    return this.http.get<LogstashFilterType>(`${this.logstashFilterResourceUrl}/${id}`, {observe: 'response'});
  }

  getStats(): Observable<HttpResponse<UtmLogstashPipeline>> {
    return this.http.get<UtmLogstashPipeline>(SERVER_API_URL + 'api/logstash-pipelines/stats', {observe: 'response'});
  }
}
