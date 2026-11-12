import {AfterViewInit, Component, ElementRef, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../alert/utm-toast.service';

import {PasswordResetInitService} from './password-reset-init.service';

@Component({
    standalone: false,
    selector: 'app-password-reset-init',
    templateUrl: './password-reset-init.component.html',
    styleUrls: ['./password-reset-init.scss'],
})
export class PasswordResetInitComponent implements OnInit, AfterViewInit {
  error: string;
  errorEmailNotExists: string;
  resetAccount: any;
  success: string;
  sending = false;

  constructor(
    private passwordResetInitService: PasswordResetInitService,
    private elementRef: ElementRef,
    public activeModal: NgbActiveModal,
    private utmToast: UtmToastService
  ) {
  }

  ngOnInit() {
    this.resetAccount = {};
  }

  ngAfterViewInit() {
    const el = this.elementRef.nativeElement.querySelector('#email') as HTMLInputElement;
    if (el) { el.focus(); }
  }

  requestReset() {
    this.error = null;
    this.errorEmailNotExists = null;
    this.sending = true;
    this.passwordResetInitService.save(this.resetAccount.email).subscribe(
      () => {
        this.success = 'OK';
        this.sending = false;
        this.activeModal.close();
        this.utmToast.showSuccessBottom('An email has been sent with the link to reset your password');
      },
      response => {
        this.success = null;
        this.sending = false;
        this.utmToast.showError('Error', response.error.detail.split(':')[1]);
      });
  }
}
