import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {createRequestOption} from '../../shared/util/request-util';
import {InputSourceDataType} from './input-source-data.type';

@Injectable({
  providedIn: 'root'
})
export class InputSourceMngService {
  serverApiUrl = SERVER_API_URL + 'api/utm-datasource-configs';

  constructor(private httpClient: HttpClient) {
  }

  query(req): Observable<HttpResponse<InputSourceDataType[]>> {
    const options = createRequestOption(req);
    return this.httpClient.get<InputSourceDataType[]>(this.serverApiUrl, {params: options, observe: 'response'});
  }

  update(changes: InputSourceDataType[]): Observable<HttpResponse<InputSourceDataType>> {
    return this.httpClient.put<InputSourceDataType>(this.serverApiUrl, changes, {observe: 'response'});
  }
}
