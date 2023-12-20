import {HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {LocalStorageService, SessionStorageService} from 'ngx-webstorage';

import {Observable} from 'rxjs';

import {ACCESS_KEY, SERVER_API_URL, SESSION_AUTH_TOKEN} from '../../app.constants';

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private localStorage: LocalStorageService,
              private sessionStorage: SessionStorageService) {
  }

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    if (!request || !request.url || (/^http/.test(request.url) && !(SERVER_API_URL && request.url.startsWith(SERVER_API_URL)))) {
      return next.handle(request);
    }
    const sessionToken = this.sessionStorage.retrieve(SESSION_AUTH_TOKEN);
    const localStorageToken = this.localStorage.retrieve(SESSION_AUTH_TOKEN);
    const token = sessionToken || localStorageToken;

    const sessionKey = this.sessionStorage.retrieve(ACCESS_KEY);
    const localStorageKey = this.localStorage.retrieve(ACCESS_KEY);
    const key = sessionKey || localStorageKey;

    if (!!token) {
      request = request.clone({
        setHeaders: {
          Authorization: 'Bearer ' + token
        }
      });
    }

    if (!!key) {
      request = request.clone({
        setHeaders: {
          'Utm-Internal-Key': key
        }
      });
    }
    return next.handle(request);
  }
}
