import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {VisualizationType} from '../../../../shared/chart/types/visualization.type';
import {createRequestOption} from '../../../../shared/util/request-util';

@Injectable({
  providedIn: 'root'
})
export class VisualizationService {
  private resourceUrl = SERVER_API_URL + 'api/utm-visualizations';

  constructor(private http: HttpClient) {
  }

  create(visualization: VisualizationType): Observable<HttpResponse<any>> {
    return this.http.post<VisualizationType>(this.resourceUrl, visualization, {observe: 'response'});
  }

  update(visualization: VisualizationType): Observable<HttpResponse<any>> {
    return this.http.put<VisualizationType>(this.resourceUrl, visualization, {observe: 'response'});
  }

  find(visualization: number): Observable<HttpResponse<VisualizationType>> {
    return this.http.get<VisualizationType>(`${this.resourceUrl}/${visualization}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<VisualizationType[]>> {
    const options = createRequestOption(req);
    return this.http.get<VisualizationType[]>(this.resourceUrl, {
      params: options,
      observe: 'response'
    });
  }

  delete(visualization: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${visualization}`, {observe: 'response'});
  }

  bulkDelete(visualizations: string): Observable<HttpResponse<any>> {
    return this.http.delete(this.resourceUrl + '/bulk-delete/' + visualizations, {observe: 'response'});
  }

  run(visualization: VisualizationType): Observable<HttpResponse<any>> {
    return this.http.post<VisualizationType>(this.resourceUrl + '/run', visualization, {observe: 'response'});
  }

  bulkImport(visualizations: { visualizations: VisualizationType[], override: boolean }) {
    return this.http.post<VisualizationType>(this.resourceUrl + '/batch', visualizations, {observe: 'response'});
  }
}
