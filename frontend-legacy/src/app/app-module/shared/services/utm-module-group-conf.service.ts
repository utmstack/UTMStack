import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {UtmModuleGroupConfType} from '../type/utm-module-group-conf.type';


@Injectable({
  providedIn: 'root'
})
export class UtmModuleGroupConfService {
  public resourceUrl = SERVER_API_URL + 'api/module-group-configurations';

  constructor(private http: HttpClient) {
  }

  saveConfig(conf: any): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl + '/save-or-update', conf, {observe: 'response'});
  }

  update(conf: { keys: UtmModuleGroupConfType[], moduleId: number }): Observable<HttpResponse<any>> {
    return this.http.put<any>(this.resourceUrl + '/update', conf, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<UtmModuleGroupConfType[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmModuleGroupConfType[]>(this.resourceUrl + '/by-group-and-key', {params: options, observe: 'response'});
  }

  delete(conf: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${conf}`, {observe: 'response'});
  }

  find(id: number): Observable<HttpResponse<UtmModuleGroupConfType>> {
    return this.http.get<UtmModuleGroupConfType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
