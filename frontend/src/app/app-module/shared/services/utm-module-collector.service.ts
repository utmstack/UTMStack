import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {UtmModulesEnum} from '../enum/utm-module.enum';
import { UtmModuleGroupType } from '../type/utm-module-group.type';
import {UtmModuleCollectorType} from "../type/utm-module-collector.type";
import {UtmListCollectorType} from "../type/utm-list-collector-type";


@Injectable({
  providedIn: 'root'
})
export class UtmModuleCollectorService {
  public resourceUrl = SERVER_API_URL + 'api';

  constructor(private http: HttpClient) {
  }

  create(body: any): Observable<HttpResponse<UtmModuleGroupType>> {
    const options = createRequestOption(body.collector);
    return this.http.post<any>(`${this.resourceUrl}/collector-config/`, body.collectorConfig, {params: options, observe: 'response'});
  }

  update(conf: UtmModuleGroupType): Observable<HttpResponse<UtmModuleGroupType>> {
    return this.http.put<UtmModuleGroupType>(this.resourceUrl, conf, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<UtmListCollectorType>> {
    const options = createRequestOption(req);
    return this.http.get<UtmListCollectorType>(`${this.resourceUrl}/collectors-list`, {params: options, observe: 'response'});
  }

  delete(conf: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl + '/delete-single-module-group?groupId='}${conf}`, {observe: 'response'});
  }

  groups(collectorId: string): Observable<HttpResponse<UtmModuleGroupType[]>> {
    return this.http.get<UtmModuleGroupType[]>(`${this.resourceUrl}/groups-by-collectors/${collectorId}`, {observe: 'response'});
  }

  groupsWithCollectors() {
    return this.http.get<UtmModuleGroupType[]>(`${this.resourceUrl}/groups-with-collectors`, {observe: 'response'});
  }
  formatCollectorResponse(groups: UtmModuleGroupType[], collectors: UtmModuleCollectorType[]) {
    return  groups.reduce((accumulator, currentValue) => {
      const { collector } = currentValue;

      const existingGroup = accumulator.find(group => group.collector === this.getCollectorName(Number(collector), collectors));

      if (existingGroup) {
        existingGroup.groups.push(currentValue);
      } else {
        accumulator.push({
          collector: this.getCollectorName(Number(collector), collectors),
          groups: [currentValue]
        });
      }
      return accumulator;
    }, []);
  }

  private getCollectorName(id: number, collectors: UtmModuleCollectorType[]) {
    const collector = collectors.find(c => c.id === id);
    return collector ? collector.hostname : '';
  }

  updateGroup(asset: { assetGroupId: number, assetsIds: number[] }): Observable<HttpResponse<any>> {
    return this.http.put<any>(this.resourceUrl + '/updateGroup', asset, {observe: 'response'});
  }
}