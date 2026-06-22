import {HttpClient, HttpParams} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {CountAlertsByStatusEntry} from '../domain/count-alerts-by-status.model';

@Injectable({providedIn: 'root'})
export class FederationOverviewService {
  private readonly endpoint = SERVER_API_URL + 'api/v1/overview/count-alerts-by-status';

  constructor(private http: HttpClient) {}

  countAlertsByStatus(from: string, to: string): Observable<CountAlertsByStatusEntry[]> {
    const params = new HttpParams().set('from', from).set('to', to);
    return this.http.get<CountAlertsByStatusEntry[]>(this.endpoint, {params});
  }
}
