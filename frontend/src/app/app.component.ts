import {Component, HostListener, OnInit, Renderer2} from '@angular/core';
import {ActivatedRoute, NavigationEnd, Router} from '@angular/router';
import {TranslateService} from '@ngx-translate/core';
import {NgxSpinnerService} from 'ngx-spinner';
import {UtmToastService} from './shared/alert/utm-toast.service';
import {DashboardBehavior} from './shared/behaviors/dashboard.behavior';
import {MenuBehavior} from './shared/behaviors/menu.behavior';
import {ThemeChangeBehavior} from './shared/behaviors/theme-change.behavior';
import {ADMIN_ROLE, USER_ROLE} from './shared/constants/global.constant';
import {AppThemeLocationEnum} from './shared/enums/app-theme-location.enum';
import {UtmAppThemeService} from './shared/services/theme/utm-app-theme.service';
import {retry} from "rxjs/operators";
import {ApiServiceCheckerService} from "./core/auth/api-checker-service";
import {TimezoneFormatService} from "./shared/services/utm-timezone.service";
import {parseQueryParamsToFilter} from "./shared/util/query-params-to-filter.util";
import {LoginService} from "./core/login/login.service";

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
  hideOnline = true;
  iframeView = false;
  favIcon: HTMLLinkElement;

  constructor(
    private spinner: NgxSpinnerService,
    private translate: TranslateService,
    private menuBehavior: MenuBehavior,
    private dashboardBehavior: DashboardBehavior,
    private themeChangeBehavior: ThemeChangeBehavior,
    private utmAppThemeService: UtmAppThemeService,
    private utmToastService: UtmToastService,
    private router: Router, private renderer: Renderer2,
    private apiServiceCheckerService: ApiServiceCheckerService,
    private timezoneFormatService: TimezoneFormatService,
    private activatedRoute: ActivatedRoute,
    private loginService: LoginService) {
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
    });
    this.apiServiceCheckerService.checkApiAvailability();
  }

  ngOnInit(): void {
    this.favIcon = document.querySelector('#appFavicon');
    this.apiServiceCheckerService.isOnlineApi$.subscribe(result => {
      if (result) {
        this.timezoneFormatService.loadTimezoneAndFormat();
        this.getReportLogo();
        this.offline = false;
        if (this.router.url === '/') {
            this.hideOnline = false;
            this.utmToastService.showSuccess('Connection to the UTMStack API was successful.');
          }
        setTimeout(() => {
          this.hideOnline = true;
        }, 3000);
      } else if (result != null && !result && !this.offline) {
        this.offline = true;
        this.utmToastService.showError('Error trying to connect to API', 'An error occurred while trying to connect to the API, ' +
          'please check the UTMStack API connection and try again.');
      }
    });
    this.router.events.subscribe(evt => {
      if (evt instanceof NavigationEnd && evt.url.endsWith('dashboard')) {
      }
    });

    this.themeChangeBehavior.$themeChange.subscribe(value => {
      if (value) {
        this.getReportLogo();
      }
    });
    /**
     * Sync fields of index patterns every 5 min
     */
  }

  @HostListener('window', ['$event'])
  onResize(event) {
    this.height = (window.innerHeight - 85).toString();
  }

  getReportLogo() {
    this.utmAppThemeService.getTheme({page: 0, size: 100})
      .pipe(retry(5))
      .subscribe(response => {
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
    }, error => {
      this.offline = true;
      this.utmToastService.showError('Error trying to connect to API', 'An error occurred while trying to connect to the API, ' +
        'please check the UTMStack API connection and try again.');
    });
  }

  isInExportRoute() {
    return this.router.url.includes('dashboard/export/') || this.router.url.includes('dashboard/export-compliance') ||
      this.router.url.includes('/getting-started') ||
      this.router.url.includes('/dashboard/export-report/') || this.iframeView || this.router.url.includes('/data/alert/detail/');
  }
}
