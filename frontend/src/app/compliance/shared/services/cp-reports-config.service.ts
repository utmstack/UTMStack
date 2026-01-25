import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {RefreshDataService} from '../../../shared/services/util/refresh-data.service';
import {createRequestOption} from '../../../shared/util/request-util';
import {ComplianceReportConfigType} from '../type/compliance-report-config.type';
import {ComplianceReportType} from '../type/compliance-report.type';

export interface ReportParams  {
  template: ComplianceReportType;
  sectionId: number;
  standardId: number;
}

@Injectable({
  providedIn: 'root'
})
export class CpReportsConfigService extends RefreshDataService<{ sectionId: number,
                  loading: boolean, reportSelected: number, page?: number }, HttpResponse<ComplianceReportType[]>> {

  private resourceUrl = SERVER_API_URL + 'api/compliance/control-config';
  private loadReportSubject = new BehaviorSubject<ReportParams>(null);
  private onLoadReportNoteSubject = new BehaviorSubject<ComplianceReportType>(null);

  constructor(private http: HttpClient) {
    super();
  }

  create(report: ComplianceReportConfigType): Observable<HttpResponse<any>> {
    return this.http.post<ComplianceReportConfigType>(this.resourceUrl, report, {observe: 'response'});
  }

  import(reports: {
    override: boolean
    reports: ComplianceReportConfigType[]
  }): Observable<HttpResponse<any>> {
    return this.http.post<ComplianceReportConfigType>(this.resourceUrl + '/import', reports, {observe: 'response'});
  }

  update(report: ComplianceReportConfigType): Observable<HttpResponse<any>> {
    return this.http.put<ComplianceReportConfigType>(this.resourceUrl, report, {observe: 'response'});
  }

  find(report: number): Observable<HttpResponse<ComplianceReportConfigType>> {
    return this.http.get<ComplianceReportConfigType>(`${this.resourceUrl}/${report}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ComplianceReportConfigType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceReportConfigType[]>(this.resourceUrl, {
      params: options,
      observe: 'response'
    });
  }

  // GET /api/compliance/report-config/get-by-report
  queryByStandard(req?: any): Observable<HttpResponse<ComplianceReportConfigType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceReportConfigType[]>(this.resourceUrl + '/get-by-filters', {
      params: options,
      observe: 'response'
    });
  }

  delete(report: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${report}`, {observe: 'response'});
  }

  fetchData(request: any): Observable<HttpResponse<ComplianceReportType[]>> {
    return this.queryByStandard(request);
  }

}
