import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {IraHistoryType} from '../type/ira-history.type';

@Injectable({
  providedIn: 'root'
})
export class IncidentResponseRuleHistoryService {

  public resourceUrl = SERVER_API_URL + 'api/utm-alert-response-rule-histories';

  constructor(private http: HttpClient) {
  }

  query(req?: any): Observable<HttpResponse<IraHistoryType[]>> {
    const options = createRequestOption(req);
    return this.http.get<IraHistoryType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }


}
