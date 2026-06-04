import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';
import {createRequestOption} from '../../../../../shared/util/request-util';
import {IScannerConfigModel} from '../../../../shared/model/scanner-config.model';

@Injectable({
  providedIn: 'root'
})
export class ConfigScannerService {
  private resourceUrl = SERVER_API_URL + 'api/openvas/configs/get-by-filters';

  constructor(private http: HttpClient) {
  }

  query(req?: any): Observable<HttpResponse<IScannerConfigModel[]>> {
    const options = createRequestOption(req);
    return this.http.get<IScannerConfigModel[]>(this.resourceUrl, {params: options, observe: 'response'});
  }
}
