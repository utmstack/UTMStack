import {HttpClient, HttpHeaders, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {createRequestOption} from '../../../../shared/util/request-util';
import {AdReportType} from '../type/ad-report.type';

@Injectable({
  providedIn: 'root'
})
export class AdReportService {
  public resourceUrl = SERVER_API_URL + 'api/utm-ad-reports';

  constructor(private http: HttpClient) {
  }

  create(report: any): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl, report, {observe: 'response'});
  }

  update(report: AdReportType): Observable<HttpResponse<AdReportType>> {
    return this.http.put<AdReportType>(this.resourceUrl, report, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<AdReportType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AdReportType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(report: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${report}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<AdReportType>> {
    return this.http.get<AdReportType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  public exportToPdf(report): Observable<Blob> {
    const headers = new HttpHeaders();
    headers.append('Accept', 'application/octet-stream');
    return this.http.post(this.resourceUrl + '/build-report', report,
      {headers, responseType: 'blob'});
  }

}
