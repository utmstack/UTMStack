import {HttpClient, HttpHeaders, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {IncidentNoteType} from '../../../shared/types/incident/incident-note.type';
import {UtmIncidentType} from '../../../shared/types/incident/utm-incident.type';


@Injectable({
  providedIn: 'root'
})
export class UtmIncidentNoteService {
  public resourceUrl = SERVER_API_URL + 'api/utm-incident-notes';

  constructor(private http: HttpClient) {
  }


  query(req?: any): Observable<HttpResponse<IncidentNoteType[]>> {
    const options = createRequestOption(req);
    return this.http.get<IncidentNoteType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  update(incident: UtmIncidentType): Observable<HttpResponse<IncidentNoteType>> {
    return this.http.put<IncidentNoteType>(this.resourceUrl, incident, {observe: 'response'});
  }


  create(incident: any): Observable<HttpResponse<IncidentNoteType>> {
    return this.http.post<IncidentNoteType>(this.resourceUrl, incident, {observe: 'response'});
  }


}
