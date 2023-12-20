import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {NetScanSoftwares} from '../types/net-scan-softwares';

@Injectable({
  providedIn: 'root'
})
export class AssetSoftwareService {
  private resourceUrl = SERVER_API_URL + 'api/utm-softwares';

  constructor(private http: HttpClient) {
  }

  query(req?: any): Observable<HttpResponse<NetScanSoftwares[]>> {
    const options = createRequestOption(req);
    return this.http.get<NetScanSoftwares[]>(this.resourceUrl, {
      params: options,
      observe: 'response'
    });
  }

  create(software: any): Observable<HttpResponse<NetScanSoftwares>> {
    return this.http.post<NetScanSoftwares>(this.resourceUrl, software, {observe: 'response'});
  }
}
