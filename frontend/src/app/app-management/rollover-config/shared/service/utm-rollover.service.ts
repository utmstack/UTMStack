import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {UtmRolloverType} from '../type/utm-rollover.type';

@Injectable({
  providedIn: 'root'
})
export class UtmRolloverService {
  public resourceUrl = SERVER_API_URL + 'api/index-policy/policy';

  constructor(private http: HttpClient) {
  }

  update(rolloverConfig: UtmRolloverType): Observable<HttpResponse<UtmRolloverType>> {
    return this.http.put<UtmRolloverType>(this.resourceUrl, rolloverConfig, {observe: 'response'});
  }

  getRollover(): Observable<HttpResponse<UtmRolloverType>> {
    return this.http.get<UtmRolloverType>(this.resourceUrl, {observe: 'response'});
  }


}
