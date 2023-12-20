import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';

@Injectable({
  providedIn: 'root'
})
export class EncryptService {
  public resourceUrl = SERVER_API_URL + 'api/encrypt';

  constructor(private http: HttpClient) {
  }

  encrypt(text: string): Observable<HttpResponse<any>> {
    return this.http.post(this.resourceUrl, text,
      {observe: 'response', responseType: 'text'});
  }

}
