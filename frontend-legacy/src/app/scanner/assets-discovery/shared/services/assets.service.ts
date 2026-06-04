import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../app.constants';
import {createRequestOption} from '../../../../shared/util/request-util';

@Injectable({
  providedIn: 'root'
})
export class AssetsService {
  private resourceUrl = SERVER_API_URL + 'api/openvas/assets';

  constructor(private http: HttpClient) {
  }

  query(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + '/get-by-filters', {params: options, observe: 'response'});
  }
}
