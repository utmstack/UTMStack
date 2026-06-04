import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {UtmDashboardVisualizationType} from '../../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {UtmDashboardType} from '../../../../shared/chart/types/dashboard/utm-dashboard.type';
import {createRequestOption} from '../../../../shared/util/request-util';


@Injectable({
  providedIn: 'root'
})
export class UtmDashboardService {
  public resourceUrl = SERVER_API_URL + 'api/utm-dashboards';

  constructor(private http: HttpClient) {
  }

  create(dashboard: UtmDashboardType): Observable<HttpResponse<UtmDashboardType>> {
    return this.http.post<UtmDashboardType>(this.resourceUrl, dashboard, {observe: 'response'});
  }

  update(alert: UtmDashboardType): Observable<HttpResponse<UtmDashboardType>> {
    return this.http.put<UtmDashboardType>(this.resourceUrl, alert, {observe: 'response'});
  }

  find(id: number): Observable<HttpResponse<UtmDashboardType>> {
    return this.http.get<UtmDashboardType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<UtmDashboardType[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmDashboardType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(id: string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  import(dashboards: { dashboards: UtmDashboardVisualizationType[], override: boolean }):
    Observable<HttpResponse<UtmDashboardVisualizationType[]>> {
    return this.http.post<UtmDashboardVisualizationType[]>(this.resourceUrl + '/import', dashboards, {observe: 'response'});
  }

}
