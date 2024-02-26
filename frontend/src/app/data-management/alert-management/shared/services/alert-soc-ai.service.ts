import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';


@Injectable({
  providedIn: 'root'
})
export class AlertSocAiService {

  public resourceUrl = SERVER_API_URL + 'api/soc-ai';

  constructor(private http: HttpClient) {
  }

  processAlertBySoc(alertId: string[]): Observable<HttpResponse<any>> {
    return this.http.post<HttpResponse<any>>(this.resourceUrl + '/alerts', alertId, {observe: 'response'});
  }

}
