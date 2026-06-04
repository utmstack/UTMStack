import {Component, Input} from '@angular/core';
import {NgbActiveModal} from "@ng-bootstrap/ng-bootstrap";
import {TfaMethod, TfaService} from '../../services/tfa/tfa.service';

@Component({
  selector: 'app-utm-tfa-verification',
  templateUrl: './utm-tfa-verification.component.html',
  styleUrls: ['./utm-tfa-verification.component.css']
})
export class UtmTfaVerificationComponent {

  @Input() method: TfaMethod;
  @Input() qrCodeUrl?: string;
  @Input() expiresInSeconds = 300;

  code = '';
  verifying = false;
  errorMessage = '';
  readonly TfaMethod = TfaMethod;
  codeVerified = false;

  constructor(private tfaService: TfaService,
              private activeModal: NgbActiveModal) {}

  onSubmit() {
    this.verifying = true;
    this.tfaService.verifyTfa({
      method: this.method,
      code: this.code
    }).subscribe(response => {
      this.verifying = false;
      this.errorMessage = !response.valid ? 'Verification code is invalid' : response.expired ? 'Verification code has expired' : '';
      if (response.valid) {
        this.showSuccessState();
      }
    }, error => {
      this.verifying = false;
      this.errorMessage = 'An error occurred while verifying the code, please try again later';
    });

  }

  showSuccessState() {
    this.codeVerified = true;
  }

  saveConfiguration() {
    this.activeModal.close(true);
  }

  resendCode() {
  }

  onExpire() {
    this.errorMessage = 'Verification code has expired. Please request a new code.';
  }
}
