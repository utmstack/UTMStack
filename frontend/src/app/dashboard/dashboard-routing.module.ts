import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {ComplianceExportComponent} from './compliance-export/compliance-export.component';
import {DashboardExportCustomComponent} from './dashboard-export-custom/dashboard-export-custom.component';
import {DashboardExportPdfComponent} from './dashboard-export-pdf/dashboard-export-pdf.component';
import {DashboardLogSourcesComponent} from './dashboard-log-sources/dashboard-log-sources.component';
import {DashboardOverviewComponent} from './dashboard-overview/dashboard-overview.component';
import {DashboardRenderComponent} from './dashboard-render/dashboard-render.component';
import {DashboardViewComponent} from './dashboard-view/dashboard-view.component';
import {ReportExportComponent} from './report-export/report-export.component';
import {DashboardResolverService} from './shared/services/dashboard-resolver.service';

const routes: Routes = [
  {path: '', redirectTo: 'overview', pathMatch: 'full'},
  {
    path: 'overview',
    component: DashboardOverviewComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'log-sources',
    component: DashboardLogSourcesComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'view/:name',
    component: DashboardViewComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'render/:id/:dashboard',
    component: DashboardRenderComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]},
    resolve: {
      response: DashboardResolverService
    }
  },
  {
    path: 'export/:id/:dashboard',
    component: DashboardExportPdfComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'export-compliance/:id',
    component: ComplianceExportComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'export-report/:id/:report',
    component: ReportExportComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'customize-export/:id/:dashboard',
    component: DashboardExportCustomComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class DashboardRoutingModule {
}
