import {HttpClient, HttpHeaders, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {NetScanType} from '../types/net-scan.type';
import {RefreshDataService} from '../../../shared/services/util/refresh-data.service';


@Injectable({
  providedIn: 'root'
})
// GET /api/utm-network-scans/search-by-filters
export class UtmNetScanService extends RefreshDataService<boolean, HttpResponse<NetScanType[]>> {
  // GET /api/utm-network-scans
  public resourceUrl = SERVER_API_URL + 'api/utm-network-scans';

  constructor(private http: HttpClient) {
    super();
  }

  fetchData(request: any): Observable<HttpResponse<NetScanType[]>> {
    return this.query(request);
  }


  query(req?: any): Observable<HttpResponse<NetScanType[]>> {
    const options = createRequestOption(req);
    return this.http.get<NetScanType[]>(this.resourceUrl + '/search-by-filters', {params: options, observe: 'response'});
  }

  update(asset: NetScanType): Observable<HttpResponse<NetScanType>> {
    return this.http.put<NetScanType>(this.resourceUrl, asset, {observe: 'response'});
  }

  // POST /api/utm-network-scans/saveOrUpdateCustomAsset
  create(asset: any): Observable<HttpResponse<NetScanType>> {
    return this.http.post<NetScanType>(this.resourceUrl + '/saveOrUpdateCustomAsset', asset, {observe: 'response'});
  }

  getFieldValues(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + '/searchPropertyValues', {params: options, observe: 'response'});
  }

  findAssetByID(id: number): Observable<HttpResponse<NetScanType>> {
    return this.http.get<NetScanType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  countNewAssets(): Observable<HttpResponse<number>> {
    return this.http.get<number>(`${this.resourceUrl}/countNewAssets`, {observe: 'response'});
  }

  updateType(asset: { assetTypeId: number, assetsIds: number[] }): Observable<HttpResponse<any>> {
    return this.http.put<any>(this.resourceUrl + '/updateType', asset, {observe: 'response'});
  }

  updateGroup(asset: { assetGroupId: number, assetsIds: number[] }): Observable<HttpResponse<any>> {
    return this.http.put<any>(this.resourceUrl + '/updateGroup', asset, {observe: 'response'});
  }

  deleteCustomAsset(id: number): Observable<HttpResponse<any>> {
    return this.http.delete<any>(this.resourceUrl + '/deleteCustomAsset/' + id, {observe: 'response'});
  }

  public exportAssetsToPdf(req?: any): Observable<any> {
    const headers = new HttpHeaders();
    const options = createRequestOption(req);
    headers.append('Accept', 'application/octet-stream');
    return this.http.get(this.resourceUrl + '/getNetworkScanReport',
      {headers, responseType: 'blob', params: options, observe: 'response'});
  }

  getAssetsPlatforms(): Observable<HttpResponse<string[]>> {
    return this.http.get<string[]>(`${this.resourceUrl}/agent-os-platform`, {observe: 'response'});
  }

}
