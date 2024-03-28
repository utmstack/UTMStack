import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {LicenceChangeBehavior} from '../../../shared/behaviors/licence-change.behavior';
import {ADMIN_ROLE} from '../../../shared/constants/global.constant';
import {CheckLicenseService} from '../../../shared/services/license/check-license.service';
import {isSubdomainOfUtmstack} from '../../../shared/util/url.util';

@Component({
  selector: 'app-management-builder-sidebar',
  templateUrl: './app-management-sidebar.component.html',
  styleUrls: ['./app-management-sidebar.component.scss']
})
export class AppManagementSidebarComponent implements OnInit {
  adminAuth = ADMIN_ROLE;
  alertDocumentationRoute = '/app-management/settings/alert-documentation';
  rolloverRoute = '/app-management/settings/rollover';
  isFree: boolean;
  inSass: boolean;

  constructor(public router: Router,
              private licenceChangeBehavior: LicenceChangeBehavior,
              private checkLicenseService: CheckLicenseService,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    this.inSass = isSubdomainOfUtmstack();
    /*if (!this.inSass) {
      this.updateView();
      this.licenceChangeBehavior.$licenceChange.subscribe(licenceChange => {
        if (licenceChange) {
          this.updateView();
        }
      });
    }*/
  }

  private updateView(): void {
    this.checkLicenseService.checkLicense().subscribe(response => {
      this.isFree = !response.body;
    });
  }

  isActive(url): boolean {
    return this.router.isActive(url, false);
  }

  navigateTo(link: string) {
    this.spinner.show('loadingSpinner');
    this.router.navigate([link]).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  checkLicenseAndNavigate(link: string) {
    this.spinner.show('licenseSpinner');
    this.router.navigate([link]).then(() => {
      this.spinner.hide('licenseSpinner');
    });
  }
}
