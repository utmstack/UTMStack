import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {UtmAlertType} from '../../../../shared/types/alert/utm-alert.type';


@Injectable({
  providedIn: 'root'
})
export class AlertSocAiService {

  public resourceUrl = SERVER_API_URL + 'api/soc-ai';

  constructor(private http: HttpClient) {
  }

  /**
   * Submit an alert for SOC-AI analysis
   * @param alert The complete alert object to analyze
   */
  analyzeAlert(alert: UtmAlertType): Observable<HttpResponse<any>> {
    return this.http.post<any>(this.resourceUrl + '/analyze', alert, {observe: 'response'});
  }

}
