import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE} from '../shared/constants/global.constant';
import {LogstashPipelinesComponent} from './logstash-pipelines/logstash-pipelines.component';

const routes: Routes = [
  {
    path: 'pipelines',
    component: LogstashPipelinesComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE]
    },
  },
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class LogstashRoutingModule {
}

