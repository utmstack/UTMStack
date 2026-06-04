import {HttpClient, HttpHeaders, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {TableBuilderResponseType} from '../../../shared/chart/types/response/table-builder-response.type';
import {createRequestOption} from '../../../shared/util/request-util';
import {ReportType} from '../type/report.type';


@Injectable({
  providedIn: 'root'
})
export class ReportService {
  public resourceUrl = SERVER_API_URL + 'api/utm-reports';

  constructor(private http: HttpClient) {
  }

  create(report: any): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl, report, {observe: 'response'});
  }

  update(report: ReportType): Observable<HttpResponse<ReportType>> {
    return this.http.put<ReportType>(this.resourceUrl, report, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ReportType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ReportType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(report: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${report}`, {observe: 'response'});
  }

  find(id: number): Observable<HttpResponse<ReportType>> {
    return this.http.get<ReportType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  public exportReportToPdfGET(req?: any, url?: string): Observable<any> {
    const headers = new HttpHeaders();
    const options = createRequestOption(req);
    headers.append('Accept', 'application/octet-stream');
    return this.http.get(SERVER_API_URL + url,
      {headers, responseType: 'blob', params: options, observe: 'response'});
  }

  public exportReportToPdfPOST(req?: any, url?: string): Observable<any> {
    const headers = new HttpHeaders();
    headers.append('Accept', 'application/octet-stream');
    return this.http.post(SERVER_API_URL + url, req,
      {headers, responseType: 'blob', observe: 'response'});
  }

  getReportListData(endpoint: string): Observable<HttpResponse<TableBuilderResponseType[]>> {
    return this.http.get<TableBuilderResponseType[]>(SERVER_API_URL + endpoint, {
      observe: 'response'
    });
  }

}
