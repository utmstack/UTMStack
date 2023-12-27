import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {ComplianceScheduleType} from '../type/compliance-schedule.type';


@Injectable({
  providedIn: 'root'
})
export class ComplianceScheduleService {
  public resourceUrl = SERVER_API_URL + 'api';

  constructor(private http: HttpClient) {
  }

  create(complianceSchedule: ComplianceScheduleType): Observable<HttpResponse<ComplianceScheduleType>> {
    return this.http.post<ComplianceScheduleType>(`${this.resourceUrl}/compliance-report-schedules`,
      complianceSchedule, {observe: 'response'});
  }

  update(alert: ComplianceScheduleType): Observable<HttpResponse<ComplianceScheduleType>> {
    return this.http.put<ComplianceScheduleType>(`${this.resourceUrl}/compliance-report-schedules`,
      alert, {observe: 'response'});
  }

  find(id: number): Observable<HttpResponse<ComplianceScheduleType>> {
    return this.http.get<ComplianceScheduleType>(`${this.resourceUrl}/compliance-report-schedules-by-id/${id}`,
      {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ComplianceScheduleType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceScheduleType[]>(`${this.resourceUrl}/compliance-report-schedules-by-user`, {
      params: options, observe: 'response'
    });
  }

  delete(id: string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/compliance-report-schedules/${id}`,
      {observe: 'response'});
  }
}
