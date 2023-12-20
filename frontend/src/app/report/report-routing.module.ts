import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {ReportCustomViewComponent} from './report-custom-view/report-custom-view.component';
import {ReportTemplateResultComponent} from './report-template-result/report-template-result.component';
import {ReportTemplateComponent} from './report-template/report-template.component';

const routes: Routes = [
  {path: '', redirectTo: 'templates', pathMatch: 'full'},
  {
    path: 'templates',
    component: ReportTemplateComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'template-result',
    component: ReportTemplateResultComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'template-view',
    component: ReportCustomViewComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  }


];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  declarations: [],
  exports: []
})

export class ReportRoutingModule {
}

