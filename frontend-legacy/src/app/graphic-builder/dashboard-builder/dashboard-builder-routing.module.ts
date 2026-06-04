import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../../shared/constants/global.constant';
import {DashboardCreateComponent} from './dashboard-create/dashboard-create.component';
import {DashboardListComponent} from './dashboard-list/dashboard-list.component';


const routes: Routes = [
  {path: '', redirectTo: 'list', pathMatch: 'full'},
  {
    path: 'list',
    component: DashboardListComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'builder',
    component: DashboardCreateComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class DashboardBuilderRoutingModule {
}
