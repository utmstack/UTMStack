import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE} from '../shared/constants/global.constant';
import {RuleFileManagementComponent} from './rule-file-management/rule-file-management.component';


const routes: Routes = [
  {path: '', redirectTo: 'correlation-rules', pathMatch: 'full'},
  {
    path: 'correlation-rules',
    component: RuleFileManagementComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE]}
  }];


@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class FileBrowserRoutingModule {
}

