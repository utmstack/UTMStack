import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';

@Injectable({
  providedIn: 'root'
})
export class AlertOpenStatusService {

  public resourceUrl = SERVER_API_URL + 'api/utm-alerts/count-open-alerts';

  constructor(private http: HttpClient) {
  }

  getOpenAlert(): Observable<HttpResponse<number>> {
    return this.http.get<number>(this.resourceUrl, {observe: 'response'});
  }
}
