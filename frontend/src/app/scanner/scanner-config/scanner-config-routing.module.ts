import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {JhiResolvePagingParams} from 'ng-jhipster';
import {UserRouteAccessService} from '../../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../../shared/constants/global.constant';
import {CredentialListComponent} from './credential/credential-list/credential-list.component';
import {PortListComponent} from './port/port-list/port-list.component';
import {ScannerConfigComponent} from './scanner-config.component';
import {ScheduleListComponent} from './schedule/schedule-list/schedule-list.component';
import {TargetListComponent} from './target/target-list/target-list.component';
import {TaskListComponent} from './task/task-list/task-list.component';


const routes: Routes = [
  {path: '', redirectTo: 'task'},
  {
    path: '', component: ScannerConfigComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE],
    }
    , children: [
      {
        path: 'task', component: TaskListComponent,
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
        path: 'target', component: TargetListComponent,
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
        path: 'schedule', component: ScheduleListComponent,
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
        path: 'port', component: PortListComponent,
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
        path: 'credential', component: CredentialListComponent,
        canActivate: [UserRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE, USER_ROLE],
          defaultSort: 'id,asc',
        },
        resolve: {
          pagingParams: JhiResolvePagingParams
        }
      },
    ]
  },

];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class ScannerConfigRoutingModule {
}

