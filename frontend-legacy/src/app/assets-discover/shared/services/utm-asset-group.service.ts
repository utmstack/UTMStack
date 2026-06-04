import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {AssetGroupType} from '../../asset-groups/shared/type/asset-group.type';


@Injectable({
  providedIn: 'root'
})
export class UtmAssetGroupService {
  public resourceUrl = SERVER_API_URL + 'api/utm-asset-groups';

  constructor(private http: HttpClient) {
  }


  query(req?: any): Observable<HttpResponse<AssetGroupType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AssetGroupType[]>(this.resourceUrl + '/searchGroupsByFilter', {params: options, observe: 'response'});
  }

  create(group: any): Observable<HttpResponse<AssetGroupType>> {
    return this.http.post<AssetGroupType>(this.resourceUrl, group, {observe: 'response'});
  }

  update(group: any): Observable<HttpResponse<any>> {
    return this.http.put<any>(this.resourceUrl, group, {observe: 'response'});
  }

  find(group: string): Observable<HttpResponse<AssetGroupType>> {
    return this.http.get<AssetGroupType>(`${this.resourceUrl}/${group}`, {observe: 'response'});
  }

  delete(groupID: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${groupID}`, {observe: 'response'});
  }


}
