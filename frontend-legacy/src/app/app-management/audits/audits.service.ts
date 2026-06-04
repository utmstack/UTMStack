import {HttpClient, HttpParams, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {createRequestOption} from '../../shared/util/request-util';
import {UtmAudit} from './shared/type/audit.model';


@Injectable({providedIn: 'root'})
export class UtmAuditsService {
  constructor(private http: HttpClient) {
  }

  query(req: any): Observable<HttpResponse<UtmAudit[]>> {
    const params: HttpParams = createRequestOption(req);
    params.set('fromDate', req.fromDate);
    params.set('toDate', req.toDate);

    const requestURL = SERVER_API_URL + 'management/audits';

    return this.http.get<UtmAudit[]>(requestURL, {
      params,
      observe: 'response'
    });
  }
}
