import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {ElasticHealthType} from '../../types/elasticsearch/elastic-health.type';

@Injectable({
  providedIn: 'root'
})
export class ElasticHealthService {
  public resourceUrl = SERVER_API_URL + 'api/elasticsearch/cluster/status';

  constructor(private http: HttpClient) {
  }

  queryHealth(): Observable<HttpResponse<ElasticHealthType>> {
    return this.http.get<ElasticHealthType>(this.resourceUrl, {
      observe: 'response'
    });
  }
}
