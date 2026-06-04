import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {JhiResolvePagingParams} from 'ng-jhipster';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {AdComponent} from './active-directory.component';
import {AdDashboardComponent} from './dashboard/active-directory-dashboard/active-directory-dashboard.component';
import {AdUserDetailComponent} from './dashboard/ad-user-detail/ad-user-detail.component';
// tslint:disable-next-line:max-line-length
import {AdReportsComponent} from './reports/active-directory-reports.component';
import {AdReportListComponent} from './reports/ad-report-list/ad-report-list.component';
// tslint:disable-next-line:max-line-length
import {AdScheduleListComponent} from './reports/schedule/ad-schedule-list/ad-schedule-list.component';
import {AdTrackerListComponent} from './tracker/ad-tracker-list/ad-tracker-list.component';
import {AdViewComponent} from './view/active-directory-view/active-directory-view.component';


const routes: Routes = [
  {path: '', redirectTo: 'view'},
  {
    path: '', component: AdComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE],
    }
    , children: [
      {
        path: 'view', component: AdViewComponent,
        canActivate: [UserRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE, USER_ROLE],
        },
      },
      {
        path: 'tracker', component: AdTrackerListComponent,
        canActivate: [UserRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE, USER_ROLE],
          defaultSort: 'id,asc'
        },
        resolve: {
          pagingParams: JhiResolvePagingParams
        }
      },
    ]
  },
  {
    path: 'overview', component: AdDashboardComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE],
    }
  },
  {
    path: 'detail/users', component: AdUserDetailComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE],
    }
  },
  {
    path: 'reports', component: AdReportsComponent,
    canActivate: [UserRouteAccessService],
    children: [
      {path: '', redirectTo: 'list'},
      {
        path: 'list', component: AdReportListComponent,
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
        path: 'schedule', component: AdScheduleListComponent,
        canActivate: [UserRouteAccessService],
        data: {
          authorities: [ADMIN_ROLE, USER_ROLE],
          defaultSort: 'id,asc'
        },
        resolve: {
          pagingParams: JhiResolvePagingParams
        }
      },
    ]
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class ActiveDirectoryRoutingModule {
}

