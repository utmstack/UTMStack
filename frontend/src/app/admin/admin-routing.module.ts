import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {JhiResolvePagingParams} from 'ng-jhipster';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE} from '../shared/constants/global.constant';
import {UserMgmtComponent} from './user/user-list/user-management.component';
import {UserMgmtResolve} from './user/user-management.route';
import {UserMgmtUpdateComponent} from './user/user-update/user-management-update.component';


const routes: Routes = [
  {
    path: 'user', component: UserMgmtComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE],
      defaultSort: 'id,asc'
    },
    resolve: {
      pagingParams: JhiResolvePagingParams
    }
  },
  {
    path: 'new-user',
    component: UserMgmtUpdateComponent,
    resolve: {
      user: UserMgmtResolve
    }
  },
  {
    path: 'user/:login/edit',
    component: UserMgmtUpdateComponent,
    resolve: {
      user: UserMgmtResolve
    }
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class AdminRoutingModule {
}

