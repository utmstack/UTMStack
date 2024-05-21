import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {AssetGroupsComponent} from './asset-groups/asset-groups.component';
import {AssetsViewComponent} from './assets-view/assets-view.component';
import {CollectorsViewComponent} from "./collectors-view/collectors-view.component";

const routes: Routes = [
  {path: '', redirectTo: 'sources'},
  {
    path: 'sources', component: AssetsViewComponent,
    canActivate: [UserRouteAccessService],
    runGuardsAndResolvers: 'always',
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE]
    }
  },
  {
    path: 'sources-groups', component: AssetGroupsComponent,
    canActivate: [UserRouteAccessService],
    runGuardsAndResolvers: 'always',
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE]
    }
  },
  {
    path: 'collectors', component: CollectorsViewComponent,
    canActivate: [UserRouteAccessService],
    runGuardsAndResolvers: 'always',
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE]
    }
  },
  {
    path: 'collectors-groups', component: AssetGroupsComponent,
    canActivate: [UserRouteAccessService],
    runGuardsAndResolvers: 'always',
    data: {
      authorities: [ADMIN_ROLE, USER_ROLE]
    }
  },
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class AssetsDiscoverRoutingModule {
}

