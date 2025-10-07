import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Router} from '@angular/router';
import {LocalStorageService, SessionStorageService} from 'ngx-webstorage';
import {Observable} from 'rxjs';
import {map} from 'rxjs/operators';

import {ACCESS_KEY, COOKIE_AUTH_TOKEN, SERVER_API_URL, SESSION_AUTH_TOKEN} from '../../app.constants';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {TfaMethod} from '../../shared/services/tfa/tfa.service';
import {CSRFService} from './csrf.service';

@Injectable({providedIn: 'root'})
export class AuthServerProvider {

  private _tfaMethod: TfaMethod;
  private _tfaExpireTimeInSec: number;

  constructor(private http: HttpClient,
              private $localStorage: LocalStorageService,
              private $cookie: CSRFService,
              private router: Router,
              private toast: UtmToastService,
              private $sessionStorage: SessionStorageService) {
  }

  getToken() {
    return this.$localStorage.retrieve(SESSION_AUTH_TOKEN) || this.$sessionStorage.retrieve(SESSION_AUTH_TOKEN);
  }

  login(credentials): Observable<any> {
    const data = {
      username: credentials.username,
      password: credentials.password,
      rememberMe: credentials.rememberMe
    };
    const authenticateSuccess = (resp: HttpResponse<any>) => {
      this.storeAuthenticationToken(resp.body.token);
      if (resp.body.method && resp.body.forceTfa) {
        this.tfaMethod = resp.body.method as TfaMethod;
        this.tfaExpireTimeInSec = resp.body.tfaExpiresInSeconds;
      }
      return resp.body;
    };
    return this.http.post(SERVER_API_URL + 'api/authenticate',
      data,
      {observe: 'response'}).pipe(map(authenticateSuccess.bind(this)));
  }

  renewCode(): Observable<any> {
    return this.http.get(SERVER_API_URL + 'api/tfa/refresh');
  }

  verifyCode(code): Observable<any> {
    const authenticateSuccess = (resp) => {
      this.storeAuthenticationToken(resp.body.id_token);
      return resp.body.authenticated;
    };
    return this.http.post(SERVER_API_URL + 'api/tfa/verify-code',
      code, {observe: 'response'}).pipe(map(authenticateSuccess.bind(this)));
  }

  loginWithToken(jwt, rememberMe) {
    if (jwt) {
      this.storeAuthenticationToken(jwt);
      return Promise.resolve(jwt);
    } else {
      return Promise.reject('auth-jwt-service Promise reject '); // Put appropriate error message here
    }
  }

  loginWithAccessKey(jwt, rememberMe) {
    if (jwt) {
      this.storeAccessKey(jwt);
      return Promise.resolve(jwt);
    } else {
      return Promise.reject('auth-jwt-service Promise reject '); // Put appropriate error message here
    }
  }

  storeAuthenticationToken(jwt) {
    this.$localStorage.store(SESSION_AUTH_TOKEN, jwt);
    this.$sessionStorage.store(SESSION_AUTH_TOKEN, jwt);
    this.$cookie.setCookie(COOKIE_AUTH_TOKEN, jwt);
  }

  storeAccessKey(key) {
    this.$localStorage.store(ACCESS_KEY, key);
    this.$sessionStorage.store(ACCESS_KEY, key);
    this.$cookie.setCookie(ACCESS_KEY, key);
  }

  logout(): Observable<any> {
    return new Observable(observer => {
      this.$localStorage.clear(SESSION_AUTH_TOKEN);
      this.$sessionStorage.clear(SESSION_AUTH_TOKEN);
      this.$cookie.deleteCookie(COOKIE_AUTH_TOKEN);
      this.router.navigate(['/']).then(() => {
        if (this.getToken()) {
          this.toast.showInfo('Error authenticating service', 'as been an error while trying to authenticate with the service account ' +
            'please try to login again. If the problem continues, contact the support team');
        }
      });
      observer.next();
      observer.complete();
    });
  }

  get tfaMethod(): TfaMethod {
    return this._tfaMethod;
  }

  set tfaMethod(ftaMethod: TfaMethod) {
    this._tfaMethod = ftaMethod;
  }

  get tfaExpireTimeInSec(): number {
    return this._tfaExpireTimeInSec;
  }

  set tfaExpireTimeInSec(tfaExpireTimeInSec: number) {
    this._tfaExpireTimeInSec = tfaExpireTimeInSec;
  }

}
