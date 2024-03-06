import {HttpErrorResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {FormBuilder, FormGroup, NgModel, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {AccountService} from '../../../../core/auth/account.service';
import {UtmToastService} from '../../../alert/utm-toast.service';
import {SAAS_DEFAULT_PASSWORD} from '../../../constants/global.constant';
import {InputClassResolve} from '../../../util/input-class-resolve';
import {isSubdomainOfUtmstack} from '../../../util/url.util';
import {PasswordService} from '../../layout/header/shared/notification/utm-notification-user-setting/password/password.service';

@Component({
  selector: 'app-utm-admin-change-email',
  templateUrl: './utm-admin-change-email.component.html',
  styleUrls: ['./utm-admin-change-email.component.css']
})
export class UtmAdminChangeEmailComponent implements OnInit {
  @Input() account: any;
  @Output() setupSuccess = new EventEmitter<boolean>();
  @Input() gettingStarted: boolean;
  @ViewChild('currentPasswordInput') currentPasswordInput: NgModel;
  saving = false;
  formEmail: FormGroup;
  doNotMatch: string;
  error: string;
  success: string;
  currentPassword: string;
  newPassword = '';
  confirmPassword: string;
  inSass: boolean;

  constructor(private fb: FormBuilder,
              public inputClassResolve: InputClassResolve,
              private accountService: AccountService,
              private activeModal: NgbActiveModal,
              private passwordService: PasswordService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.inSass = isSubdomainOfUtmstack();
    this.currentPassword = this.inSass ? SAAS_DEFAULT_PASSWORD : '';
    this.formEmail = this.fb.group({
      email: ['', [Validators.required, Validators.email]]
    });
  }

  saveAccount() {
    this.saving = true;
    this.account.email = this.formEmail.get('email').value;
    this.accountService.save(this.account).subscribe(() => {
      this.passwordService.save(this.newPassword, this.currentPassword).subscribe(
        () => {
          this.setupReady();
        },
        (error: HttpErrorResponse ) => {
          this.saving = false;
          if (error.error && error.error.detail && error.error.detail.includes('Incorrect password')) {
            this.currentPasswordInput.control.setErrors({wrongPassword: true});
          }
          this.toastService.showError('Error changing default password',
            'has been an error while trying to setup your account');
        }
      );
    }, () => {
      this.saving = false;
      this.toastService.showError('Error changing admin account', 'has been an error while trying to setup your account');
    });
  }

  setupReady() {
    this.saving = false;
    this.activeModal.close();
    this.toastService.showSuccessBottom('Your account is ready to use');
    this.setupSuccess.emit(true);
  }

  isValid() {
    if (this.inSass) {
      return this.formEmail.get('email').valid && this.newPassword !== ''
        && this.newPassword === this.confirmPassword && this.currentPassword !== '';
    } else {
      return this.formEmail.get('email').valid;
    }
  }
}
