import {Component, OnDestroy, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {AccountService} from '../../../../core/auth/account.service';
import {User} from '../../../../core/user/user.model';
import {ThemeChangeBehavior} from '../../../behaviors/theme-change.behavior';
import {ADMIN_ROLE} from '../../../constants/global.constant';
import {AppThemeLocationEnum} from '../../../enums/app-theme-location.enum';
import {CheckForUpdatesService} from '../../../services/updates/check-for-updates.service';

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
  currentVersion: any;

  constructor(private accountService: AccountService,
              public sanitizer: DomSanitizer,
              private themeChangeBehavior: ThemeChangeBehavior,
              private checkForUpdatesService: CheckForUpdatesService) {
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

    this.getVersionInfo();
  }

  getVersionInfo() {
    this.checkForUpdatesService.getVersion().subscribe(response => {
      this.currentVersion = response.body;
    });
  }

  ngOnDestroy() {
  }

}
