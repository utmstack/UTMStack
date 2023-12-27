import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {IncidentJobType} from '../type/incident-job.type';

@Injectable({
  providedIn: 'root'
})
export class IncidentResponseJobService {

  public resourceUrl = SERVER_API_URL + 'api/utm-incident-jobs';

  constructor(private http: HttpClient) {
  }

  create(job: IncidentJobType): Observable<HttpResponse<IncidentJobType>> {
    return this.http.post<IncidentJobType>(this.resourceUrl, job, {observe: 'response'});
  }

  update(job: IncidentJobType): Observable<HttpResponse<IncidentJobType>> {
    return this.http.put<IncidentJobType>(this.resourceUrl, job, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<IncidentJobType[]>> {
    const options = createRequestOption(req);
    return this.http.get<IncidentJobType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(job: number): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${job}`, {observe: 'response'});
  }

  find(id: string): Observable<HttpResponse<IncidentJobType>> {
    return this.http.get<IncidentJobType>(`${this.resourceUrl}/${id}`, {observe: 'response'});
  }
}
