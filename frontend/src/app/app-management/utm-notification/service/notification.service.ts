import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../app.constants';
import {createRequestOption} from '../../../shared/util/request-util';
import {UtmNotification} from '../models/utm-notification.model';

const URL = SERVER_API_URL + 'api/notifications';
@Injectable({
  providedIn: 'root'
})
export class NotificationService {
  constructor(private http: HttpClient) {
  }

  getAll(request: any): Observable<UtmNotification> {
    const req = createRequestOption(request);
    return this.http.get<UtmNotification>(`${URL}`, {params: req});
  }

  getUnreadNotificationCount(): Observable<number> {
    return this.http.get<number>(`${URL}/unread-count`);
  }

  updateNotificationReadStatus(id: number) {
    const options = createRequestOption({read: true});
    return this.http.put<UtmNotification>(`${URL}/${id}/read`, {}, {params: options});
  }
}