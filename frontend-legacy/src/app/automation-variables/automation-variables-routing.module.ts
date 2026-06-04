import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {ActiveModuleRouteAccessService} from '../core/auth/active-module-route-access-service';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {VariablesComponent} from "./variables/variables.component";

const routes: Routes = [
  {path: '', redirectTo: 'overview'},
  {
    path: 'list',
    component: VariablesComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]}
  },

];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class AutomationVariablesRoutingModule {
}

