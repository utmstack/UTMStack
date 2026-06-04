import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {createRequestOption} from '../../../../shared/util/request-util';
import {AdTrackerType} from '../type/ad-tracker.type';

@Injectable({
  providedIn: 'root'
})
export class AdTrackerService {
  public resourceUrl = SERVER_API_URL + 'api/ad/utm-ad-trackers';

  constructor(private http: HttpClient) {
  }

  create(tracker: any): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl, tracker, {observe: 'response'});
  }

  update(tracker: AdTrackerType): Observable<HttpResponse<AdTrackerType>> {
    return this.http.put<AdTrackerType>(this.resourceUrl, tracker, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<AdTrackerType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AdTrackerType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(tracker: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${tracker}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<AdTrackerType>> {
    return this.http.get<AdTrackerType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
