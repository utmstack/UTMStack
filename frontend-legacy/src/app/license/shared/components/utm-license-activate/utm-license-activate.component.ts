import {HttpErrorResponse} from '@angular/common/http';
import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {LicenceChangeBehavior} from '../../../../shared/behaviors/licence-change.behavior';
import {CheckLicenseService} from '../../../../shared/services/license/check-license.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';

@Component({
  selector: 'app-utm-license-activate',
  templateUrl: './utm-license-activate.component.html',
  styleUrls: ['./utm-license-activate.component.scss']
})
export class UtmLicenseActivateComponent implements OnInit {
  @Output() licenseChecked = new EventEmitter<boolean>();
  licenseForm: FormGroup;
  checked = false;
  activated = false;
  validating = false;

  constructor(public inputClassResolve: InputClassResolve,
              private checkLicenseService: CheckLicenseService,
              public activeModal: NgbActiveModal,
              private licenceChangeBehavior: LicenceChangeBehavior,
              private toastService: UtmToastService,
              private fb: FormBuilder) {
  }

  ngOnInit() {
    this.licenseForm = this.fb.group({
      licenceKey: ['', [Validators.required]],
      email: ['', [Validators.required]],
    });
    if (!window.navigator.onLine) {
      this.toastService.showError('Connection error', 'To active your licence need internet connection');
    }
  }


  activate() {
    this.validating = true;
    this.checkLicenseService.activate(this.licenseForm.value).subscribe(responseActivate => {
      this.validating = false;
      this.licenseChecked.emit(true);
      this.licenceChangeBehavior.$licenceChange.next(true);
      this.toastService.showSuccessBottom('Your enterprise licence is activated and ready to use');
      this.activeModal.close();
    }, (error: HttpErrorResponse) => {
      this.validating = false;
      if (error.status === 404) {
        this.toastService.showError('Problem validating licence',
          'We found a problem while validate your licence, please contact the support team');
      } else {
        this.toastService.showError('Error validating licence', error.message);
      }
    });
  }

}
