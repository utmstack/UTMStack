import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {QueryType} from '../../../../shared/types/query-type';


@Injectable({
  providedIn: 'root'
})
export class AlertManagementService {

  public resourceUrl = SERVER_API_URL + 'api/utm-alerts';

  constructor(private http: HttpClient) {
  }

  updateAlertStatus(alertId: string[], status: number, statusObservation: string = ''): Observable<HttpResponse<any>> {
    const req = {
      alertIds: alertId,
      status,
      statusObservation
    };
    return this.http.post<HttpResponse<any>>(this.resourceUrl + '/status', req, {observe: 'response'});
  }

  updateAlertTags(alertId: string[], ptags: string[], createRule?: boolean): Observable<HttpResponse<any>> {
    const body = {
      alertIds: alertId,
      tags: ptags,
      createRule: createRule ? createRule : false,
    };
    return this.http.post<HttpResponse<any>>(this.resourceUrl + '/tags', body, {observe: 'response'});
  }

  automaticTags(alertId: string): Observable<HttpResponse<any>> {
    const query = new QueryType();
    query.add('alertId', alertId);
    return this.http.get<HttpResponse<any>>(this.resourceUrl + '/auto-tags' + query.toString(), {observe: 'response'});
  }

  updateNotes(alertId: string, notes: string) {
    const query = new QueryType();
    query.add('alertId', alertId);
    return this.http.post<HttpResponse<any>>(this.resourceUrl + '/notes' + query.toString(), notes, {observe: 'response'});
  }

  updateSolution(solution: { alertName: string, solution: string }) {
    return this.http.post<HttpResponse<any>>(this.resourceUrl + '/solution', solution, {observe: 'response'});
  }

  markAsIncident(eventIds: string[], incidentName: string = '', incidentId: number, incidentSource: string): Observable<HttpResponse<any>> {
    const req = {
      eventIds,
      incidentName,
      incidentId,
      incidentSource
    };
    return this.http.post<HttpResponse<any>>(this.resourceUrl + '/convert-to-incident', req, {observe: 'response'});
  }

}
