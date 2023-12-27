import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UtmModulesEnum} from '../app-module/shared/enum/utm-module.enum';
import {ActiveModuleRouteAccessService} from '../core/auth/active-module-route-access-service';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';

const routes: Routes = [
  {
    path: 'alert',
    loadChildren: './alert-management/alert-management.module#AlertManagementModule',
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },
  {
    path: 'file',
    loadChildren: './file-management/file-management.module#FileManagementModule',
    canActivate: [UserRouteAccessService, ActiveModuleRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE], module: UtmModulesEnum.FILE_INTEGRITY}
  }

];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class DataManagementRouting {
}

