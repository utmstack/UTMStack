import {HttpClient, HttpHeaders, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import { UtmModuleGroupType } from '../type/utm-module-group.type';


@Injectable({
  providedIn: 'root'
})
export class UtmModuleGroupService {
  public resourceUrl = SERVER_API_URL + 'api/utm-configuration-groups';

  constructor(private http: HttpClient) {
  }

  create(conf: any): Observable<HttpResponse<UtmModuleGroupType>> {
    return this.http.post<any>(this.resourceUrl, conf, {observe: 'response'});
  }

  update(conf: UtmModuleGroupType): Observable<HttpResponse<UtmModuleGroupType>> {
    return this.http.put<UtmModuleGroupType>(this.resourceUrl, conf, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<UtmModuleGroupType[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmModuleGroupType[]>(this.resourceUrl + '/module-groups', {params: options, observe: 'response'});
  }

  delete(conf: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl + '/delete-single-module-group?groupId='}${conf}`, {observe: 'response'});
  }

  find(id: number): Observable<HttpResponse<UtmModuleGroupType>> {
    return this.http.get<UtmModuleGroupType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }
}
