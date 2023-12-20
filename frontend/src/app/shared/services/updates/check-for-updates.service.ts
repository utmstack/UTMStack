import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {VersionInfo} from '../../types/updates/updates.type';


@Injectable({
  providedIn: 'root'
})
export class CheckForUpdatesService {
  public resourceUrl = SERVER_API_URL + 'api/';
  public resourceInfo = SERVER_API_URL + 'management/';

  constructor(private http: HttpClient) {
  }

  getVersion(): Observable<HttpResponse<VersionInfo>> {
    return this.http.get<VersionInfo>(this.resourceInfo + 'info', {observe: 'response'});
  }

  getMode(): Observable<HttpResponse<boolean>> {
    return this.http.get<boolean>(this.resourceUrl + 'isInDevelop', {observe: 'response'});
  }


}
