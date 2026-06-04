import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../../app.constants';
import {createRequestOption} from '../../../../../../shared/util/request-util';
import {PortRangeModel} from '../../../../../shared/model/port-range.model';

@Injectable({
  providedIn: 'root'
})
export class PortRangeService {
  private resourceUrl = SERVER_API_URL + 'api/openvas/port-ranges';

  constructor(private http: HttpClient) {
  }

  create(portRange: any): Observable<HttpResponse<PortRangeModel>> {
    return this.http.post<any>(this.resourceUrl, portRange, {observe: 'response'});
  }

  update(port: PortRangeModel): Observable<HttpResponse<PortRangeModel>> {
    return this.http.put<any>(this.resourceUrl, port, {observe: 'response'});
  }

  find(port: string): Observable<HttpResponse<PortRangeModel>> {
    return this.http.get<any>(`${this.resourceUrl}/${port}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + '/get-by-filters', {params: options, observe: 'response'});
  }

  delete(portRangeId: string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${portRangeId}`, {observe: 'response'});
  }


}
