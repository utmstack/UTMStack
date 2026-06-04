import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../app.constants';
import {createRequestOption} from '../../../../../shared/util/request-util';
import {IScannerModel} from '../../../../shared/model/scanner.model';

@Injectable({
  providedIn: 'root'
})
export class ScannerService {
  private resourceUrl = SERVER_API_URL + 'api/openvas/scanners/get-by-filters';

  constructor(private http: HttpClient) {
  }

  query(req?: any): Observable<HttpResponse<IScannerModel[]>> {
    const options = createRequestOption(req);
    return this.http.get<IScannerModel[]>(this.resourceUrl, {params: options, observe: 'response'});
  }
}
