import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {IncidentHistoryType} from '../../../shared/types/incident/incident-history.type';
import {createRequestOption} from '../../../shared/util/request-util';


@Injectable({
  providedIn: 'root'
})
export class UtmIncidentHistoryService {
  public resourceUrl = SERVER_API_URL + 'api/utm-incident-histories';

  constructor(private http: HttpClient) {
  }


  query(req?: any): Observable<HttpResponse<IncidentHistoryType[]>> {
    const options = createRequestOption(req);
    return this.http.get<IncidentHistoryType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }





}
