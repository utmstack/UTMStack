import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {ComplianceStandardSectionType} from '../type/compliance-standard-section.type';

@Injectable({
  providedIn: 'root'
})
// GET /api/compliance/section
export class CpStandardSectionService {
  private resourceUrl = SERVER_API_URL + 'api/compliance/standard-section';

  constructor(private http: HttpClient) {
  }

  create(section: ComplianceStandardSectionType): Observable<HttpResponse<any>> {
    return this.http.post<ComplianceStandardSectionType>(this.resourceUrl, section, {observe: 'response'});
  }

  update(section: ComplianceStandardSectionType): Observable<HttpResponse<any>> {
    return this.http.put<ComplianceStandardSectionType>(this.resourceUrl, section, {observe: 'response'});
  }

  find(section: string): Observable<HttpResponse<ComplianceStandardSectionType>> {
    return this.http.get<ComplianceStandardSectionType>(`${this.resourceUrl}/${section}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ComplianceStandardSectionType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceStandardSectionType[]>(this.resourceUrl, {
      params: options,
      observe: 'response'
    });
  }

  // GET /api/compliance/standard-section/sections-with-reports
  queryWithReports(req?: any): Observable<HttpResponse<ComplianceStandardSectionType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceStandardSectionType[]>(this.resourceUrl + '/sections-with-reports', {
      params: options,
      observe: 'response'
    });
  }

  delete(section: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${section}`, {observe: 'response'});
  }
}
