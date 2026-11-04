import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {TextBuilderComponent} from './text-builder/text-builder.component';


const routes: Routes = [
  {path: '', redirectTo: 'visualization', pathMatch: 'full'},
  {
    path: 'visualization',
    loadChildren: () => import('./visualization/visualization.module').then(m => m.VisualizationModule),
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'builder',
    loadChildren: () => import('./chart-builder/chart-builder.module').then(m => m.ChartBuilderModule),
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'builder/text-builder',
    component: TextBuilderComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'dashboard',
    loadChildren: () => import('./dashboard-builder/dashboard-builder.module').then(m => m.DashboardBuilderModule),
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class GraphicBuilderRoutingModule {
}

