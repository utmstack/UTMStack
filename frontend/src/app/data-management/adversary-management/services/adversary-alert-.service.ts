import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {AdversaryAlerts} from '../models';

@Injectable()
export class AdversaryAlertService {

  resourceUrl = SERVER_API_URL + 'api/adversary';

  constructor(private http: HttpClient) {}

  query(filters: any[]): Observable<AdversaryAlerts[]> {
    return this.http.post<AdversaryAlerts[]>(`${this.resourceUrl}/alerts`, filters);
  }
}
