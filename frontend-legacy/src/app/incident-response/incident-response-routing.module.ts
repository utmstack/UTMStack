import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE} from '../shared/constants/global.constant';
import {IncidentResponseViewComponent} from './incident-response-view/incident-response-view.component';
import {PlaybookBuilderComponent} from './playbook-builder/playbook-builder.component';
import {PlaybooksComponent} from './playbooks/playbooks.component';
import {InteractiveConsoleComponent} from "./interactive-console/interactive-console.component";

const routes: Routes = [
  {path: '', redirectTo: 'flows', pathMatch: 'full'},
  {
    path: 'audit',
    component: IncidentResponseViewComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE]}
  },
  {
    path: 'create-flow',
    component: PlaybookBuilderComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE]}
  },
  {
    path: 'flows',
    component: PlaybooksComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE]}
  },
  {
    path: 'interactive-console',
    component: InteractiveConsoleComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [ADMIN_ROLE]}
  },

];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class IncidentResponseRoutingModule {
}

