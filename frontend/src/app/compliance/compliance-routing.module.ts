import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {ComplianceCustomViewComponent} from './compliance-custom-view/compliance-custom-view.component';
import {
  ComplianceLatestEvalPrintViewComponent
// tslint:disable-next-line:max-line-length
} from './compliance-latest-evaluations-view/components/compliance-latest-eval-print-view/compliance-latest-eval-print-view.component';
import {
  ComplianceLatestEvalDetailPrintViewComponent
// tslint:disable-next-line:max-line-length
} from './compliance-latest-evaluations-view/components/compliance-latest-evaluation-view-detail/compliance-latest-eval-detail-print-view/compliance-latest-eval-detail-print-view.component';
import {CpStandardManagementComponent} from './compliance-management/cp-standard-management/cp-standard-management.component';
import {ComplianceReportViewerComponent} from './compliance-report-viewer/compliance-report-viewer.component';
import { CompliancePrintViewComponent } from './compliance-reports-view/components/compliance-print-view/compliance-print-view.component';
import {ComplianceResultViewComponent} from './compliance-result-view/compliance-result-view.component';
import {ComplianceScheduleComponent} from './compliance-schedule/compliance-schedule.component';
import {ComplianceTemplatesComponent} from './compliance-templates/compliance-templates.component';

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
  {
    path: 'evaluations-print-view',
    component: ComplianceLatestEvalPrintViewComponent,
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'evaluation-detail-print-view/:id',
    component: ComplianceLatestEvalDetailPrintViewComponent,
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: []
})
export class ComplianceRoutingModule {
}
