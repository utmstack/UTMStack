import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {LogAnalyzerTabsComponent} from './explorer/log-analyzer-tabs/log-analyzer-tabs.component';
import {LogAnalyzerQueryListComponent} from './queries/log-analyzer-query-list/log-analyzer-query-list.component';

const routes: Routes = [
  {path: '', redirectTo: 'log-analyzer', pathMatch: 'full'},
  {
    path: 'log-analyzer',
    component: LogAnalyzerTabsComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE, USER_ROLE]}
  },
  {
    path: 'log-analyzer-queries',
    component: LogAnalyzerQueryListComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE, USER_ROLE]}
  }


];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  declarations: [],
  exports: []
})

export class LogAnalyzerRoutingModule {
}

