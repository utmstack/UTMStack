import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {FederationInstance} from '../domain/federation-instance.model';

export interface FederationInstanceInput {
  name: string;
  baseUrl: string;
  apiKey?: string;
  tlsSkipVerify: boolean;
}

@Injectable({providedIn: 'root'})
export class FederationInstancesService {
  private readonly endpoint = SERVER_API_URL + 'api/v1/instances';

  constructor(private http: HttpClient) {}

  list(): Observable<FederationInstance[]> {
    return this.http.get<FederationInstance[]>(this.endpoint);
  }

  get(id: number): Observable<FederationInstance> {
    return this.http.get<FederationInstance>(`${this.endpoint}/${id}`);
  }

  create(payload: FederationInstanceInput): Observable<FederationInstance> {
    return this.http.post<FederationInstance>(this.endpoint, payload);
  }

  update(id: number, payload: FederationInstanceInput): Observable<FederationInstance> {
    return this.http.put<FederationInstance>(`${this.endpoint}/${id}`, payload);
  }

  delete(id: number): Observable<void> {
    return this.http.delete<void>(`${this.endpoint}/${id}`);
  }
}
