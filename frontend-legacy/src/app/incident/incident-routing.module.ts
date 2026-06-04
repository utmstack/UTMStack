import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {IncidentManagementComponent} from './incident-management/incident-management.component';

const routes: Routes = [
  {path: '', redirectTo: 'view'},
  {
    path: 'view',
    component: IncidentManagementComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE, USER_ROLE]}
  }

];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class IncidentRoutingModule {
}

