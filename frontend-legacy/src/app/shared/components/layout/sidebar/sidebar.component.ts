import {Component, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {AccountService} from '../../../../core/auth/account.service';
import {LoginService} from '../../../../core/login/login.service';
import {User} from '../../../../core/user/user.model';
import {MenuBehavior} from '../../../behaviors/menu.behavior';
import {ADMIN_ROLE} from '../../../constants/global.constant';
import {SYSTEM_MENU_ICONS_PATH} from '../../../constants/menu_icons.constants';
import {UtmRunModeService} from '../../../services/active-modules/utm-run-mode.service';

@Component({
  selector: 'app-sidebar',
  templateUrl: './sidebar.component.html',
  styleUrls: ['./sidebar.component.scss']
})
export class SidebarComponent implements OnInit {
  roleAdmin = ADMIN_ROLE;
  user: User;
  private menu: boolean;
  iconPath = SYSTEM_MENU_ICONS_PATH;

  constructor(public router: Router,
              private spinner: NgxSpinnerService,
              private loginService: LoginService,
              private accountService: AccountService,
              private menuBehavior: MenuBehavior,
              private utmRunModeService: UtmRunModeService) {
    this.menuBehavior.$menu.subscribe(menu => {
      this.menu = menu;
    });
  }

  ngOnInit() {
    this.accountService.identity().then(account => {
      this.user = account;
    });
  }

  navigateTo(link: string) {
    this.spinner.show('loadingSpinner');
    this.router.navigate([link]).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  signOut() {
    this.router.navigate(['/']);
    this.loginService.logout();
  }

  closeMenu() {
    this.menuBehavior.$menu.next(false);
  }

  isActive(link: string) {
    return this.router.url === link;
  }

}
