import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {createRequestOption} from '../../../../shared/util/request-util';
import {OpenvasOptionModel} from '../../../shared/model/assets/openvas-chart/openvas-option.model';


@Injectable({
  providedIn: 'root'
})
export class AssetHostDetailDashboardService {
  private resourceUrl = SERVER_API_URL + 'api/openvas/results/dashboard/';

  constructor(private http: HttpClient) {
  }

  resultsVulnerabilityWordCloud(req?: any): Observable<HttpResponse<OpenvasOptionModel>> {
    const options = createRequestOption(req);
    return this.http.get<OpenvasOptionModel>(this.resourceUrl + 'results-vulnerability-word-cloud', {
      params: options,
      observe: 'response'
    });
  }

  resultsBySeverityClass(req?: any): Observable<HttpResponse<OpenvasOptionModel>> {
    const options = createRequestOption(req);
    return this.http.get<OpenvasOptionModel>(this.resourceUrl + 'results-by-severity-class', {
      params: options,
      observe: 'response'
    });
  }
}
