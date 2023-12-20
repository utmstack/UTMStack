import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {WinlogbeatEventType} from '../types/winlogbeat-event.type';


@Injectable({
  providedIn: 'root'
})
export class WinlogbeatService {
  public resourceUrl = SERVER_API_URL + 'api/winlogbeat-info-by-filter';

  constructor(private http: HttpClient) {
  }

  query(req?: any): Observable<HttpResponse<WinlogbeatEventType[]>> {
    const options = createRequestOption(req);
    return this.http.get<WinlogbeatEventType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

}
