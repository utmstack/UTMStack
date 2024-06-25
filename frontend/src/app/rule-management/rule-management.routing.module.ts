import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import {RouterModule} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {AppRuleComponent} from './app-rule/app-rule.component';
import {RulesResolverService} from './services/rules.resolver.service';
import {AppCorrelationManagementComponent} from './app-correlation-management/app-correlation-management.component';
import {AssetsComponent} from "./app-correlation-management/components/assets/assets.component";
import {TypesComponent} from "./app-correlation-management/components/types/types.component";
import {PatternsComponent} from "./app-correlation-management/components/patterns/patterns.component";

const routes = [
  {path: '', redirectTo: 'correlation-rules', pathMatch: 'full'},
  {
    path: 'correlation-rules',
    component:  AppRuleComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]},
    resolve: {
      rules: RulesResolverService
    }
  },
  {
    path: 'manage-correlation',
    component:  AppCorrelationManagementComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]},
    children: [
      {path: '', redirectTo: 'assets', pathMatch: 'full'},
      {
        path: 'assets',
        component:  AssetsComponent,
        canActivate: [UserRouteAccessService],
        data: {authorities: [USER_ROLE, ADMIN_ROLE]},
      },
      {
        path: 'types',
        component:  TypesComponent,
        canActivate: [UserRouteAccessService],
        data: {authorities: [USER_ROLE, ADMIN_ROLE]},
      },
      {
        path: 'patterns',
        component:  PatternsComponent,
        canActivate: [UserRouteAccessService],
        data: {authorities: [USER_ROLE, ADMIN_ROLE]},
      },
    ]
  }
];

@NgModule({
  declarations: [],
  imports: [
    RouterModule.forChild(routes),
    CommonModule
  ],
  exports: [
      RouterModule
  ]
})
export class RuleManagementRoutingModule { }
