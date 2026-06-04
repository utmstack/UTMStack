import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../util/request-util';


@Injectable({
  providedIn: 'root'
})
export class CheckLicenseService {
  public resourceUrl = SERVER_API_URL + 'api/licence';

  constructor(private http: HttpClient) {
  }


  activate(req?: any): Observable<HttpResponse<any>> {
    const options = createRequestOption(req);
    return this.http.get<any>(this.resourceUrl + '/v2/activate-licence', {params: options, observe: 'response'});
  }

  validateLicense(req?: any): Observable<HttpResponse<boolean>> {
    const options = createRequestOption(req);
    return this.http.get<boolean>(this.resourceUrl + '/v2/validate-licence', {params: options, observe: 'response'});
  }

  checkLicense(): Observable<HttpResponse<boolean>> {
    return this.http.get<boolean>(this.resourceUrl + '/v2/check-licence', {observe: 'response'});
  }


}
