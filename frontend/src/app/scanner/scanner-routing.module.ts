import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';

const routes: Routes = [
  {path: '', redirectTo: 'assets-discovery'},
  {
    path: 'config',
    loadChildren: () => import('./scanner-config/scanner-config.module').then(m => m.ScannerConfigModule),
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'assets-discovery',
    loadChildren: () => import('./assets-discovery/assets-discovery.module').then(m => m.AssetsDiscoveryModule),
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },

];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class ScannerRoutingModule {
}

