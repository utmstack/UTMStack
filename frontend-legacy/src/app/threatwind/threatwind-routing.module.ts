import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE} from '../shared/constants/global.constant';
import {ThreatwindComponent} from './threatwind.component';


const routes: Routes = [
  {
    path: '',
    component: ThreatwindComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE]}
  }];


@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class ThreatWindRoutingModule {
}

