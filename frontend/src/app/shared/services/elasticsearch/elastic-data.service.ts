import {HttpClient, HttpHeaders, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {ElasticFilterType} from '../../types/filter/elastic-filter.type';
import {QueryType} from '../../types/query-type';

@Injectable({
  providedIn: 'root'
})
export class ElasticDataService {
  public resourceUrl = SERVER_API_URL + 'api/elasticsearch/';

  constructor(private http: HttpClient) {
  }

  search(page: number, size: number, top: number,
         pattern: string, filters?: any, sortBy?: string): Observable<HttpResponse<any>> {
    const query = new QueryType();
    query.add('page', page).add('size', size).add('top', top).add('indexPattern', pattern);
    if (sortBy) {
      query.add('sort', sortBy);
    }
    return this.http.post<HttpResponse<any>>(this.resourceUrl + 'search' + query.toString(),
      filters
      , {observe: 'response'});
  }

  public exportToCsv(params): Observable<Blob> {
    const headers = new HttpHeaders();
    headers.append('Accept', 'application/octet-stream');
    return this.http.post(this.resourceUrl + 'search/csv', params,
      {headers, responseType: 'blob'});
  }

  genericSearchInIndex(req: { filters: ElasticFilterType[], index: string, top: number },
                       page: number, size: number, sort: string) {
    const params = new QueryType();
    params.add('page', page).add('size', size);
    if (sort) {
      params.add('sort', sort);
    }
    return this.http.post<HttpResponse<any>>(this.resourceUrl + 'generic-search' + params.toString(), req, {observe: 'response'});
  }
}
