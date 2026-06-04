import {Injectable, isDevMode} from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot} from '@angular/router';
import {ModalService} from '../modal/modal.service';

import {AccountService} from './account.service';
import {StateStorageService} from './state-storage.service';

@Injectable({providedIn: 'root'})
export class UserRouteAccessService implements CanActivate {
  constructor(
    private router: Router,
    private accountService: AccountService,
    private stateStorageService: StateStorageService
  ) {
  }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): boolean | Promise<boolean> {
    const authorities = route.data.authorities;
    // We need to call the checkLogin / and so the accountService.identity() function, to ensure,
    // that the client has a principal too, if they already logged in by the server.
    // This could happen on a page refresh.
    return this.checkLogin(authorities, state.url);
  }

  async checkLogin(authorities: string[], url: string): Promise<boolean> {
    const account = await this.accountService.identity();
    if (!authorities || authorities.length === 0) {
      return false;
    }
    if (account) {
      const hasAnyAuthority = this.accountService.hasAnyAuthority(authorities);
      if (hasAnyAuthority) {
        return true;
      }
      if (isDevMode()) {
        console.error('User has not any of required authorities: ', authorities);
      }
      return false;
    } else {
      this.stateStorageService.storeUrl(url);
      return false;
    }
  }
}
