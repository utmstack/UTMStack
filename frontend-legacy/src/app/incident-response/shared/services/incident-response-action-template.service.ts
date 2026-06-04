import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {ActionConditionalEnum} from '../component/action-conditional/action-conditional.component';
import {IncidentRuleType} from '../type/incident-rule.type';

export interface IncidentResponseActionTemplate {
  id: number;
  title: string;
  description: string;
  command: string;
  conditional?: { key: ActionConditionalEnum, value: string };
}

@Injectable({
  providedIn: 'root'
})
export class IncidentResponseActionTemplateService {

  public resourceUrl = SERVER_API_URL + 'api/utm-alert-response-action-templates';

  constructor(private http: HttpClient) {
  }
  query(req?: any): Observable<HttpResponse<IncidentResponseActionTemplate[]>> {
    const options = createRequestOption(req);
    return this.http.get<IncidentResponseActionTemplate[]>(this.resourceUrl, {params: options, observe: 'response'});
  }
}
