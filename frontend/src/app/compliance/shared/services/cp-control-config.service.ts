import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {RefreshDataService} from '../../../shared/services/util/refresh-data.service';
import {createRequestOption} from '../../../shared/util/request-util';
import {ComplianceControlConfigType} from '../type/compliance-control-config.type';
import {ComplianceControlEvaluationType} from '../type/compliance-control-evaluation.type';
import {ComplianceControlEvaluationsType} from '../type/compliance-control-evaluations.type';
import {ComplianceReportType} from '../type/compliance-report.type';
import {ReportParams} from "./cp-reports.service";
import {ComplianceControlEvaluationsResponse} from "../type/compliance-control-evaluations-response.type";

export interface ControlParams  { // TODO: ELENA para que
  template: ComplianceControlConfigType;
  sectionId: number;
  standardId: number;
}

@Injectable({
  providedIn: 'root'
})
export class CpControlConfigService extends RefreshDataService<{ sectionId: number,
                  loading: boolean, reportSelected: number, page?: number }, HttpResponse<ComplianceReportType[]>> {

  private resourceUrl = SERVER_API_URL + 'api/compliance/control-config';
  private loadControlSubject = new BehaviorSubject<ControlParams>(null);
  readonly onLoadControl$ = this.loadControlSubject.asObservable();

  constructor(private http: HttpClient) {
    super();
  }

  create(control: ComplianceControlConfigType): Observable<HttpResponse<any>> {
    return this.http.post<ComplianceControlConfigType>(
      this.resourceUrl,
      control,
      {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ComplianceControlConfigType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceControlConfigType[]>(this.resourceUrl, {
      params: options,
      observe: 'response'
    });
  }


  update(control: ComplianceControlConfigType): Observable<HttpResponse<any>> {
    return this.http.put<ComplianceControlConfigType>(
      `${this.resourceUrl}/${control.id}`,
      control,
      { observe: 'response' }
    );
  }

  delete(control: number): Observable<HttpResponse<any>> {
    return this.http.delete(
      `${this.resourceUrl}/${control}`,
      {observe: 'response'}
    );
  }

  controlsBySection(req?: any): Observable<HttpResponse<ComplianceControlEvaluationType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceControlConfigType[]>(this.resourceUrl + '/get-by-section', {
      params: options,
      observe: 'response'
    });
  }

  fetchData(request: any): Observable<HttpResponse<ComplianceControlEvaluationType[]>> {
    return this.controlsBySection(request);
  }

  evaluationsByControl(controlId: number, req?: any): Observable<HttpResponse<ComplianceControlEvaluationsResponse>> {
    const options = createRequestOption(req); // TODO: ELENA valorar si dejo el req
    return this.http.get<ComplianceControlEvaluationsResponse>(
      `${this.resourceUrl}/${controlId}/evaluations`,
      { params: options, observe: 'response' }
    );
  }

  loadReport(params: ControlParams) {
    this.loadControlSubject.next(params);
  }

}
