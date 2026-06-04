import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {createRequestOption} from '../../../../shared/util/request-util';
import {AdNotificationConfType} from '../type/ad-notification-conf.type';


@Injectable({
  providedIn: 'root'
})
export class AdNotificationConfService {
  public resourceUrl = SERVER_API_URL + 'api/ad/utm-ad-notification-confs';

  constructor(private http: HttpClient) {
  }

  create(notify: AdNotificationConfType): Observable<HttpResponse<AdNotificationConfType>> {
    return this.http.post<AdNotificationConfType>(this.resourceUrl, notify, {observe: 'response'});
  }

  update(notify: AdNotificationConfType): Observable<HttpResponse<AdNotificationConfType>> {
    return this.http.put<AdNotificationConfType>(this.resourceUrl, notify, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<AdNotificationConfType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AdNotificationConfType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(notify: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${notify}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<AdNotificationConfType>> {
    return this.http.get<AdNotificationConfType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
