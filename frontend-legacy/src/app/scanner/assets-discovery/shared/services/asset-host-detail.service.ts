import {HttpClient, HttpHeaders, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {ReportFormatEnum} from '../../../../shared/enums/report-format.enum';
import {createRequestOption} from '../../../../shared/util/request-util';

@Injectable({
  providedIn: 'root'
})
export class AssetHostDetailService {
  private resourceUrl = SERVER_API_URL + 'api/openvas/results';

  constructor(private http: HttpClient) {
  }

  query(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + '/get-by-filters', {params: options, observe: 'response'});
  }


  public exportVulnerabilitiesToPdf(fullReport?: any,
                                    lastReportUuid?: string,
                                    reportFormatName?: ReportFormatEnum): Observable<Blob> {
    const headers = new HttpHeaders();
    headers.append('Accept', 'application/octet-stream');
    return this.http.get(this.resourceUrl + '/get-result-report?fullReport=' + fullReport +
      '&lastReportUuid=' + lastReportUuid + '&reportFormatName=' + reportFormatName,
      {headers, responseType: 'blob'});
  }
}
