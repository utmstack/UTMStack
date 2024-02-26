import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../util/request-util';
import {IncidentVariableType} from '../../types/incident/incident-variable.type';

@Injectable({
  providedIn: 'root'
})
export class IncidentResponseVariableService {

  public resourceUrl = SERVER_API_URL + 'api/utm-incident-variables';

  constructor(private http: HttpClient) {
  }

  create(variable: IncidentVariableType): Observable<HttpResponse<IncidentVariableType>> {
    return this.http.post<IncidentVariableType>(this.resourceUrl, variable, {observe: 'response'});
  }

  update(variable: IncidentVariableType): Observable<HttpResponse<IncidentVariableType>> {
    return this.http.put<IncidentVariableType>(this.resourceUrl, variable, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<IncidentVariableType[]>> {
    const options = createRequestOption(req);
    return this.http.get<IncidentVariableType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(variable: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${variable}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<IncidentVariableType>> {
    return this.http.get<IncidentVariableType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }
}
