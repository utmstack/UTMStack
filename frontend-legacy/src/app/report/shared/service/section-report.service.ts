import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {ReportSectionType} from '../type/report-section.type';

@Injectable({
  providedIn: 'root'
})
export class SectionReportService {
  public resourceUrl = SERVER_API_URL + 'api/utm-report-sections';

  // GET /api/GET /api/utm-reportSectionnologies
  constructor(private http: HttpClient) {
  }

  create(reportSection: any): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl, reportSection, {observe: 'response'});
  }

  update(reportSection: ReportSectionType): Observable<HttpResponse<ReportSectionType>> {
    return this.http.put<ReportSectionType>(this.resourceUrl, reportSection, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ReportSectionType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ReportSectionType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(reportSection: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${reportSection}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<ReportSectionType>> {
    return this.http.get<ReportSectionType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
