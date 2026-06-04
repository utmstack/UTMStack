import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {AssetType} from '../types/asset-type';


@Injectable({
  providedIn: 'root'
})
export class UtmAssetTypeService {
  public resourceUrl = SERVER_API_URL + 'api/utm-asset-types/all';

  constructor(private http: HttpClient) {
  }


  query(req?: any): Observable<HttpResponse<AssetType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AssetType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }


}
