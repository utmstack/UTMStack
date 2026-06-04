import { NgModule } from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../../shared/constants/global.constant';
import {AdversaryViewComponent} from './adversary-view/adversary-view.component';

const routes: Routes = [
  {path: '', redirectTo: 'view'},
  {
    path: 'view',
    component: AdversaryViewComponent,
    canActivate: [UserRouteAccessService],
    runGuardsAndResolvers: 'always',
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE]
    }
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class AdversaryManagementRoutingModule { }
