import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {UtmIncidentType} from '../../types/incident/utm-incident.type';
import {createRequestOption} from '../../util/request-util';


@Injectable({
  providedIn: 'root'
})
export class UtmIncidentService {
  public resourceUrl = SERVER_API_URL + 'api/utm-incidents';

  constructor(private http: HttpClient) {
  }


  query(req?: any): Observable<HttpResponse<UtmIncidentType[]>> {
    const options = createRequestOption(req);
    return this.http.get<UtmIncidentType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  update(incident: UtmIncidentType): Observable<HttpResponse<UtmIncidentType>> {
    return this.http.put<UtmIncidentType>(this.resourceUrl, incident, {observe: 'response'});
  }

  changeStatus(incident: UtmIncidentType): Observable<HttpResponse<UtmIncidentType>> {
    return this.http.put<UtmIncidentType>(this.resourceUrl + '/change-status', incident, {observe: 'response'});
  }

  create(incident: any): Observable<HttpResponse<UtmIncidentType>> {
    return this.http.post<UtmIncidentType>(this.resourceUrl, incident, {observe: 'response'});
  }

  addAlerts(incident: any): Observable<HttpResponse<UtmIncidentType>> {
    return this.http.post<UtmIncidentType>(this.resourceUrl + '/add-alerts', incident, {observe: 'response'});
  }

  getUsersAssigned(): Observable<HttpResponse<{ id: number, login: string }[]>> {
    return this.http.get<{ id: number, login: string }[]>(this.resourceUrl + '/users-assigned', {observe: 'response'});
  }

  find(id: number): Observable<HttpResponse<UtmIncidentType>> {
    return this.http.get<UtmIncidentType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }

}
