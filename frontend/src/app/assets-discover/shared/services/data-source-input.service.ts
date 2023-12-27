import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {UtmDataInputStatus} from '../types/data-source-input.type';

@Injectable({
  providedIn: 'root'
})
export class DataSourceInputService {
  serverApiUrl = SERVER_API_URL + 'api/utm-data-input-statuses';

  constructor(private httpClient: HttpClient) {
  }

  query(req): Observable<HttpResponse<UtmDataInputStatus[]>> {
    const options = createRequestOption(req);
    return this.httpClient.get<UtmDataInputStatus[]>(this.serverApiUrl, {params: options, observe: 'response'});
  }

  delete(id: string): Observable<HttpResponse<any>> {
    return this.httpClient.delete(`${this.serverApiUrl}/${id}`, {observe: 'response'});
  }
}
