import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {TableBuilderResponseType} from '../../../shared/chart/types/response/table-builder-response.type';
import {createRequestOption} from '../../../shared/util/request-util';
import {ComplianceReportType} from '../type/compliance-report.type';

@Injectable({
  providedIn: 'root'
})
// GET /api/compliance/report
export class CpReportsService {
  private resourceUrl = SERVER_API_URL + 'api/compliance/report-config';

  constructor(private http: HttpClient) {
  }

  create(report: ComplianceReportType): Observable<HttpResponse<any>> {
    return this.http.post<ComplianceReportType>(this.resourceUrl, report, {observe: 'response'});
  }

  import(reports: {
    override: boolean
    reports: ComplianceReportType[]
  }): Observable<HttpResponse<any>> {
    return this.http.post<ComplianceReportType>(this.resourceUrl + '/import', reports, {observe: 'response'});
  }

  update(report: ComplianceReportType): Observable<HttpResponse<any>> {
    return this.http.put<ComplianceReportType>(this.resourceUrl, report, {observe: 'response'});
  }

  find(report: number): Observable<HttpResponse<ComplianceReportType>> {
    return this.http.get<ComplianceReportType>(`${this.resourceUrl}/${report}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ComplianceReportType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceReportType[]>(this.resourceUrl, {
      params: options,
      observe: 'response'
    });
  }

  // GET /api/compliance/report-config/get-by-report
  queryByStandard(req?: any): Observable<HttpResponse<ComplianceReportType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceReportType[]>(this.resourceUrl + '/get-by-filters', {
      params: options,
      observe: 'response'
    });
  }

  delete(report: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${report}`, {observe: 'response'});
  }

  getCustomReportData(endpoint: string): Observable<HttpResponse<TableBuilderResponseType[]>> {
    return this.http.get<TableBuilderResponseType[]>(SERVER_API_URL + endpoint, {
      observe: 'response'
    });
  }

}
