import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {ElasticFilterType} from '../../../shared/types/filter/elastic-filter.type';

@Injectable({
  providedIn: 'root'
})
export class LogAnalyzerService {

  public resourceUrl = SERVER_API_URL + 'api/log-analyzer/';

  constructor(private http: HttpClient) {
  }

  getFieldTopValues(indexPattern: string,
                    fieldName: string,
                    top: number,
                    filters?: ElasticFilterType[],
                    sort?: string): Observable<any> {
    let url = this.resourceUrl + 'top-x-values/' + indexPattern + '/' + fieldName + '/' + top;
    if (sort) {
      url += '?sort=' + sort;
    }
    return this.http.post(url, filters);
  }

  chartView(req): Observable<any> {
    return this.http.post(this.resourceUrl + 'chart-view', req, {observe: 'response'});
  }
}
