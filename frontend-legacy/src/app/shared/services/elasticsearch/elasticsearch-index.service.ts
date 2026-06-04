import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {ElasticSearchFieldInfoType} from '../../types/elasticsearch/elastic-search-field-info.type';
import {ElasticsearchIndexType} from '../../types/elasticsearch/elasticsearch-index.type';
import {createRequestOption} from '../../util/request-util';

@Injectable({
  providedIn: 'root'
})
export class ElasticSearchIndexService {
  public resourceUrl = SERVER_API_URL + 'api/elasticsearch/';

  constructor(private http: HttpClient) {
  }

  getElasticIndex(req?: any): Observable<HttpResponse<ElasticsearchIndexType>> {
    const options = createRequestOption(req);
    return this.http.get<ElasticsearchIndexType>(this.resourceUrl + 'index/all', {
      params: options,
      observe: 'response'
    });
  }

  getElasticIndexField(req?: any): Observable<HttpResponse<ElasticSearchFieldInfoType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ElasticSearchFieldInfoType[]>(this.resourceUrl + 'index/properties', {
      params: options,
      observe: 'response'
    });
  }

  getElasticFieldValues(req?: any): Observable<HttpResponse<string[]>> {
    const options = createRequestOption(req);
    return this.http.get<string[]>(this.resourceUrl + 'property/values', {
      params: options,
      observe: 'response'
    });
  }

  getValuesWithCount(rq) {
    return this.http.post(this.resourceUrl + 'property/values-with-count', rq, {observe: 'response'});
  }

  deleteIndexes(indexes: string[]) {
    return this.http.post(this.resourceUrl + 'index/delete-index', indexes, {observe: 'response'});
  }

}
