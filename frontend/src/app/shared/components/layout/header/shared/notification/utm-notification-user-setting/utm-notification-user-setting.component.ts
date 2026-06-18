import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {Observable} from 'rxjs';
import {LoginService} from '../../../../../../../core/login/login.service';
import {
  TeamManagementModalComponent
} from '../../../../../../../federation/components/team-management-modal/team-management-modal.component';
import {FederationModeService} from '../../../../../../../federation/services/federation-mode.service';
import {PasswordComponent} from './password/password.component';
import {SettingsComponent} from './settings/settings.component';

@Component({
  selector: 'app-utm-notification-user-setting',
  templateUrl: './utm-notification-user-setting.component.html',
  styleUrls: ['./utm-notification-user-setting.component.css']
})
export class UtmNotificationUserSettingComponent implements OnInit {
  federationActive$: Observable<boolean>;

  constructor(public router: Router,
              private modalService: NgbModal,
              private loginService: LoginService,
              private federationModeService: FederationModeService) {
    this.federationActive$ = this.federationModeService.active$;
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

  openTeamManagement() {
    this.modalService.open(TeamManagementModalComponent, {
      centered: true,
      size: 'lg',
      backdrop: 'static'
    });
  }
}
