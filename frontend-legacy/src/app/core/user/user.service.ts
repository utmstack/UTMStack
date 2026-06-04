import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';

import {SERVER_API_URL} from '../../app.constants';
import {createRequestOption} from '../../shared/util/request-util';
import {IUser, User} from './user.model';

@Injectable({providedIn: 'root'})
export class UserService {
  public resourceUrl = SERVER_API_URL + 'api/users';

  constructor(private http: HttpClient) {
  }

  create(user: IUser): Observable<HttpResponse<IUser>> {
    return this.http.post<IUser>(this.resourceUrl, user, {observe: 'response'});
  }

  update(user: IUser): Observable<HttpResponse<IUser>> {
    return this.http.put<IUser>(this.resourceUrl, user, {observe: 'response'});
  }

  find(login: string): Observable<HttpResponse<User>> {
    return this.http.get<User>(`${this.resourceUrl + '/filter'}/${login}`, {observe: 'response'});
  }

  query(req?: any): Observable<HttpResponse<IUser[]>> {
    const options = createRequestOption(req);
    return this.http.get<IUser[]>(this.resourceUrl, {params: options, observe: 'response'});
  }

  delete(login: string): Observable<HttpResponse<any>> {
    return this.http.delete(`${this.resourceUrl}/${login}`, {observe: 'response'});
  }

  authorities(): Observable<string[]> {
    return this.http.get<string[]>(SERVER_API_URL + 'api/users/authorities');
  }
}
