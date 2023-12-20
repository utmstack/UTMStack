import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {AuthServerProvider} from '../../../../core/auth/auth-jwt.service';
import {UtmToastService} from '../../../alert/utm-toast.service';


@Component({
  selector: 'app-totp',
  templateUrl: './totp.component.html',
  styleUrls: ['./totp.component.scss']
})
export class TotpComponent implements OnInit {
  form: any = {};
  isLoggedIn = false;
  isLoginFailed = false;
  errorMessage = '';
  currentUser: any;
  verifying = false;

  constructor(private authService: AuthServerProvider,
              private router: Router,
              private spinner: NgxSpinnerService,
              private utmToast: UtmToastService) {
  }

  ngOnInit(): void {
  }


  onSubmit() {
    this.verifying = true;
    this.authService
      .verifyCode(this.form.code).subscribe((auth) => {
      if (auth) {
        this.verifying = false;
        this.spinner.show();
        this.router.navigate(['/dashboard/overview'])
          .then(() => this.spinner.hide());
      }
    }, error => {
      this.utmToast.showError('Code validation', 'Verification code is invalid or has expired');
      this.verifying = false;
    });
  }

  navigateToLogin() {
    this.router.navigate(['/']);
  }
}
