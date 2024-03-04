import {Component, HostListener, OnInit, Renderer2} from '@angular/core';
import {Router} from '@angular/router';
import {TranslateService} from '@ngx-translate/core';
import {filter} from 'rxjs/operators';
import {ApiServiceCheckerService} from './core/auth/api-checker-service';
import {MenuBehavior} from './shared/behaviors/menu.behavior';
import {ThemeChangeBehavior} from './shared/behaviors/theme-change.behavior';
import {ADMIN_ROLE, USER_ROLE} from './shared/constants/global.constant';
import {AppThemeLocationEnum} from './shared/enums/app-theme-location.enum';
import {UtmAppThemeService} from './shared/services/theme/utm-app-theme.service';
import {TimezoneFormatService} from './shared/services/utm-timezone.service';

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
  offline = null;
  iframeView = false;
  favIcon: HTMLLinkElement;

  constructor(
    private translate: TranslateService,
    private menuBehavior: MenuBehavior,
    private themeChangeBehavior: ThemeChangeBehavior,
    private utmAppThemeService: UtmAppThemeService,
    private router: Router, private renderer: Renderer2,
    private apiServiceCheckerService: ApiServiceCheckerService,
    private timezoneFormatService: TimezoneFormatService) {

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
  }

  ngOnInit(): void {
    this.favIcon = document.querySelector('#appFavicon');
    this.init();

    this.themeChangeBehavior.$themeChange.subscribe(value => {
      if (value) {
        this.getReportLogo();
      }
    });

    this.apiServiceCheckerService.isOnlineApi$
      .pipe(
        filter(isOnline => isOnline))
      .subscribe(isOnline => {
        if (this.offline) {
          this.init();
        }
        setTimeout(() => this.offline = null, 3000);
      });
  }

  @HostListener('window', ['$event'])
  onResize(event) {
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
    return this.router.url.includes('dashboard/export/') || this.router.url.includes('dashboard/export-compliance') ||
      this.router.url.includes('/getting-started') ||
      this.router.url.includes('/dashboard/export-report/') || this.iframeView || this.router.url.includes('/data/alert/detail/');
  }

   init() {
    this.timezoneFormatService.loadTimezoneAndFormat();
    this.getReportLogo();
  }
}
