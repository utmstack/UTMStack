import {Component, OnDestroy, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {EMPTY, Subject} from 'rxjs';
import {catchError, filter, map, takeUntil, tap} from 'rxjs/operators';
import {AccountService} from '../../../../core/auth/account.service';
import {User} from '../../../../core/user/user.model';
import {ThemeChangeBehavior} from '../../../behaviors/theme-change.behavior';
import {ADMIN_ROLE} from '../../../constants/global.constant';
import {AppThemeLocationEnum} from '../../../enums/app-theme-location.enum';
import {VersionType, VersionTypeService} from "../../../services/util/version-type.service";
import {AppVersionInfo} from "../../../types/updates/updates.type";
import {CheckForUpdatesService} from "../../../services/updates/check-for-updates.service";

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
  versionInfo: AppVersionInfo;
  destroy$: Subject<void> = new Subject();

  constructor(private accountService: AccountService,
              public sanitizer: DomSanitizer,
              private themeChangeBehavior: ThemeChangeBehavior,
              private versionTypeService: VersionTypeService,
              private checkForUpdatesService: CheckForUpdatesService) {
  }

  ngOnInit() {
    this.themeChangeBehavior.$themeNavbarIcon
      .pipe(
        takeUntil(this.destroy$),
        filter(icon => !!icon))
          .subscribe(icon => this.logoImage = icon);

    this.accountService.identity().then(account => {
      this.user = account;
    });

    this.checkForUpdatesService.getVersion()
      .pipe(
        map(response => response.body || null),
        tap((versionInfo: AppVersionInfo) => {
          const version = versionInfo && versionInfo.version || '';
          const versionType = versionInfo.edition.includes('community') || version === ''
            ? VersionType.COMMUNITY
            : VersionType.ENTERPRISE;

          if (versionType !== this.versionTypeService.versionType()) {
            this.versionTypeService.changeVersionType(versionType);
          }
        }),
        catchError(() => {
          return EMPTY;
        })
      )
      .subscribe(versionInfo => {
        this.versionInfo = versionInfo;
      });
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

}
