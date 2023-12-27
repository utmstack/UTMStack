import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';
import {AlertReportType} from '../../../../../shared/types/alert/alert-report.type';
import {createRequestOption} from '../../../../../shared/util/request-util';

@Injectable({
  providedIn: 'root'
})
export class AlertReportService {
  public resourceUrl = SERVER_API_URL + 'api/utm-reports';

  constructor(private http: HttpClient) {
  }

  create(report: AlertReportType): Observable<HttpResponse<AlertReportType>> {
    return this.http.post<AlertReportType>(this.resourceUrl, report, {observe: 'response'});
  }

  update(report: AlertReportType): Observable<HttpResponse<AlertReportType>> {
    return this.http.put<AlertReportType>(this.resourceUrl, report, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<AlertReportType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AlertReportType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(report: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${report}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<AlertReportType>> {
    return this.http.get<AlertReportType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
