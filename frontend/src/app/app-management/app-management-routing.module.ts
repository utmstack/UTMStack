import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {SaasRouteAccessService} from '../core/auth/saas-route-access-service';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {AppConfigComponent} from './app-config/app-config.component';
import {AppLogsComponent} from './app-logs/app-logs.component';
import {AppManagementComponent} from './app-management.component';
import {AppThemeComponent} from './app-theme/app-theme.component';
import {AuditsComponent} from './audits/audits.component';
import {ConnectionKeyComponent} from './connection-key/connection-key.component';
import {HealthChecksComponent} from './health-checks/health-checks.component';
import {IndexManagementComponent} from './index-management/index-management.component';
import {IndexPatternListComponent} from './index-pattern/index-pattern-list/index-pattern-list.component';
import {MenuComponent} from './menu/menu.component';
import {RolloverConfigComponent} from './rollover-config/rollover-config.component';
import {UtmApiDocComponent} from './utm-api-doc/utm-api-doc.component';

const routes: Routes = [
  {path: '', redirectTo: 'settings', pathMatch: 'full'},
  {
    path: 'settings',
    component: AppManagementComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE]
    },
    children: [
      {path: '', redirectTo: 'connection-key', pathMatch: 'full'},
      /*{
        path: 'license',
        loadChildren: '../license/license.module#LicenseModule',
        canActivate: [UserRouteAccessService, SaasRouteAccessService],
        data: {authorities: [ADMIN_ROLE]}
      },*/
      {
        path: 'rollover',
        component: RolloverConfigComponent,
        canActivate: [UserRouteAccessService, SaasRouteAccessService],
        data: {authorities: [ADMIN_ROLE]}
      },
      {
        path: 'menu-management',
        component: MenuComponent,
        canActivate: [UserRouteAccessService],
        data: {authorities: [ADMIN_ROLE]}
      },
      {
        path: 'app-logs',
        component: AppLogsComponent,
        canActivate: [UserRouteAccessService],
        data: {authorities: [ADMIN_ROLE]}
      },
      {
        path: 'index-pattern',
        component: IndexPatternListComponent,
        canActivate: [UserRouteAccessService, SaasRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE]
        },
      },
      {
        path: 'health-checks',
        component: HealthChecksComponent,
        canActivate: [UserRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE]
        },
      },
      {
        path: 'about',
        component: UtmApiDocComponent,
        canActivate: [UserRouteAccessService],
        data: {
          authorities: [USER_ROLE, ADMIN_ROLE]
        },
      },
      {
        path: 'user-access-audit',
        component: AuditsComponent,
        canActivate: [UserRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE]
        }
      },
      {
        path: 'index-management',
        component: IndexManagementComponent,
        canActivate: [UserRouteAccessService, SaasRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE]
        }
      },
      {
        path: 'connection-key',
        component: ConnectionKeyComponent,
        canActivate: [UserRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE]
        }
      },
      {
        path: 'application-config',
        component: AppConfigComponent,
        canActivate: [UserRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE]
        },
      },
      {
        path: 'application-theme',
        component: AppThemeComponent,
        canActivate: [UserRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE]
        },
      }],
  },
];


@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class AppManagementRoutingModule {
}

