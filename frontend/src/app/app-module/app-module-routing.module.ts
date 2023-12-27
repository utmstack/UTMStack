import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {AppModuleViewComponent} from './app-module-view/app-module-view.component';


const routes: Routes = [
  {path: '', redirectTo: 'explore', pathMatch: 'full'},
  {
    path: 'explore',
    component: AppModuleViewComponent,
    canActivate: [UserRouteAccessService],
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE]
    },
  }];


@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class AppModuleRoutingModuleRouting {
}

