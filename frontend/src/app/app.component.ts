import {Component, HostListener, OnInit, Renderer2} from '@angular/core';
import {Router} from '@angular/router';
import {TranslateService} from '@ngx-translate/core';
import {catchError, delay, distinctUntilChanged, filter, map, takeUntil, tap} from 'rxjs/operators';
import {AccountService} from './core/auth/account.service';
import {ApiServiceCheckerService} from './core/auth/api-checker-service';
import {MenuBehavior} from './shared/behaviors/menu.behavior';
import {ThemeChangeBehavior} from './shared/behaviors/theme-change.behavior';
import {ADMIN_ROLE, USER_ROLE} from './shared/constants/global.constant';
import {AppThemeLocationEnum} from './shared/enums/app-theme-location.enum';
import {UtmAppThemeService} from './shared/services/theme/utm-app-theme.service';
import {AppVersionInfo} from "./shared/types/updates/updates.type";
import {VersionType, VersionInfoService} from "./shared/services/version/version-info.service";
import {EMPTY} from "rxjs";
import {AppVersionService} from './shared/services/version/app-version.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit {
  title = 'utm-stack';
  roles = [ADMIN_ROLE, USER_ROLE];
  menu = false;
  private height: string;
  offline = false;
  online = false;
  iframeView = false;
  favIcon: HTMLLinkElement;
  appLoading: HTMLElement;
  hideStatus = false;
  isAuth = false;
  viewportHeight: number;

  constructor(
    private translate: TranslateService,
    private menuBehavior: MenuBehavior,
    private themeChangeBehavior: ThemeChangeBehavior,
    private utmAppThemeService: UtmAppThemeService,
    private router: Router, private renderer: Renderer2,
    private apiServiceCheckerService: ApiServiceCheckerService,
    private accountService: AccountService,
    private checkForUpdatesService: AppVersionService,
    private versionTypeService: VersionInfoService) {

    this.translate.setDefaultLang('en');

    this.menuBehavior.$menu.subscribe(men => {
      this.menu = men;
    });

    window.addEventListener('beforeprint', (event) => {
      this.menu = false;
      this.menuBehavior.$menu.next(false);
    });

    this.router.events.subscribe((event) => {
      if (router.url.toString().includes('/dashboard/view/')) {
        this.renderer.addClass(document.body, 'scroll-0');
      } else {
        this.renderer.removeClass(document.body, 'scroll-0');
      }
      if (this.router.url.includes('iframe')) {
        this.iframeView = true;
      }
      if (this.router.url.includes('url')) {
        this.hideStatus = true;
      }
    });
  }

  ngOnInit(): void {
    this.favIcon = document.querySelector('#appFavicon');
    this.appLoading = document.getElementById('app-loading');
    this.viewportHeight = window.innerHeight;
    this.init();

    this.themeChangeBehavior.$themeChange.subscribe(value => {
      if (value) {
        this.getReportLogo();
      }
    });

    this.accountService.getAuthenticationState()
      .pipe(distinctUntilChanged((prev, next) =>  prev && next && prev.id === next.id))
      .subscribe(identity => {
        this.isAuth = !!identity;
        if (this.isAuth) {
          const appBg = document.getElementById('app-background');
          appBg.style.display = 'none';
        }
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
      .subscribe();

    this.appLoading.style.display = 'none';
  }

  @HostListener('window', ['$event'])
  onResize(event) {
    this.viewportHeight = window.innerHeight;
    this.height = (window.innerHeight - 85).toString();
  }

  getReportLogo() {
    this.utmAppThemeService.getTheme({page: 0, size: 100})
      .subscribe(response => {
        this.offline = false;
        for (const img of response.body) {
        switch (img.shortName) {
          case AppThemeLocationEnum.LOGIN:
            this.favIcon.href = img.userImg;
            this.themeChangeBehavior.$themeIcon.next(img.userImg);
            break;
          case AppThemeLocationEnum.HEADER:
            this.themeChangeBehavior.$themeNavbarIcon.next(img.userImg);
            break;
          case AppThemeLocationEnum.REPORT:
            this.themeChangeBehavior.$themeReportIcon.next(img.userImg);
            break;
          case AppThemeLocationEnum.REPORT_COVER:
            this.themeChangeBehavior.$themeReportCover.next(img.userImg);
            break;
        }
      }
        this.apiServiceCheckerService.setOnlineStatus(true);
    }, error => {
        this.offline = true;
        this.apiServiceCheckerService.checkApiAvailability();
      });
  }

  isInExportRoute() {
    return this.router.url.includes('dashboard/export/') ||
           this.router.url.includes('dashboard/export-compliance') ||
           this.router.url.includes('compliance/print-view') ||
      this.router.url.includes('/getting-started') ||
      this.router.url.includes('/dashboard/export-report/') || this.iframeView || this.router.url.includes('/data/alert/detail/');
  }

   init() {
    this.getReportLogo();
  }

  getContentHeight() {
    return this.isAuth ? 100 - ((120 / this.viewportHeight) * 100) + 'vh' : '100vh';
  }
}
