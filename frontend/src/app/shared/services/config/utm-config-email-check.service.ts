import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';

@Injectable({
  providedIn: 'root'
})
export class UtmConfigEmailCheckService {
  public resourceUrl = SERVER_API_URL + 'api/checkEmailConfiguration';

  constructor(private http: HttpClient) {
  }

  checkEmail(): Observable<HttpResponse<any>> {
    return this.http.get<any>(this.resourceUrl, {observe: 'response'});
  }


}
