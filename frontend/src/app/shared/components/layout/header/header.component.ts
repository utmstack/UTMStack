import {Component, OnDestroy, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {UtmModulesService} from '../../../../app-module/shared/services/utm-modules.service';
import {AccountService} from '../../../../core/auth/account.service';
import {LoginService} from '../../../../core/login/login.service';
import {User} from '../../../../core/user/user.model';
import {NavBehavior} from '../../../behaviors/nav.behavior';
import {ThemeChangeBehavior} from '../../../behaviors/theme-change.behavior';
import {ADMIN_ROLE} from '../../../constants/global.constant';
import {AppThemeLocationEnum} from '../../../enums/app-theme-location.enum';
import {ActiveAdModuleActiveService} from '../../../services/active-modules/active-ad-module.service';
import {UtmRunModeService} from '../../../services/active-modules/utm-run-mode.service';

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.scss']
})
export class HeaderComponent implements OnInit, OnDestroy {
  user: User;
  menuActive = false;
  roleAdmin = ADMIN_ROLE;
  place = AppThemeLocationEnum;
  logoImage: string;
  altImage: string;

  constructor(private loginService: LoginService,
              private accountService: AccountService,
              private adModuleActiveService: ActiveAdModuleActiveService,
              public sanitizer: DomSanitizer,
              private moduleService: UtmModulesService,
              private navBehavior: NavBehavior,
              private utmRunModeService: UtmRunModeService,
              private themeChangeBehavior: ThemeChangeBehavior) {
  }

  ngOnInit() {
    this.themeChangeBehavior.$themeNavbarIcon.subscribe(icon => {
      if (icon) {
        this.logoImage = icon;
      }
    });
    this.accountService.identity().then(account => {
      this.user = account;
    });
  }

  ngOnDestroy() {
  }

}
