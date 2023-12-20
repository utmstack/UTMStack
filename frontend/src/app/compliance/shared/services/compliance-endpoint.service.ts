import {HttpClient, HttpHeaders} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';

@Injectable({
  providedIn: 'root'
})
export class ComplianceEndpointService {
  private resourceUrl = SERVER_API_URL;

  constructor(private http: HttpClient) {
  }

  getDataFromTemplate(endpoint: string, queryParams?: any, bodyParams?: any) {
    return this.http.get<any[]>(this.resourceUrl + endpoint + queryParams,
      {observe: 'response'});
  }

  postDataFromTemplate(endpoint: string, queryParams?: any, bodyParams?: any) {
    return this.http.post<any[]>(this.resourceUrl + endpoint + queryParams,
      {
        filters: bodyParams
      }, {
        observe: 'response'
      });
  }

  public exportToCsv(params, url): Observable<Blob> {
    const headers = new HttpHeaders();
    headers.append('Accept', 'application/octet-stream');
    return this.http.post(this.resourceUrl + url, params,
      {headers, responseType: 'blob'});
  }
}
