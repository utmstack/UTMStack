import {Component, OnDestroy, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {NavigationEnd, Router} from '@angular/router';
import {Observable, Subject} from 'rxjs';
import {filter, map, startWith, takeUntil} from 'rxjs/operators';
import {AccountService} from '../../../../core/auth/account.service';
import {User} from '../../../../core/user/user.model';
import {FederationModeService} from '../../../../federation/services/federation-mode.service';
import {ThemeChangeBehavior} from '../../../behaviors/theme-change.behavior';
import {ADMIN_ROLE} from '../../../constants/global.constant';
import {AppThemeLocationEnum} from '../../../enums/app-theme-location.enum';
import {VersionInfoService} from '../../../services/version/version-info.service';
import {AppVersionInfo} from '../../../types/updates/updates.type';

const FEDERATION_WELCOME_ROUTE = '/federation/welcome';

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
  federationActive$: Observable<boolean>;
  federationWelcomeRoute$: Observable<boolean>;
  destroy$: Subject<void> = new Subject();

  constructor(private accountService: AccountService,
              public sanitizer: DomSanitizer,
              private router: Router,
              private themeChangeBehavior: ThemeChangeBehavior,
              private versionTypeService: VersionInfoService,
              private federationModeService: FederationModeService) {
    this.federationActive$ = this.federationModeService.active$;

    this.federationWelcomeRoute$ = this.router.events.pipe(
      filter(event => event instanceof NavigationEnd),
      map((event: NavigationEnd) => this.isFederationWelcome(event.urlAfterRedirects)),
      startWith(this.isFederationWelcome(this.router.url))
    );
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

    this.versionTypeService.appVersionInfo$
      .pipe(takeUntil(this.destroy$))
      .subscribe((versionInfo: AppVersionInfo) => this.versionInfo = versionInfo);
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private isFederationWelcome(url: string): boolean {
    return url.startsWith(FEDERATION_WELCOME_ROUTE);
  }
}
