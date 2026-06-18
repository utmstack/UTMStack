import {HttpClient, HttpParams} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {
  TeamUser,
  TeamUserCreatePayload,
  TeamUserListQuery,
  TeamUserListResponse,
  TeamUserUpdatePayload
} from '../domain/team-user.model';

@Injectable({providedIn: 'root'})
export class FederationTeamService {
  private readonly endpoint = SERVER_API_URL + 'api/v1/users';

  constructor(private http: HttpClient) {}

  list(query: TeamUserListQuery = {}): Observable<TeamUserListResponse> {
    let params = new HttpParams();
    if (query.page !== undefined) {
      params = params.set('page', String(query.page));
    }
    if (query.page_size !== undefined) {
      params = params.set('page_size', String(query.page_size));
    }
    if (query.search) {
      params = params.set('search', query.search);
    }
    return this.http.get<TeamUserListResponse>(this.endpoint, {params});
  }

  get(id: number): Observable<TeamUser> {
    return this.http.get<TeamUser>(`${this.endpoint}/${id}`);
  }

  create(payload: TeamUserCreatePayload): Observable<TeamUser> {
    return this.http.post<TeamUser>(this.endpoint, payload);
  }

  update(id: number, payload: TeamUserUpdatePayload): Observable<TeamUser> {
    return this.http.put<TeamUser>(`${this.endpoint}/${id}`, payload);
  }

  deactivate(id: number): Observable<void> {
    return this.http.delete<void>(`${this.endpoint}/${id}`);
  }

  resendInvite(id: number): Observable<void> {
    return this.http.post<void>(`${this.endpoint}/${id}/resend-invite`, {});
  }

  disableTfa(id: number): Observable<void> {
    return this.http.post<void>(`${this.endpoint}/${id}/tfa/disable`, {});
  }
}
