import {HttpClient, HttpHeaders} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../../app.constants';

@Injectable({
  providedIn: 'root'
})
export class OmpReportService {
  public resourceUrl = SERVER_API_URL + 'api/utm-reports';

  constructor(private http: HttpClient) {
  }


  public exportAssetsToPdf(req?: any, limit?: number): Observable<any> {
    const headers = new HttpHeaders();
    headers.append('Accept', 'application/octet-stream');
    return this.http.post(this.resourceUrl + '/reports/openvas-assets?limit=' + limit, req,
      {headers, responseType: 'blob'});
  }

  public exportAssetsResultToPdf(req?: any, limit?: number): Observable<Blob> {
    const headers = new HttpHeaders();
    headers.append('Accept', 'application/octet-stream');
    return this.http.post(this.resourceUrl + '/reports/openvas-assets-results?limit=' + limit, req,
      {headers, responseType: 'blob'});
  }

  public exportTaskResultToPdf(req?: any, limit?: number): Observable<Blob> {
    const headers = new HttpHeaders();
    headers.append('Accept', 'application/octet-stream');
    return this.http.post(this.resourceUrl + '/reports/openvas-task-results?limit=' + limit, req,
      {headers, responseType: 'blob'});
  }

}
