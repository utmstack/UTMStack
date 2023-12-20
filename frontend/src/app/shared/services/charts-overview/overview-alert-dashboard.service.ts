import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../util/request-util';


@Injectable({
  providedIn: 'root'
})
export class OverviewAlertDashboardService {

  public resourceUrl = SERVER_API_URL + 'api/overview/';

  constructor(private http: HttpClient) {
  }

  getCardAlertTodayWeek(req?: any, top?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'count-alerts-today-and-last-week', {
      params: options,
      observe: 'response'
    });
  }

  getCardAlertByStatus(req?: any, top?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'count-alerts-by-status', {
      params: options,
      observe: 'response'
    });
  }

  getDataTable(endpoint: string, req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + endpoint, {
      params: options,
      observe: 'response'
    });
  }

  getEventInTime(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'events-in-time', {
      params: options,
      observe: 'response'
    });
  }

  getDataPie(endpoint: string, req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + endpoint, {
      params: options,
      observe: 'response'
    });
  }

  getAlertByCategory(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'top-alerts-by-category', {
      params: options,
      observe: 'response'
    });
  }

  getEventByType(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + 'count-events-by-type', {
      params: options,
      observe: 'response'
    });
  }
}
