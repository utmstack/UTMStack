import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {UtmIndexPattern} from '../../types/index-pattern/utm-index-pattern';
import {createRequestOption} from '../../util/request-util';

@Injectable({
  providedIn: 'root'
})
export class IndexPatternService {
  private resourceUrl = SERVER_API_URL + 'api/utm-index-patterns';

  constructor(private http: HttpClient) {
  }

  create(indexPattern: any): Observable<HttpResponse<UtmIndexPattern>> {
    return this.http.post<any>(this.resourceUrl, indexPattern, {observe: 'response'});
  }

  update(indexPattern: UtmIndexPattern): Observable<HttpResponse<UtmIndexPattern>> {
    return this.http.put<UtmIndexPattern>(this.resourceUrl, indexPattern, {observe: 'response'});
  }

  find(indexPattern: string): Observable<HttpResponse<UtmIndexPattern>> {
    return this.http.get<UtmIndexPattern>(`${this.resourceUrl}/${indexPattern}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<UtmIndexPattern[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmIndexPattern[]>(this.resourceUrl, {
      params: options,
      observe: 'response'
    });
  }

  delete(indexPattern: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${indexPattern}`, {observe: 'response'});
  }
}
