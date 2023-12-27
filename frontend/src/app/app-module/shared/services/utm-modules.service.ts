import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {UtmModulesEnum} from '../enum/utm-module.enum';
import {UtmModuleCheckType} from '../type/utm-module-check.type';
import {UtmModuleType} from '../type/utm-module.type';


@Injectable({
  providedIn: 'root'
})
export class UtmModulesService {
  private resourceUrl = SERVER_API_URL + 'api/utm-modules';

  constructor(private http: HttpClient) {
  }

  getModules(req): Observable<HttpResponse<UtmModuleType[]>> {
    const request = createRequestOption(req);
    return this.http.get<UtmModuleType[]>(this.resourceUrl, {params: request, observe: 'response'});
  }

  checkConfig(req): Observable<HttpResponse<UtmModuleCheckType>> {
    const request = createRequestOption(req);
    return this.http.get<UtmModuleCheckType>(this.resourceUrl + '/checkRequirements',
      {params: request, observe: 'response'});
  }

  getModulesDetails(req): Observable<HttpResponse<UtmModuleType>> {
    const request = createRequestOption(req);
    return this.http.get<UtmModuleType>(this.resourceUrl + '/moduleDetails', {params: request, observe: 'response'});
  }

  setModuleActivate(activationStatus: boolean, nameShort: UtmModulesEnum, serverId: number): Observable<any> {
    return this.http.put<any>(this.resourceUrl + '/activateDeactivate' +
      '?activationStatus=' + activationStatus + '&nameShort=' + nameShort + '&serverId=' + serverId, {observe: 'response'});
  }

  getModuleCategories(req): Observable<HttpResponse<string[]>> {
    const request = createRequestOption(req);
    return this.http.get<string[]>(this.resourceUrl + '/moduleCategories', {params: request, observe: 'response'});
  }

  isActive(module: UtmModulesEnum): Observable<HttpResponse<boolean>> {
    const request = createRequestOption({moduleName: module});
    return this.http.get<boolean>(this.resourceUrl + '/is-active', {params: request, observe: 'response'});
  }


}
