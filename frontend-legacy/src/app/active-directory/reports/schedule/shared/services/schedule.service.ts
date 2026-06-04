import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';
import {ISchedule} from '../../../../../scanner/shared/model/schedule.model';
import {createRequestOption} from '../../../../../shared/util/request-util';

@Injectable({
  providedIn: 'root'
})
export class ScheduleService {
  private resourceUrl = SERVER_API_URL + 'api/openvas/schedules';

  constructor(private http: HttpClient) {
  }

  create(schedule: any): Observable<HttpResponse<ISchedule>> {
    return this.http.post<ISchedule>(this.resourceUrl, schedule, {observe: 'response'});
  }

  update(schedule: any): Observable<HttpResponse<any>> {
    return this.http.put<any>(this.resourceUrl, schedule, {observe: 'response'});
  }

  find(login: string): Observable<HttpResponse<ISchedule>> {
    return this.http.get<ISchedule>(`${this.resourceUrl}/${login}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + '/get-by-filters', {params: options, observe: 'response'});
  }

  delete(login: string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${login}`, {observe: 'response'});
  }
}
