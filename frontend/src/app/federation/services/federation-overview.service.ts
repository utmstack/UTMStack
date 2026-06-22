import {HttpClient, HttpParams} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../app.constants';
import {CountAlertsBySeverityEntry} from '../domain/count-alerts-by-severity.model';
import {CountAlertsByStatusEntry} from '../domain/count-alerts-by-status.model';

@Injectable({providedIn: 'root'})
export class FederationOverviewService {
  private readonly statusEndpoint = SERVER_API_URL + 'api/v1/overview/count-alerts-by-status';
  private readonly severityEndpoint = SERVER_API_URL + 'api/v1/overview/count-alerts-by-severity';

  constructor(private http: HttpClient) {}

  countAlertsByStatus(from: string, to: string): Observable<CountAlertsByStatusEntry[]> {
    const params = new HttpParams().set('from', from).set('to', to);
    return this.http.get<CountAlertsByStatusEntry[]>(this.statusEndpoint, {params});
  }

  countAlertsBySeverity(from: string, to: string): Observable<CountAlertsBySeverityEntry[]> {
    const params = new HttpParams().set('from', from).set('to', to);
    return this.http.get<CountAlertsBySeverityEntry[]>(this.severityEndpoint, {params});
  }
}
