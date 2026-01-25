import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {RefreshDataService} from '../../../shared/services/util/refresh-data.service';
import {createRequestOption} from '../../../shared/util/request-util';
import {ComplianceControlConfigType} from '../type/compliance-control-config.type';
import {ComplianceReportType} from '../type/compliance-report.type';

export interface ReportParams  {
  template: ComplianceReportType;
  sectionId: number;
  standardId: number;
}

@Injectable({
  providedIn: 'root'
})
export class CpControlConfigService extends RefreshDataService<{ sectionId: number,
                  loading: boolean, reportSelected: number, page?: number }, HttpResponse<ComplianceReportType[]>> {

  private resourceUrl = SERVER_API_URL + 'api/compliance/control-config';
  private loadReportSubject = new BehaviorSubject<ReportParams>(null);
  private onLoadReportNoteSubject = new BehaviorSubject<ComplianceReportType>(null);

  constructor(private http: HttpClient) {
    super();
  }

  create(control: ComplianceControlConfigType): Observable<HttpResponse<any>> {
    return this.http.post<ComplianceControlConfigType>(this.resourceUrl, control, {observe: 'response'});
  }

  import(reports: {
    override: boolean
    reports: ComplianceControlConfigType[]
  }): Observable<HttpResponse<any>> {
    return this.http.post<ComplianceControlConfigType>(this.resourceUrl + '/import', reports, {observe: 'response'});
  }

  update(control: ComplianceControlConfigType): Observable<HttpResponse<any>> {
    return this.http.put<ComplianceControlConfigType>(this.resourceUrl, control, {observe: 'response'});
  }

  find(control: number): Observable<HttpResponse<ComplianceControlConfigType>> {
    return this.http.get<ComplianceControlConfigType>(`${this.resourceUrl}/${control}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ComplianceControlConfigType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceControlConfigType[]>(this.resourceUrl, {
      params: options,
      observe: 'response'
    });
  }

  // GET /api/compliance/report-config/get-by-report
  queryByStandard(req?: any): Observable<HttpResponse<ComplianceControlConfigType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceControlConfigType[]>(this.resourceUrl + '/get-by-filters', {
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
