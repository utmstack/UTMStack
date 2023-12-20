import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';

import {SERVER_API_URL} from '../../app.constants';

@Injectable({providedIn: 'root'})
export class Register {
  constructor(private http: HttpClient) {
  }

  save(account: any): Observable<any> {
    return this.http.post(SERVER_API_URL + 'api/register', account);
  }
}
