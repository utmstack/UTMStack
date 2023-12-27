import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {SectionConfigType} from '../../types/configuration/section-config.type';
import {createRequestOption} from '../../util/request-util';

@Injectable({
  providedIn: 'root'
})
export class UtmConfigSectionService {
  public resourceUrl = SERVER_API_URL + 'api/utm-configuration-sections';

  // GET /api/GET /api/utm-technologies
  constructor(private http: HttpClient) {
  }

  create(tech: any): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl, tech, {observe: 'response'});
  }

  update(tech: SectionConfigType): Observable<HttpResponse<SectionConfigType>> {
    return this.http.put<SectionConfigType>(this.resourceUrl, tech, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<SectionConfigType[]>> {
    const options = createRequestOption(req);
    return this.http.get<SectionConfigType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(sectionID: number | string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${sectionID}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<SectionConfigType>> {
    return this.http.get<SectionConfigType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
