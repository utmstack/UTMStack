import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {SERVER_API_URL} from '../../../../../../../app.constants';
import {createRequestOption} from '../../../../../../util/request-util';
import {UtmNotification} from '../models/utm-notification.model';

const URL = SERVER_API_URL + 'api/notifications';
@Injectable()
export class NotificationService {
  constructor(private http: HttpClient) {
  }

  getAll(request: any): Observable<HttpResponse<UtmNotification>> {
    const req = createRequestOption(request);
    return this.http.get<UtmNotification>(`${URL}`, {params: req, observe: 'response'});
  }

  getUnreadNotificationCount(): Observable<number>{
    return this.http.get<number>(`${URL}/unread-count`);
  }
}
