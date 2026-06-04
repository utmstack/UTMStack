import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot} from '@angular/router';
import {isSubdomainOfUtmstack} from '../../shared/util/url.util';

@Injectable({providedIn: 'root'})
export class SaasRouteAccessService implements CanActivate {
  constructor(
    private router: Router,
  ) {
  }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): boolean | Promise<boolean> {
    return this.checkActiveModule(state.url);
  }

  checkActiveModule(url: string): Promise<boolean> {
    return new Promise<boolean>(resolve => {
      if (!isSubdomainOfUtmstack()) {
        resolve(true);
      } else {
        this.router.navigate(['/app-management/settings/connection-key']);
        resolve(false);
      }
    });
  }
}
