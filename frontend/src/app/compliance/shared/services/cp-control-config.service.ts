import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BehaviorSubject, Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {RefreshDataService} from '../../../shared/services/util/refresh-data.service';
import {createRequestOption} from '../../../shared/util/request-util';
import {ComplianceControlEvaluationHistoryResponse} from '../type/compliance-control-evaluation-history-response.type';
import {ComplianceControlLatestEvaluationType} from '../type/compliance-control-latest-evaluation.type';
import {ComplianceControlType} from '../type/compliance-control.type';

export interface ControlParams  { // TODO: ELENA para que
  template: ComplianceControlType;
  sectionId: number;
  standardId: number;
}

@Injectable({
  providedIn: 'root'
})
export class CpControlConfigService extends RefreshDataService<{ sectionId: number,
                  loading: boolean, reportSelected: number, page?: number }, HttpResponse<ComplianceControlType[]>> {

  private resourceUrl = SERVER_API_URL + 'api/compliance/control-config';
  private loadControlSubject = new BehaviorSubject<ControlParams>(null);
  readonly onLoadControl$ = this.loadControlSubject.asObservable();

  constructor(private http: HttpClient) {
    super();
  }

  create(control: ComplianceControlType): Observable<HttpResponse<any>> {
    return this.http.post<ComplianceControlType>(
      this.resourceUrl,
      control,
      {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ComplianceControlType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceControlType[]>(this.resourceUrl, {
      params: options,
      observe: 'response'
    });
  }


  update(control: ComplianceControlType): Observable<HttpResponse<any>> {
    return this.http.put<ComplianceControlType>(
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

  controlsBySection(req?: any): Observable<HttpResponse<ComplianceControlLatestEvaluationType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceControlType[]>(this.resourceUrl + '/get-by-section', {
      params: options,
      observe: 'response'
    });
  }

  fetchData(request: any): Observable<HttpResponse<ComplianceControlLatestEvaluationType[]>> {
    return this.controlsBySection(request);
  }

  evaluationsByControl(controlId: number, req?: any): Observable<HttpResponse<ComplianceControlEvaluationHistoryResponse>> {
    const options = createRequestOption(req); // TODO: ELENA valorar si dejo el req
    return this.http.get<ComplianceControlEvaluationHistoryResponse>(
      `${this.resourceUrl}/${controlId}/evaluations`,
      { params: options, observe: 'response' }
    );
  }

  loadReport(params: ControlParams) {
    this.loadControlSubject.next(params);
  }

}
