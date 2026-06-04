import {AfterViewInit, Component, ElementRef, OnInit, Renderer} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../alert/utm-toast.service';

import {PasswordResetInitService} from './password-reset-init.service';

@Component({
  selector: 'app-password-reset-init',
  templateUrl: './password-reset-init.component.html',
  styleUrls: ['./password-reset-init.scss']
})
export class PasswordResetInitComponent implements OnInit, AfterViewInit {
  error: string;
  errorEmailNotExists: string;
  resetAccount: any;
  success: string;
  sending = false;

  constructor(
    private passwordResetInitService: PasswordResetInitService,
    private elementRef: ElementRef, private renderer: Renderer,
    public activeModal: NgbActiveModal,
    private utmToast: UtmToastService
  ) {
  }

  ngOnInit() {
    this.resetAccount = {};
  }

  ngAfterViewInit() {
    this.renderer.invokeElementMethod(this.elementRef.nativeElement.querySelector('#email'), 'focus', []);
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
