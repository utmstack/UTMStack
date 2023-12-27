import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {SectionConfigParamType} from '../../types/configuration/section-config-param.type';
import {createRequestOption} from '../../util/request-util';

@Injectable({
  providedIn: 'root'
})
export class UtmConfigParamsService {
  public resourceUrl = SERVER_API_URL + 'api/utm-configuration-parameters';

  constructor(private http: HttpClient) {
  }

  create(param: any): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl, param, {observe: 'response'});
  }

  update(param: SectionConfigParamType[]): Observable<HttpResponse<SectionConfigParamType[]>> {
    return this.http.put<SectionConfigParamType[]>(this.resourceUrl, param, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<SectionConfigParamType[]>> {
    const options = createRequestOption(req);
    return this.http.get<SectionConfigParamType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(param: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${param}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<SectionConfigParamType>> {
    return this.http.get<SectionConfigParamType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
