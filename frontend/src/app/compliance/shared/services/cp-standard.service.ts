import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {ComplianceStandardType} from '../type/compliance-standard.type';

@Injectable({
  providedIn: 'root'
})
// GET /api/compliance/standard
export class CpStandardService {
  private resourceUrl = SERVER_API_URL + 'api/compliance/standard';

  constructor(private http: HttpClient) {
  }

  create(standard: ComplianceStandardType): Observable<HttpResponse<any>> {
    return this.http.post<ComplianceStandardType>(this.resourceUrl, standard, {observe: 'response'});
  }

  update(standard: ComplianceStandardType): Observable<HttpResponse<any>> {
    return this.http.put<ComplianceStandardType>(this.resourceUrl, standard, {observe: 'response'});
  }

  find(standard: string): Observable<HttpResponse<ComplianceStandardType>> {
    return this.http.get<ComplianceStandardType>(`${this.resourceUrl}/${standard}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<ComplianceStandardType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceStandardType[]>(this.resourceUrl, {
      params: options,
      observe: 'response'
    });
  }

  delete(standard: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${standard}`, {observe: 'response'});
  }
}
