import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {AssetGroupType} from '../../../assets-discover/asset-groups/shared/type/asset-group.type';
import {createRequestOption} from '../../../shared/util/request-util';
import {UtmModulesEnum} from '../enum/utm-module.enum';
import {UtmListCollectorType} from '../type/utm-list-collector-type';
import {UtmModuleCollectorType} from '../type/utm-module-collector.type';
import { UtmModuleGroupType } from '../type/utm-module-group.type';


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

  bulkCreate(body: any): Observable<HttpResponse<{results: {collectorId: string, status: string}[]}>> {
    return this.http.post<any>(`${this.resourceUrl}/collectors-config/`, body, {observe: 'response'});
  }

  reset(body: any): Observable<HttpResponse<UtmModuleGroupType>> {
    return this.http.post<any>(`${this.resourceUrl}/collectors-config/`, body.collectorConfig, {observe: 'response'});
  }

  update(conf: UtmModuleGroupType): Observable<HttpResponse<UtmModuleGroupType>> {
    return this.http.put<UtmModuleGroupType>(this.resourceUrl, conf, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<UtmListCollectorType>> {
    const options = createRequestOption(req);
    return this.http.get<UtmListCollectorType>(`${this.resourceUrl}/collectors-list`, {params: options, observe: 'response'});
  }

  queryFilter(req?: any): Observable<HttpResponse<UtmModuleCollectorType[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmModuleCollectorType[]>(`${this.resourceUrl}/search-by-filters`, {params: options, observe: 'response'});
  }

  delete(conf: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl + '/delete-single-module-group?groupId='}${conf}`, {observe: 'response'});
  }

  deleteCollector(id: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/collectors/${id}`, {observe: 'response'});
  }

  groups(collectorId: string): Observable<HttpResponse<UtmModuleGroupType[]>> {
    return this.http.get<UtmModuleGroupType[]>(`${this.resourceUrl}/groups-by-collectors/${collectorId}`, {observe: 'response'});
  }

  getCollectorGroupConfig(groups: UtmModuleGroupType[], collectors: UtmModuleCollectorType[]) {

    return  groups.reduce((accumulator, currentValue) => {
      const { collector } = currentValue;

      const existingGroup = accumulator.find(group => group.collector === this.getCollectorName(Number(collector), collectors));

      if (existingGroup) {
        existingGroup.groups.push(currentValue);
      } else {
        accumulator.push({
          id: Number(collector),
          collector: this.getCollectorName(Number(collector), collectors),
          groups: [currentValue]
        });
      }
      return accumulator;
    }, []);
  }

  getCollectors(groups: UtmModuleGroupType[], collectors: UtmModuleCollectorType[]) {

    return  collectors.map(collector  => {

      return {
          id: collector.id,
          collector: collector.hostname,
          groups: groups.filter( g => Number(g.collector) === collector.id)
        };
      });
  }

  private getCollectorName(id: number, collectors: UtmModuleCollectorType[]) {
    const collector = collectors.find(c => c.id === id);
    return collector ? collector.hostname : '';
  }

  updateGroup(asset: { assetGroupId: number, assetsIds: number[] }): Observable<HttpResponse<any>> {
    return this.http.put<any>(this.resourceUrl + '/updateGroup', asset, {observe: 'response'});
  }

  queryGroups(req?: any): Observable<HttpResponse<AssetGroupType[]>> {
    const options = createRequestOption(req);
    return this.http.get<AssetGroupType[]>(this.resourceUrl + '/searchGroupsByFilter', {params: options, observe: 'response'});
  }

  generateUniqueName(collectorName: string, groups: UtmModuleGroupType[]): string {
    const baseName = `Configuration-`;
    let count = 1;
    let newName = `${baseName}${count} ${collectorName}`;
    const existingNames = groups.map(group => group.groupName);

    while (existingNames.includes(newName)) {
      count++;
      newName = `${baseName}${count} ${collectorName}`;
    }

    return newName;
  }
}
