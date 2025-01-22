import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {ComplianceCustomViewComponent} from './compliance-custom-view/compliance-custom-view.component';
import {CpStandardManagementComponent} from './compliance-management/cp-standard-management/cp-standard-management.component';
import {ComplianceResultViewComponent} from './compliance-result-view/compliance-result-view.component';
import {ComplianceScheduleComponent} from './compliance-schedule/compliance-schedule.component';
import {ComplianceTemplatesComponent} from './compliance-templates/compliance-templates.component';
import {ComplianceReportViewerComponent} from "./compliance-report-viewer/compliance-report-viewer.component";
import { CompliancePrintViewComponent } from './compliance-reports-view/components/compliance-print-view/compliance-print-view.component';

const routes: Routes = [
  {path: '', redirectTo: 'templates'},
  {
    path: 'templates',
    component: ComplianceTemplatesComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'template-result',
    component: ComplianceResultViewComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'template-custom',
    component: ComplianceCustomViewComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'management',
    component: CpStandardManagementComponent,
    data: {authorities: [ADMIN_ROLE]}
  },
  {
    path: 'schedule',
    component: ComplianceScheduleComponent,
    data: {authorities: [ADMIN_ROLE]}
  },
  {
    path: 'report-viewer',
    component: ComplianceReportViewerComponent,
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'print-view',
    component: CompliancePrintViewComponent,
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: []
})
export class ComplianceRoutingModule {
}
