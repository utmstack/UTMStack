import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {LoginService} from '../../../../../../../core/login/login.service';
import {PasswordComponent} from './password/password.component';
import {SettingsComponent} from './settings/settings.component';

@Component({
  selector: 'app-utm-notification-user-setting',
  templateUrl: './utm-notification-user-setting.component.html',
  styleUrls: ['./utm-notification-user-setting.component.css']
})
export class UtmNotificationUserSettingComponent implements OnInit {

  constructor(public router: Router,
              private modalService: NgbModal,
              private loginService: LoginService) {
  }

  ngOnInit() {
  }

  signOut() {
    this.router.navigate(['/']);
    this.loginService.logout();
  }

  viewSettings() {
    this.modalService.open(SettingsComponent, {centered: true});
  }

  changePassword() {
    this.modalService.open(PasswordComponent, {centered: true});
  }
}
