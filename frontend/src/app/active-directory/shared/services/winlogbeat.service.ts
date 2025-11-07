import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {Event} from '../../../shared/types/event/event';
import {createRequestOption} from '../../../shared/util/request-util';


@Injectable({
  providedIn: 'root'
})
export class WinlogbeatService {
  public resourceUrl = SERVER_API_URL + 'api/winlogbeat-info-by-filter';

  constructor(private http: HttpClient) {
  }

  query(req?: any): Observable<HttpResponse<Event[]>> {
    const options = createRequestOption(req);
    return this.http.get<Event[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

}
