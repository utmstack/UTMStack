import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {ISource} from './sources.model';

@Injectable({
  providedIn: 'root'
})
export class SourcesService {

  public resourceUrl = SERVER_API_URL + 'api/utm-alert-sources';

  constructor(private http: HttpClient) {
  }

  /**
   * Get all alert sources
   */
  findAllSources(): Observable<HttpResponse<ISource[]>> {
    return this.http.get<ISource[]>(`${this.resourceUrl}`, {observe: 'response'});
  }
}
