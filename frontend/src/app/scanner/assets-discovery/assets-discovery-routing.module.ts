import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {JhiResolvePagingParams} from 'ng-jhipster';
import {UserRouteAccessService} from '../../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../../shared/constants/global.constant';
import {AssetsDashboardComponent} from './assets-dashboard/assets-dashboard.component';
import {AssetsHostDetailComponent} from './assets-host-detail/assets-host-detail.component';
import {AssetsHostComponent} from './assets-host/assets-host.component';
import {TaskResultComponent} from './task-result/task-result.component';

const routes: Routes = [
  {path: '', redirectTo: 'dashboard'},
  {
    path: 'dashboard',
    component: AssetsDashboardComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'assets',
    component: AssetsHostComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE],
      defaultSort: 'id,asc'
    },
    resolve: {
      pagingParams: JhiResolvePagingParams
    }
  },
  {
    path: 'assets-detail',
    component: AssetsHostDetailComponent,
    canActivate: [UserRouteAccessService]
  },
  {
    path: 'task-result',
    component: TaskResultComponent,
    canActivate: [UserRouteAccessService]
  }

];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class AssetsDiscoveryRoutingModule {
}

