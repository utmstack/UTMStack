import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../../shared/constants/global.constant';
import {AlertFullDetailComponent} from './alert-full-detail/alert-full-detail.component';
import {AlertReportViewComponent} from './alert-report-view/alert-report-view.component';
import {AlertReportsComponent} from './alert-reports/alert-reports.component';
import {AlertRulesComponent} from './alert-rules/alert-rules.component';
import {AlertViewComponent} from './alert-view/alert-view.component';

const routes: Routes = [
  {
    path: 'view', component: AlertViewComponent,
    canActivate: [UserRouteAccessService],
    runGuardsAndResolvers: 'always',
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE]
    }
  },
  {
    path: 'detail/:id', component: AlertFullDetailComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE]
    }
  },
  {
    path: 'report/view/:id/:name', component: AlertReportViewComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE]
    }
  },
  {
    path: 'reports/list', component: AlertReportsComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE]
    }
  },
  {
    path: 'alert-rule-management',
    component: AlertRulesComponent,
    data: {authorities: [ADMIN_ROLE]}
  },

];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class AlertManagementRouting {
}

