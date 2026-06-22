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
import {SYSTEM_MENU_ICONS_PATH} from '../../../shared/constants/menu_icons.constants';

@Component({
  selector: 'app-federation-sidebar',
  templateUrl: './federation-sidebar.component.html',
  styleUrls: ['./federation-sidebar.component.scss']
})
export class FederationSidebarComponent {
  iconPath = SYSTEM_MENU_ICONS_PATH;

  constructor(private router: Router,
              private modalService: NgbModal,
              private loginService: LoginService) {}

  isActive(link: string): boolean {
    return this.router.url.startsWith(link);
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
