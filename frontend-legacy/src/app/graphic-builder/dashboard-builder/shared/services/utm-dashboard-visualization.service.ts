import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {UtmDashboardVisualizationType} from '../../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {createRequestOption} from '../../../../shared/util/request-util';

@Injectable({
  providedIn: 'root'
})
export class UtmDashboardVisualizationService {

  public resourceUrl = SERVER_API_URL + 'api/utm-dashboard-visualizations';

  constructor(private http: HttpClient) {
  }

  create(dashboardVis: UtmDashboardVisualizationType): Observable<HttpResponse<UtmDashboardVisualizationType>> {
    return this.http.post<UtmDashboardVisualizationType>(this.resourceUrl, dashboardVis, {observe: 'response'});
  }

  update(alert: UtmDashboardVisualizationType): Observable<HttpResponse<UtmDashboardVisualizationType>> {
    return this.http.put<UtmDashboardVisualizationType>(this.resourceUrl, alert, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<UtmDashboardVisualizationType>> {
    return this.http.get<UtmDashboardVisualizationType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<UtmDashboardVisualizationType[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmDashboardVisualizationType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(id: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }
}
