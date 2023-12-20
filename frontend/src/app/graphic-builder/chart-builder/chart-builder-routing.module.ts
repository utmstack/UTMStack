import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../../shared/constants/global.constant';
import {ChartBuilderComponent} from './chart-builder.component';


const routes: Routes = [
  {path: '', redirectTo: 'chart-builder', pathMatch: 'full'},
  {
    path: 'chart-builder',
    component: ChartBuilderComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class ChartBuilderRoutingModule {
}

