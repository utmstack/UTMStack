import {HttpClient, HttpHeaders, HttpParams, HttpResponse} from '@angular/common/http';
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
         pattern: string, filters?: any, sortBy?: string, includeChildren?: boolean): Observable<HttpResponse<any>> {
    const query = new QueryType();
    query.add('page', page).add('size', size).add('top', top).add('indexPattern', pattern);

    if (includeChildren) {
      query.add('includeChildren', true);
    }

    if (sortBy) {
      query.add('sort', sortBy);
    }
    return this.http.post<HttpResponse<any>>(this.resourceUrl + 'search' + query.toString(),
      filters
      , {observe: 'response'});
  }

  searchBySqlQuery(sql: string, fetchSize: number, filters: any = {},
                   page: number, size: number): Observable<HttpResponse<any>> {

    const params = new HttpParams()
      .set('page', page.toString())
      .set('size', size.toString());

    const body = {
      query: sql,
      fetchSize: fetchSize,
      filters: filters
    };

    return this.http.post<any>(
      `${this.resourceUrl}search/sql`,
      body,
      { observe: 'response', params }
    );
  }


  exists(pattern: string, filters?: any): Observable<boolean> {
    const query = new QueryType();
    query.add('indexPattern', pattern);

    return this.http.post<boolean>(this.resourceUrl + 'count' + query.toString(), filters);
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
