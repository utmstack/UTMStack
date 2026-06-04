import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {ComplianceStandardType} from '../type/compliance-standard.type';
import {RefreshDataService} from '../../../shared/services/util/refresh-data.service';
import {map} from "rxjs/operators";

@Injectable({
  providedIn: 'root'
})
// GET /api/compliance/standard
export class CpStandardService extends RefreshDataService<boolean, HttpResponse<ComplianceStandardType[]>> {
  private resourceUrl = SERVER_API_URL + 'api/compliance/standard';

  constructor(private http: HttpClient) {
    super();
  }

  fetchData(request: any): Observable<HttpResponse<ComplianceStandardType[]>> {
   return this.query(request);
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
    }).pipe(
      map((response) => {
        const data = response.body as ComplianceStandardType[];
        return response.clone({body: data.filter(s => s.id >= 500 ) || []});
      })
    );
  }

  delete(standard: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${standard}`, {observe: 'response'});
  }
}
