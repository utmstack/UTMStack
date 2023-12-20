import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {NetScanType} from '../../../assets-discover/shared/types/net-scan.type';
import {createRequestOption} from '../../../shared/util/request-util';
import {IncidentAssetType} from '../type/incident-asset.type';

@Injectable({
  providedIn: 'root'
})
export class IncidentResponseAssetService {

  public resourceUrl = SERVER_API_URL + 'api/utm-network-scans';

  constructor(private http: HttpClient) {
  }

  cantExecuteAction(hostname: string): Observable<HttpResponse<boolean>> {
    return this.http.get<boolean>(`${this.resourceUrl}/can-run-command?assetName=${hostname}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<NetScanType[]>> {
    const options = createRequestOption(req);
    return this.http.get<NetScanType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }
}
