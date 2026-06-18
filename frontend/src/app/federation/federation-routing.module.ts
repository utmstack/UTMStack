import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {WelcomeComponent} from './pages/welcome/welcome.component';

const routes: Routes = [
  {
    path: 'federation',
    children: [
      {path: '', redirectTo: 'welcome', pathMatch: 'full'},
      {
        path: 'welcome',
        component: WelcomeComponent,
        data: {authorities: ['ROLE_USER']},
        canActivate: [UserRouteAccessService]
      },
      {
        path: 'instances',
        component: WelcomeComponent,
        data: {authorities: ['ROLE_USER']},
        canActivate: [UserRouteAccessService]
      }
    ]
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class FederationRoutingModule {}
