import {Component, OnInit} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {AccountService} from '../../../../core/auth/account.service';
import {LoginService} from '../../../../core/login/login.service';
import {UtmToastService} from '../../../alert/utm-toast.service';
import {MenuBehavior} from '../../../behaviors/menu.behavior';
import {ALERT_DETAIL_ROUTE} from '../../../constants/app-routes.constant';
import {PasswordResetInitComponent} from '../password-reset/init/password-reset-init.component';

@Component({
  selector: 'app-confirm-identity',
  templateUrl: './confirm-identity.component.html',
  styleUrls: ['./confirm-identity.component.scss']
})
export class ConfirmIdentityComponent implements OnInit {
  authenticationError: boolean;
  password: string;
  rememberMe: boolean;
  username: string;
  credentials: any;
  formLogin: FormGroup;
  logged = false;
  startLogin = false;
  params: any;

  constructor(
    private loginService: LoginService,
    private router: Router,
    private fb: FormBuilder,
    private accountService: AccountService,
    private utmToast: UtmToastService,
    private menuBehavior: MenuBehavior,
    private modalService: NgbModal,
    private spinner: NgxSpinnerService,
    private activatedRoute: ActivatedRoute,
  ) {
    this.credentials = {};
    this.activatedRoute.params.subscribe(params => {
      this.params = params;
    });
  }

  ngOnInit() {
    this.accountService.identity(true).then(value => {
      if (value) {
        this.spinner.show();
        this.router.navigate([ALERT_DETAIL_ROUTE + '/' + this.params.id]).then(() => {
          this.spinner.hide();
        });
      }
    });
    this.initForm();
  }

  initForm() {
    this.formLogin = this.fb.group({
      username: ['', [Validators.required]],
      password: ['', [Validators.required]],
    });
  }

  cancel() {
    this.authenticationError = false;
  }

  login() {
    this.startLogin = true;
    this.loginService
      .login(this.formLogin.value)
      .then((account) => {
        this.authenticationError = false;
        this.logged = true;
        this.startLogin = false;
        this.spinner.show();
        this.router.navigate([ALERT_DETAIL_ROUTE + '/' + this.params.id])
          .then(() => this.spinner.hide());
      })
      .catch(() => {
        this.startLogin = false;
        this.authenticationError = true;
        this.utmToast.showError('Confirm identity fail',
          'Authentication error, check your data and try again');
      });
  }

  resetPassword() {
    this.modalService.open(PasswordResetInitComponent, {centered: true, size: 'sm'});
  }
}
