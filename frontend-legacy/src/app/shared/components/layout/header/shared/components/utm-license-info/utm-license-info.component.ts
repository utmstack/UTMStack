import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {LicenceChangeBehavior} from '../../../../../../behaviors/licence-change.behavior';
import {UtmClientService} from '../../../../../../services/license/utm-client.service';
import {isSubdomainOfUtmstack} from '../../../../../../util/url.util';

@Component({
  selector: 'app-utm-license-info',
  templateUrl: './utm-license-info.component.html',
  styleUrls: ['./utm-license-info.component.scss']
})
export class UtmLicenseInfoComponent implements OnInit {
  license: 'ENTERPRISE' | 'FREE';

  constructor(private spinner: NgxSpinnerService,
              private utmClientService: UtmClientService,
              private licenceChangeBehavior: LicenceChangeBehavior,
              public router: Router) {
  }

  ngOnInit() {
    this.getLicenseStatus();
    /*this.licenceChangeBehavior.$licenceChange.subscribe(licenceChange => {
      if (licenceChange) {
        this.getLicenseStatus();
      }
    });*/
  }

  getLicenseStatus() {
    if (isSubdomainOfUtmstack()) {
      this.license = 'ENTERPRISE';
    } else {
      this.utmClientService.query({size: 100, page: 0}).subscribe(response => {
        this.license = response.body[0].clientLicenceVerified ? 'ENTERPRISE' : 'FREE';
      });
    }
  }

  goToLicense() {
    this.spinner.show('loadingSpinner');
    this.router.navigate(['/app-management/settings/license/license-info']).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }
}
