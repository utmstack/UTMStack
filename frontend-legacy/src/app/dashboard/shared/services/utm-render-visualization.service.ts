import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {UtmDashboardVisualizationType} from '../../../shared/chart/types/dashboard/utm-dashboard-visualization.type';
import {createRequestOption} from '../../../shared/util/request-util';
import {RefreshDataService} from '../../../shared/services/util/refresh-data.service';


@Injectable({
  providedIn: 'root'
})
export class UtmRenderVisualization extends RefreshDataService<boolean, HttpResponse<UtmDashboardVisualizationType[]>> {

  public resourceUrl = SERVER_API_URL + 'api/utm-dashboard-visualizations';

  constructor(private http: HttpClient) {
    super();
  }

  query(req?: any): Observable<HttpResponse<UtmDashboardVisualizationType[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmDashboardVisualizationType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  fetchData(request: any): Observable<HttpResponse<UtmDashboardVisualizationType[]>> {
    return this.query(request);
  }
}
