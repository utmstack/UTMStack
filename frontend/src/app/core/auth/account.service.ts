import {HttpClient, HttpResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, Subject} from 'rxjs';

import {SERVER_API_URL} from '../../app.constants';
import {HttpCancelService} from '../../blocks/service/httpcancel.service';
import {Account} from '../user/account.model';
import {AuthServerProvider} from './auth-jwt.service';
import {extractQueryParamsForNavigation} from "../../shared/util/query-params-to-filter.util";
import {ADMIN_DEFAULT_EMAIL, ADMIN_ROLE} from "../../shared/constants/global.constant";
import {StateStorageService} from "./state-storage.service";
import {Router} from "@angular/router";
import {NgxSpinnerService} from "ngx-spinner";
import {UtmToastService} from "../../shared/alert/utm-toast.service";

@Injectable({providedIn: 'root'})
export class AccountService {
  private userIdentity: Account;
  private authenticated = false;
  private authenticationState = new Subject<any>();

  constructor(private http: HttpClient,
              private authServerProvider: AuthServerProvider,
              private httpCancelService: HttpCancelService,
              private stateStorageService: StateStorageService,
              private router: Router,
              private spinner: NgxSpinnerService,
              private utmToast: UtmToastService) {
  }

  fetch(): Observable<HttpResponse<Account>> {
    return this.http.get<Account>(SERVER_API_URL + 'api/account', {observe: 'response'});
  }

  save(account: any): Observable<HttpResponse<any>> {
    return this.http.post(SERVER_API_URL + 'api/account', account, {observe: 'response'});
  }

  checkPassword(password: string, uuid: string): Observable<HttpResponse<string>> {
    const sanitized_password = encodeURIComponent(password)
    return this.http.get(SERVER_API_URL + `api/check-credentials?password=${sanitized_password}&checkUUID=${uuid}`, {
      observe: 'response',
      responseType: 'text'
    });
  }

  authenticate(identity) {
    this.userIdentity = identity;
    this.authenticated = identity !== null;
    this.authenticationState.next(this.userIdentity);
  }

  hasAnyAuthority(authorities: string[]): boolean {
    if (!this.authenticated || !this.userIdentity || !this.userIdentity.authorities) {
      return false;
    }
    // tslint:disable-next-line:prefer-for-of
    for (let i = 0; i < authorities.length; i++) {
      if (this.userIdentity.authorities.includes(authorities[i])) {
        return true;
      }
    }

    return false;
  }

  hasAuthority(authority: string): Promise<boolean> {
    if (!this.authenticated) {
      return Promise.resolve(false);
    }

    return this.identity().then(
      id => {
        return Promise.resolve(id.authorities && id.authorities.includes(authority));
      },
      () => {
        return Promise.resolve(false);
      }
    );
  }

  identity(force?: boolean): Promise<any> {
    if (force) {
      this.userIdentity = undefined;
    }

    // check and see if we have retrieved the userIdentity data from the server.
    // if we have, reuse it by immediately resolving
    if (this.userIdentity) {
      this.authenticated = true;
      return Promise.resolve(this.userIdentity);
    }

    // retrieve the userIdentity data from the server, update the identity object, and then resolve.
    return this.fetch()
      .toPromise()
      .then(response => {
        const account = response.body;
        if (account) {
          this.userIdentity = account;
          this.authenticated = true;
        } else {
          this.userIdentity = null;
          this.authenticated = false;
        }
        this.authenticationState.next(this.userIdentity);
        return this.userIdentity;
      })
      .catch(err => {
        this.userIdentity = null;
        this.authenticated = false;
        this.authenticationState.next(this.userIdentity);
        this.authServerProvider.logout().subscribe(() => {
          this.httpCancelService.cancelPendingRequests();
          console.log('UTMStack 401');
        });
        return null;
      });
  }

  isAuthenticated(): boolean {
    return this.authenticated;
  }

  isIdentityResolved(): boolean {
    return this.userIdentity !== undefined;
  }

  getAuthenticationState(): Observable<any> {
    return this.authenticationState.asObservable();
  }

  getImageUrl(): string {
    return this.isIdentityResolved() ? this.userIdentity.imageUrl : null;
  }

  openvasID(): number {
    if (this.userIdentity) {
      return this.userIdentity.openvasUserID;
    }
  }

  startNavigation() {
    this.identity(true).then(account => {
      if (account) {
        const { path, queryParams } =
          extractQueryParamsForNavigation(this.stateStorageService.getUrl() ? this.stateStorageService.getUrl() : '' );
        if (path) {
          this.stateStorageService.resetPreviousUrl();
        }
        const redirectTo = (account.authorities.includes(ADMIN_ROLE) && account.email === ADMIN_DEFAULT_EMAIL)
          ? '/getting-started' : !!path ? path : '/dashboard/overview';
        console.log(redirectTo);
        this.router.navigate([redirectTo], {queryParams})
          .then(() => this.spinner.hide());
      } else {
        this.utmToast.showError('Login error', 'User without privileges.');
      }
    });
  }
}
