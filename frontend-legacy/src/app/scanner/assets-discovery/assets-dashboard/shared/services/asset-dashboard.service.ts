import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';

import {createRequestOption} from '../../../../../shared/util/request-util';
import {OpenvasOptionModel} from '../../../../shared/model/assets/openvas-chart/openvas-option.model';

@Injectable({
  providedIn: 'root'
})
export class AssetDashboardService {
  public resourceUrl = SERVER_API_URL + 'api/openvas/assets/dashboard/';

  constructor(private http: HttpClient) {
  }

  hostsByModificationTime(req?: any): Observable<HttpResponse<OpenvasOptionModel>> {
    const options = createRequestOption(req);
    return this.http.get<OpenvasOptionModel>(this.resourceUrl + 'hosts-by-modification-time', {
      params: options,
      observe: 'response'
    });
  }

  hostsBySeverityClass(req?: any): Observable<HttpResponse<OpenvasOptionModel>> {
    const options = createRequestOption(req);
    return this.http.get<OpenvasOptionModel>(this.resourceUrl + 'hosts-by-severity-class', {
      params: options,
      observe: 'response'
    });
  }

  mostVulnerableHost(req?: any): Observable<HttpResponse<OpenvasOptionModel>> {
    const options = createRequestOption(req);
    return this.http.get<OpenvasOptionModel>(this.resourceUrl + 'most-vulnerable-hosts?top=20', {
      params: options,
      observe: 'response'
    });
  }

  operatingSystemsByVulnerabilityScore(req?: any): Observable<HttpResponse<OpenvasOptionModel>> {
    const options = createRequestOption(req);
    return this.http.get<OpenvasOptionModel>(this.resourceUrl + 'operating-systems-by-vulnerability-score', {
      params: options,
      observe: 'response'
    });
  }

  operatingSystemsBySeverityClass(req?: any): Observable<HttpResponse<OpenvasOptionModel>> {
    const options = createRequestOption(req);
    return this.http.get<OpenvasOptionModel>(this.resourceUrl + 'operating-systems-by-severity-class', {
      params: options,
      observe: 'response'
    });
  }
}

