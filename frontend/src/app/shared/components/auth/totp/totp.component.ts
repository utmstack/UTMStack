import {Component, OnDestroy, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {interval, Observable, Subscription} from 'rxjs';
import {AccountService} from '../../../../core/auth/account.service';
import {AuthServerProvider} from '../../../../core/auth/auth-jwt.service';
import {StateStorageService} from '../../../../core/auth/state-storage.service';
import {UtmToastService} from '../../../alert/utm-toast.service';
import {ThemeChangeBehavior} from '../../../behaviors/theme-change.behavior';
import {ADMIN_DEFAULT_EMAIL, ADMIN_ROLE} from '../../../constants/global.constant';
import {TfaMethod} from '../../../services/tfa/tfa.service';
import {extractQueryParamsForNavigation} from '../../../util/query-params-to-filter.util';


@Component({
  selector: 'app-totp',
  templateUrl: './totp.component.html',
  styleUrls: ['./totp.component.scss']
})
export class TotpComponent implements OnInit, OnDestroy {
  form: any = {};
  errorMessage = '';
  isVerifying = false;
  loginImage$: Observable<string>;
  TfaMethod = TfaMethod;
  method: TfaMethod;
  isVerified = false;
  verificationCode = '';
  private expireSub: Subscription;

  constructor(private authService: AuthServerProvider,
              private router: Router,
              private spinner: NgxSpinnerService,
              private themeChangeBehavior: ThemeChangeBehavior,
              public sanitizer: DomSanitizer,
              private stateStorageService: StateStorageService,
              private accountService: AccountService,
              private utmToast: UtmToastService) {
  }

  ngOnInit(): void {
    this.method = this.authService.tfaMethod;
    this.loginImage$ = this.themeChangeBehavior.$themeIcon.asObservable();

    this.expireSub = interval(30 * 1000).subscribe(() => this.onExpire());
  }


  onSubmit() {
    this.isVerifying = true;
    this.authService
      .verifyCode(this.verificationCode).subscribe((auth) => {
      if (auth) {
        this.isVerified = true;
        this.startNavigation();
      }
    }, error => {
      this.errorMessage = error.headers.get('X-UtmStack-error');
      this.isVerifying = false;
    });
  }

  backToLogin() {
    this.router.navigate(['/']);
  }

  onExpire() {
    console.log('expired');
    this.authService.renewCode().subscribe();
  }

  clearError() {
    if (this.verificationCode.length === 6) {
      this.onSubmit();
    }
    this.errorMessage = '';
  }

  startNavigation() {
    this.accountService.identity(true).then(account => {
      if (account) {
        const { path, queryParams } =
          extractQueryParamsForNavigation(this.stateStorageService.getUrl() ? this.stateStorageService.getUrl() : '' );
        if (path) {
          this.stateStorageService.resetPreviousUrl();
        }
        const redirectTo = (account.authorities.includes(ADMIN_ROLE) && account.email === ADMIN_DEFAULT_EMAIL)
          ? '/getting-started' : !!path ? path : '/dashboard/overview';
        console.log(redirectTo);
        this.router.navigate([redirectTo], {queryParams})
          .then(() => this.spinner.hide());
      } else {
        this.utmToast.showError('Login error', 'User without privileges.');
      }
    });
  }

  ngOnDestroy() {
    this.expireSub.unsubscribe();
  }
}
