import {Injectable} from '@angular/core';
import {ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot} from '@angular/router';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {CheckLicenseService} from '../../shared/services/license/check-license.service';

@Injectable({providedIn: 'root'})
export class EnterpriseRouteAccessService implements CanActivate {
  constructor(
    private router: Router,
    private utmToastService: UtmToastService,
    private checkLicenseService: CheckLicenseService
  ) {
  }

  canActivate(route: ActivatedRouteSnapshot,
              state: RouterStateSnapshot): boolean | Promise<boolean> {
    return this.checkLicense(state.url);
  }

  checkLicense(url: string): Promise<boolean> {
    return new Promise<boolean>(resolve => {
      if (!window.navigator.onLine) {
        resolve(false);
      }
      this.checkLicenseService.checkLicense().subscribe(checking => {
        if (!checking.body) {
          this.utmToastService.showInfo('Invalid license',
            'You need purchase a license to access to this section. If you already have one,' +
            ' please activate in license section. Otherwise, please contact the support team for more information');
          this.router.navigate(['/app-management/settings/license/license-info']);
        }
        resolve(checking.body);
      });
    });
  }
}
