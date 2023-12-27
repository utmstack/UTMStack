import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {AlertTags} from '../../../../shared/types/alert/alert-tag.type';
import {createRequestOption} from '../../../../shared/util/request-util';

@Injectable({
  providedIn: 'root'
})
export class AlertTagService {
  public resourceUrl = SERVER_API_URL + 'api/utm-alert-tags';

  constructor(private http: HttpClient) {
  }

  create(tag: AlertTags): Observable<HttpResponse<AlertTags>> {
    return this.http.post<AlertTags>(this.resourceUrl, tag, {observe: 'response'});
  }

  update(tag: AlertTags): Observable<HttpResponse<AlertTags>> {
    return this.http.put<AlertTags>(this.resourceUrl, tag, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<AlertTags[]>> {
    const options = createRequestOption(req);
    return this.http.get<AlertTags[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(tag: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${tag}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<AlertTags>> {
    return this.http.get<AlertTags>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  tagFilterList(req?: any): Observable<HttpResponse<any[]>> {
    const options = createRequestOption(req);
    return this.http.get<any[]>(this.resourceUrl + '/categoriesToShowOnFilters', {
      params: options,
      observe: 'response'
    });
  }

}
