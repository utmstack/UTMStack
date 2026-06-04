import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';
import {AppStatusType} from './app-status.type';

@Injectable({
  providedIn: 'root'
})
export class AppRestartApiService {
  resourceUrl = SERVER_API_URL + 'api/system-restart/';

  constructor(private httpClient: HttpClient) {
  }

  restartApi(): Observable<HttpResponse<any>> {
    return this.httpClient.get<any>(this.resourceUrl + 'restart', {observe: 'response'});
  }

  appStatus(): Observable<HttpResponse<AppStatusType>> {
    return this.httpClient.get<AppStatusType>(this.resourceUrl + 'status', {observe: 'response'});
  }
}
