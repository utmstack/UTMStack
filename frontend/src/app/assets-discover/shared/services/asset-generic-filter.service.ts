import {HttpClient, HttpHeaders, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {RefreshDataService} from '../../../shared/services/util/refresh-data.service';
import {createRequestOption} from '../../../shared/util/request-util';
import {NetScanType} from '../types/net-scan.type';


@Injectable()
export class AssetGenericFilterService extends RefreshDataService<{fieldFilter: string, refresh: boolean}, HttpResponse<Array<[string, number]>>> {
  public resourceUrl = SERVER_API_URL + 'api/utm-network-scans';

  constructor(private http: HttpClient) {
    super();
  }

  fetchData(request: any): Observable<HttpResponse<Array<[string, number]>>> {
    return this.getFieldValues(request);
  }

  getFieldValues(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + '/searchPropertyValues', {params: options, observe: 'response'});
  }
}
