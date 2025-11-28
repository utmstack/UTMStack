import {Component, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Subject} from 'rxjs';
import {takeUntil} from 'rxjs/operators';
import {ADMIN_ROLE} from '../../../shared/constants/global.constant';
import {CheckLicenseService} from '../../../shared/services/license/check-license.service';
import {EnterpriseFeatures, VersionType, VersionTypeService} from '../../../shared/services/util/version-type.service';
import {isSubdomainOfUtmstack} from '../../../shared/util/url.util';
import {NgbModal} from "@ng-bootstrap/ng-bootstrap";

@Component({
  selector: 'app-management-builder-sidebar',
  templateUrl: './app-management-sidebar.component.html',
  styleUrls: ['./app-management-sidebar.component.scss']
})
export class AppManagementSidebarComponent implements OnInit, OnDestroy {
  adminAuth = ADMIN_ROLE;
  alertDocumentationRoute = '/app-management/settings/alert-documentation';
  rolloverRoute = '/app-management/settings/rollover';
  isFree: boolean;
  inSass: boolean;
  destroy$: Subject<void> = new Subject<void>();
  ModulesEnterprise = EnterpriseFeatures;
  versionType = VersionType;
  version: VersionType;

  constructor(public router: Router,
              public versionTypeService: VersionTypeService,
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
    this.versionTypeService.versionType$
      .pipe(takeUntil(this.destroy$))
      .subscribe(versionType => {
        console.log(versionType);
        this.version = versionType;
      });
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
