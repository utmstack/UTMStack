import {AfterViewInit, Component, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {DomSanitizer} from '@angular/platform-browser';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable} from 'rxjs';
import {AccountService} from '../../../../core/auth/account.service';
import {ApiServiceCheckerService} from '../../../../core/auth/api-checker-service';
import {LoginService} from '../../../../core/login/login.service';
import {UtmToastService} from '../../../alert/utm-toast.service';
import {MenuBehavior} from '../../../behaviors/menu.behavior';
import {ThemeChangeBehavior} from '../../../behaviors/theme-change.behavior';
import {ADMIN_DEFAULT_EMAIL, ADMIN_ROLE, DEMO_URL, USER_ROLE} from '../../../constants/global.constant';
import {stringParamToQueryParams} from '../../../util/query-params-to-filter.util';
import {PasswordResetInitComponent} from '../password-reset/init/password-reset-init.component';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit {
  authenticationError: boolean;
  password: string;
  rememberMe: boolean;
  username: string;
  credentials: any;
  formLogin: FormGroup;
  logged = false;
  roles = [ADMIN_ROLE, USER_ROLE];
  startLogin = false;
  isInDemo: boolean;
  loadingAuth = true;
  loginImage$: Observable<string>;

  constructor(
    private loginService: LoginService,
    private router: Router,
    private activatedRoute: ActivatedRoute,
    private fb: FormBuilder,
    private accountService: AccountService,
    private utmToast: UtmToastService,
    private menuBehavior: MenuBehavior,
    public sanitizer: DomSanitizer,
    private modalService: NgbModal,
    private themeChangeBehavior: ThemeChangeBehavior,
    private spinner: NgxSpinnerService,
    private apiServiceCheckerService: ApiServiceCheckerService
  ) {
    this.credentials = {};
    this.isInDemo = window.location.href.includes(DEMO_URL);
    this.loginImage$ = this.themeChangeBehavior.$themeIcon.asObservable();
  }

  ngOnInit() {
    this.apiServiceCheckerService.isOnlineApi$.subscribe(result => {
      if (result) {
        this.activatedRoute.queryParams.subscribe(params => {
          if (params.token) {
            this.loginService.loginWithToken(params.token, true).then(() => {
              if (params.url) {
                this.checkLogin(params.url);
              } else {
                this.router.navigate(['/dashboard/overview']).then(() => {
                  this.spinner.hide('loadingSpinner');
                });
              }
            });
          } else if (params.key) {
            this.loginService.loginWithKey(params.key, true).then(() => {
              this.startInternalNavigation();
            });
          }
        });
        this.initForm();
        this.loadingAuth = false;
      }
    });
  }

  checkLogin(url ?: string) {
    console.log('Checking URL token');
    this.accountService.identity(true).then(value => {
      if (value) {
        this.spinner.show('loadingSpinner');
        if (url) {
          const urlRoute = url.split('<-PARAMS->');
          const route = urlRoute[0];
          const params = urlRoute[1];
          if (params) {
            stringParamToQueryParams(params).then(queryParams => {
              this.router.navigate([route],
                {queryParams}).then(() => {
                this.menuBehavior.$menu.next(false);
                this.spinner.hide('loadingSpinner');
              });
            });
          } else {
            this.router.navigate([route]).then(() => {
              this.spinner.hide('loadingSpinner');
            });
          }
        }
      } else {
        this.spinner.hide('loadingSpinner');
        this.loadingAuth = false;
      }
    });
  }

  initForm() {
    this.formLogin = this.fb.group({
      username: ['', [Validators.required]],
      password: ['', [Validators.required]],
      rememberMe: [true]
    });
  }

  cancel() {
    this.authenticationError = false;
  }

  login() {
    this.startLogin = true;
    this.loginService
      .login(this.formLogin.value)
      .then((auth) => {
        if (auth) {
          this.authenticationError = false;
          this.logged = true;
          this.startLogin = false;
          this.spinner.show();
          this.startNavigation();
        } else {
          this.spinner.show();
          this.router.navigate(['/totp'])
            .then(() => this.spinner.hide());
        }
      })
      .catch((err) => {
        this.startLogin = false;
        this.authenticationError = true;
        const utmStackError = err.headers.get('X-UtmStack-error');
        if (utmStackError.includes('UserJWTController.authorize: blocked')) {
          this.utmToast.showError('Login blocked', 'Your ip was blocked due multiple login failures, please try again in 10 minutes');
        } else {
          this.utmToast.showError('Login fail', 'Authentication error, ' +
            'check your data and try again. More than 9 failure attempts will block your IP for 10 minutes');
        }
      });
  }

  resetPassword() {
    this.modalService.open(PasswordResetInitComponent, {centered: true, size: 'sm'});
  }

  keyDownFunction($event: KeyboardEvent) {
    if (this.formLogin.valid && $event.code === 'Enter') {
      this.login();
    }
  }

  startNavigation() {
    this.accountService.identity(true).then(account => {
      if (account) {
        const redirectTo = (account.authorities.includes(ADMIN_ROLE) && account.email === ADMIN_DEFAULT_EMAIL)
          ? '/getting-started' : '/dashboard/overview';
        this.router.navigate([redirectTo])
          .then(() => this.spinner.hide());
      } else {
        this.logged = false;
        this.utmToast.showError('Login error', 'User without privileges.');
      }
    });
  }

  startInternalNavigation() {
    this.router.navigate(['/dashboard/overview']);
  }

}
