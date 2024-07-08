import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import {RouterModule} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {AppCorrelationManagementComponent} from './app-correlation-management/app-correlation-management.component';
import {AssetsComponent} from './app-correlation-management/components/assets/assets.component';
import {PatternsComponent} from './app-correlation-management/components/patterns/patterns.component';
import {TypesComponent} from './app-correlation-management/components/types/types.component';
import {AppRuleComponent} from './app-rule/app-rule.component';
import {AddRuleComponent} from './app-rule/components/add-rule/add-rule.component';
import {RulesResolverService} from './services/rules.resolver.service';
import {RuleResolverService} from './services/rule.resolver.service';
import {TypesResolverService} from "./services/types.resolver.service";

const routes = [
  {path: '', redirectTo: 'rules', pathMatch: 'full'},
  {
    path: 'rules',
    component:  AppRuleComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]},
    resolve: {
      response: RulesResolverService
    }
  },
  {
    path: 'rule',
    component:  AddRuleComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]},
  },
  {
    path: 'rule/:id',
    component:  AddRuleComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]},
    resolve: {
      response: RuleResolverService
    }
  },
  {
    path: 'manage',
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
        resolve: {
          response: TypesResolverService
        }
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
