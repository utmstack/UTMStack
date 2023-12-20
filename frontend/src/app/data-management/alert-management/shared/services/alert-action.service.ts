import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';


@Injectable({
  providedIn: 'root'
})
export class AlertActionService {

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

}
