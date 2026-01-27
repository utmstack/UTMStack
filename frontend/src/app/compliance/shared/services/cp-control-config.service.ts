import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {RefreshDataService} from '../../../shared/services/util/refresh-data.service';
import {createRequestOption} from '../../../shared/util/request-util';
import {ComplianceControlConfigType} from '../type/compliance-control-config.type';
import {ComplianceReportType} from '../type/compliance-report.type';

@Injectable({
  providedIn: 'root'
})
export class CpControlConfigService extends RefreshDataService<{ sectionId: number,
                  loading: boolean, reportSelected: number, page?: number }, HttpResponse<ComplianceReportType[]>> {

  private resourceUrl = SERVER_API_URL + 'api/compliance/control-config';

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

  queryByStandard(req?: any): Observable<HttpResponse<ComplianceControlConfigType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ComplianceControlConfigType[]>(this.resourceUrl + '/get-by-filters', {
      params: options,
      observe: 'response'
    });
  }

  fetchData(request: any): Observable<HttpResponse<ComplianceReportType[]>> {
    return this.queryByStandard(request);
  }

}
