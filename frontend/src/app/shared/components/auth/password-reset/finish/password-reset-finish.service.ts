import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';

import {SERVER_API_URL} from '../../../../../app.constants';

@Injectable({providedIn: 'root'})
export class PasswordResetFinishService {
  constructor(private http: HttpClient) {
  }

  save(keyAndPassword: any): Observable<any> {
    return this.http.post(SERVER_API_URL + 'api/account/reset-password/finish', keyAndPassword);
  }
}
