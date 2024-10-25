import {Component, OnDestroy, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {Observable, Subject} from 'rxjs';
import {MenuBehavior} from '../../../../behaviors/menu.behavior';
import {NavBehavior} from '../../../../behaviors/nav.behavior';
import {SYSTEM_MENU_ICONS_PATH} from '../../../../constants/menu_icons.constants';
import {MenuService} from '../../../../services/menu/menu.service';
import {Menu} from '../../../../types/menu/menu.model';
import {stringParamToQueryParams} from '../../../../util/query-params-to-filter.util';
import {filter, takeUntil, tap} from 'rxjs/operators';
import {OpenAlertsService} from '../../../../../data-management/alert-management/shared/services/open-alerts.service';
import {UtmToastService} from '../../../../alert/utm-toast.service';

@Component({
  selector: 'app-header-menu-navigation',
  templateUrl: './header-menu-navigation.component.html',
  styleUrls: ['./header-menu-navigation.component.scss']
})
export class HeaderMenuNavigationComponent implements OnInit, OnDestroy {
  menus: Menu[] = [];
  defaultStructureMenu: Menu[] = [];
  searching = false;
  iconPath = SYSTEM_MENU_ICONS_PATH;
  active: number;
  destroy$: Subject<void> =  new Subject<void>();
  prevTotal$: Observable<number>;

  constructor(private navBehavior: NavBehavior,
              private spinner: NgxSpinnerService,
              public router: Router,
              private menuBehavior: MenuBehavior,
              private openAlertsService: OpenAlertsService,
              private menuService: MenuService,
              private toast: UtmToastService) {
  }

  ngOnInit() {

    this.loadMenus();

    this.navBehavior.$nav
        .pipe(takeUntil(this.destroy$))
        .subscribe( value => {
      if (value) {
        this.loadMenus();
      }
    });

    this.prevTotal$ = this.openAlertsService.openAlerts$
      .pipe(
        takeUntil(this.destroy$),
        filter(openAlerts => openAlerts > 0),
        tap((openAlerts) => {
          if (!this.router.url.includes('/dashboard/export')) {
            this.toast.showWarning('There are ' + openAlerts +
              ' pending alerts to manage', 'New alerts');
          }
        }));
  }

  /**
   * Determine if show AD or not based on service
   * @param menu MenuNavType
   */
  showAdMenu(menu: Menu) {
    return !menu.url.includes('active-directory') && menu.menuActive;
  }

  showAdParentMenu(menu: Menu) {
    return menu.menuActive;
  }

  loadMenus() {
    this.menuService.getMenuStructure(true).subscribe(reponse => {
      this.menus = reponse.body;
      this.defaultStructureMenu = reponse.body;
    });
  }

  showDashboard(menu: Menu) {
    this.spinner.show('loadingSpinner');
    // if (menu.type !== 2) {
    const url = menu.url.split('?');
    const route = url[0];
    const params = url[1];
    if (params) {
      stringParamToQueryParams(params).then(queryParams => {
        // Reload component on navigate
        this.router.routeReuseStrategy.shouldReuseRoute = () => false;
        this.router.onSameUrlNavigation = 'reload';
        // this.httpCancelService.cancelPendingRequests();
        this.router.navigate([route],
          {queryParams}).then(() => {
          this.menuBehavior.$menu.next(false);
          this.spinner.hide('loadingSpinner');
        });
      });
    } else {
      // this.httpCancelService.cancelPendingRequests();
      this.router.navigate([menu.url]).then(() => {
        this.spinner.hide('loadingSpinner');
      });
    }
  }

  onSearchMenu($event: string, menu: Menu) {
    const indexOfMenu = this.menus.findIndex(value => value.id === menu.id);
    if ($event || $event !== '') {
      this.searching = true;
      menu.childrens = this.menus[indexOfMenu].childrens
        .filter(value => value.name.toLowerCase().includes($event.toLowerCase()));
    } else {
      this.searching = false;
      this.menus = this.defaultStructureMenu;
    }
  }

  isActive(menu: Menu) {
    return menu.url ? (menu.url.replace(/\//g, '')
      === this.router.url.replace(/\//g, '')) : false;
  }

  isActiveChildren(childrens: Menu[], actions: Menu[], activeParent: number) {
    actions = actions ? actions : [];
    childrens = childrens ? childrens : [];
    const isChild = childrens.findIndex(value => value.parentId === activeParent) > -1
      || actions.findIndex(value => value.parentId === activeParent) > -1;
    if (isChild) {
      const indexChild = childrens.findIndex((value) => {
        return this.router.url.replace(/\//g, '').includes(value.url.replace(/\//g, '')) && value.parentId === activeParent;
      });
      const indexActions = actions.findIndex((value) => {
        return value.url.replace(/\//g, '')
          === this.router.url.replace(/\//g, '') && value.parentId === activeParent;
      });
      return indexChild > -1 || indexActions > -1;
    } else {
      return false;
    }
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
    this.openAlertsService.stopInterval();
  }
}
