import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {AlertIncidentStatusUpdateType} from '../../types/incident/alert-incident-status-update.type';
import {IncidentAlertType} from '../../types/incident/incident-alert.type';
import {createRequestOption} from '../../util/request-util';


@Injectable({
  providedIn: 'root'
})
export class UtmIncidentAlertsService {
  public resourceUrl = SERVER_API_URL + 'api/utm-incident-alerts';

  constructor(private http: HttpClient) {
  }


  query(req?: any): Observable<HttpResponse<IncidentAlertType[]>> {
    const options = createRequestOption(req);
    return this.http.get<IncidentAlertType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }


  updateIncidentAlertStatus(alertUpdate: AlertIncidentStatusUpdateType): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl + '/update-status', alertUpdate, {observe: 'response'});
  }

}
