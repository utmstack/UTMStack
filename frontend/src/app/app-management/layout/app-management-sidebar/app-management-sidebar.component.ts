import {Component, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';;
import {NgxSpinnerService} from 'ngx-spinner';
import {Subject} from 'rxjs';
import {ADMIN_ROLE} from '../../../shared/constants/global.constant';
import {CheckLicenseService} from '../../../shared/services/license/check-license.service';
import {ModalVersionInfoService} from '../../../shared/services/version/modal-version-info.service';
import {VersionType} from '../../../shared/services/version/version-info.service';
import {isSubdomainOfUtmstack} from '../../../shared/util/url.util';

@Component({
  selector: 'app-management-builder-sidebar',
  templateUrl: './app-management-sidebar.component.html',
  styleUrls: ['./app-management-sidebar.component.scss']
})
export class AppManagementSidebarComponent implements OnInit, OnDestroy {
  adminAuth = ADMIN_ROLE;
  isFree: boolean;
  inSass: boolean;
  destroy$: Subject<void> = new Subject<void>();

  version: VersionType;

  constructor(public router: Router,
              public modalVersionInfoService: ModalVersionInfoService,
              private checkLicenseService: CheckLicenseService,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    this.inSass = isSubdomainOfUtmstack();
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

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
