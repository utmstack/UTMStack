import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';
import {AlertHistoryType} from '../../../../../shared/types/alert/alert-history.type';
import {createRequestOption} from '../../../../../shared/util/request-util';

@Injectable({
  providedIn: 'root'
})
export class AlertHistoryService {
  public resourceUrl = SERVER_API_URL + 'api/utm-alert-logs';

  constructor(private http: HttpClient) {
  }

  create(tag: AlertHistoryType): Observable<HttpResponse<AlertHistoryType>> {
    return this.http.post<AlertHistoryType>(this.resourceUrl, tag, {observe: 'response'});
  }

  update(tag: AlertHistoryType): Observable<HttpResponse<AlertHistoryType>> {
    return this.http.put<AlertHistoryType>(this.resourceUrl, tag, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<AlertHistoryType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AlertHistoryType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(tag: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${tag}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<AlertHistoryType>> {
    return this.http.get<AlertHistoryType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  tagFilterList(req?: any): Observable<HttpResponse<any[]>> {
    const options = createRequestOption(req);
    return this.http.get<any[]>(this.resourceUrl + '/categoriesToShowOnFilters', {
      params: options,
      observe: 'response'
    });
  }

}
