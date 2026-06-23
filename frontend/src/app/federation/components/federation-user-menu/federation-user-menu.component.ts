import {Component} from '@angular/core';
import {Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {LoginService} from '../../../core/login/login.service';
import {
  PasswordComponent
} from '../../../shared/components/layout/header/shared/notification/utm-notification-user-setting/password/password.component';
import {
  SettingsComponent
} from '../../../shared/components/layout/header/shared/notification/utm-notification-user-setting/settings/settings.component';

@Component({
  selector: 'app-federation-user-menu',
  templateUrl: './federation-user-menu.component.html',
  styleUrls: ['./federation-user-menu.component.scss']
})
export class FederationUserMenuComponent {
  constructor(private router: Router,
              private modalService: NgbModal,
              private loginService: LoginService) {}

  goToTeamMembers(): void {
    this.router.navigate(['/federation/team-members']);
  }

  openProfileSettings(): void {
    this.modalService.open(SettingsComponent, {centered: true});
  }

  openChangePassword(): void {
    this.modalService.open(PasswordComponent, {centered: true});
  }

  signOut(): void {
    this.router.navigate(['/']);
    this.loginService.logout();
  }
}
