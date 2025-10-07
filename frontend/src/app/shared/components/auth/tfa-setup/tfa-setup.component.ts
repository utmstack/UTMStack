import { Component, OnDestroy, OnInit } from '@angular/core';
import {DomSanitizer, SafeUrl} from '@angular/platform-browser';
import {Router} from '@angular/router';
import { Observable, Subscription } from 'rxjs';
import {UtmToastService} from '../../../alert/utm-toast.service';
import {ThemeChangeBehavior} from '../../../behaviors/theme-change.behavior';
import {TfaMethod, TfaService, TfaStage} from '../../../services/tfa/tfa.service';
import {AuthServerProvider} from "../../../../core/auth/auth-jwt.service";
import {AccountService} from "../../../../core/auth/account.service";

@Component({
  selector: 'app-tfa-setup',
  templateUrl: './tfa-setup.component.html',
  styleUrls: ['./tfa-setup.component.scss']
})
export class TfaSetupComponent implements OnInit, OnDestroy {

  TfaMethod = TfaMethod;

  step: 'method-selection' | 'setup' | 'verification' | 'success' = 'method-selection';
  selectedMethod: TfaMethod | null = null;

  qrCodeUrl: SafeUrl = '';

  code  = '';
  verifying = false;
  codeVerified = false;
  errorMessage = '';

  expiresInSeconds = 300;
  private timerSubscription: Subscription | null = null;

  resending = false;

  loginImage$: Observable<string> | undefined;

  constructor( private themeChangeBehavior: ThemeChangeBehavior,
               public sanitizer: DomSanitizer,
               private router: Router,
               private tfaService: TfaService,
               private utmToastService: UtmToastService,
               private accountService: AccountService,
               private authServerProvider: AuthServerProvider
  ) {}

  ngOnInit(): void {
    this.loginImage$ = this.themeChangeBehavior.$themeIcon.asObservable();
  }

  selectMethod(method: TfaMethod): void {
    this.selectedMethod = method;
    this.step = 'setup';
    this.clearError();
    this.initTfa();
  }

  private initTfa(): void {
    this.tfaService.enrollTfa({
      method: this.selectedMethod,
      stage: TfaStage.INIT
    }).subscribe((response) => {
      if (this.selectedMethod === TfaMethod.TOTP) {
        this.qrCodeUrl = response.delivery.target ?
          this.sanitizer.bypassSecurityTrustUrl(`data:image/png;base64,${response.delivery.target}`) : null;
      } else {
        this.qrCodeUrl = '';
      }
    }, (error) => {
      if (error.status === 400) {
        this.utmToastService.showError('Error initializing TFA',
          'An error occurred while initializing two-factor authentication, please check configuration and try again');
      } else {
        this.utmToastService.showError('Error initializing TFA',
          'An error occurred while initializing two-factor authentication, please contact with the support team');
      }
    });
  }

  private sendEmailCode(): void {
    /*this.resending = true;

    // En producción, llamada al backend
    setTimeout(() => {
      this.resending = false;
      console.log('Código enviado por email');
      // this.tfaService.sendEmailCode().subscribe(() => {
      //   this.resending = false;
      // });
    }, 1000);*/
  }


  resendCode(): void {
    if (this.selectedMethod === TfaMethod.EMAIL) {
      this.sendEmailCode();
      this.clearError();
    }
  }


  onExpire(): void {

  }

  onSubmit() {
    this.verifying = true;
    this.tfaService.enrollTfa({
      method: this.selectedMethod,
      stage: TfaStage.VERIFY,
      code: this.code
    }).subscribe(response => {
      this.verifying = false;
      this.errorMessage = !response.valid ? 'Verification code is invalid' : response.expired ? 'Verification code has expired' : '';
      if (response.valid) {
        this.codeVerified = true;
        this.step = 'success';
      }
    }, error => {
      this.verifying = false;
      this.errorMessage = 'An error occurred while verifying the code, please try again later';
    });
  }

  clearError(): void {
    if(this.code.length == 6 && this.selectedMethod === TfaMethod.TOTP ){
      this.onSubmit()
    }
    this.errorMessage = '';
  }

  backToMethodSelection(): void {
    this.step = 'method-selection';
    this.selectedMethod = null;
    this.clearError();
    this.code = '';
    if (this.timerSubscription) {
      this.timerSubscription.unsubscribe();
    }
  }

  completeSetup(): void {
    this.verifying = true;
    this.tfaService.enrollTfa({
      stage: TfaStage.COMPLETE,
      method: this.selectedMethod,
      enable: true
    }).subscribe(response => {

      /*this.authServerProvider.tfaMethod = this.selectedMethod;
      this.router.navigate(['/totp']);*/
      this.authServerProvider.storeAuthenticationToken(response.token);
      this.accountService.startNavigation();
    }, error => {
      this.verifying = false;
      this.verifying = false;
      this.errorMessage = 'An error occurred while saving the configuration, please try again later';
    });
  }

  /**
   * Omite la configuración TFA
   */
  skipSetup(): void {
    this.router.navigate(['/']);
  }

  ngOnDestroy(): void {
    if (this.timerSubscription) {
      this.timerSubscription.unsubscribe();
    }
  }
}
