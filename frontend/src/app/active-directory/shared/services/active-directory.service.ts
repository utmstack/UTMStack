import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {ActiveDirectoryType} from '../types/active-directory.type';

@Injectable({
  providedIn: 'root'
})
export class ActiveDirectoryService {
  public resourceUrl = SERVER_API_URL + 'api/ad/active-directory-info-by-filter';

  constructor(private http: HttpClient) {
  }

  query(req?: any): Observable<HttpResponse<ActiveDirectoryType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ActiveDirectoryType[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  queryUser(req?: any, urlSegment?: string): Observable<HttpResponse<ActiveDirectoryType[]>> {
    const options = createRequestOption(req);
    return this.http.get<ActiveDirectoryType[]>(SERVER_API_URL + urlSegment,
      {params: options, observe: 'response'});
  }
}
