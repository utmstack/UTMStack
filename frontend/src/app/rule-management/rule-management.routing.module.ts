import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import {RouterModule} from '@angular/router';
import {UserRouteAccessService} from '../core/auth/user-route-access-service';
import {ADMIN_ROLE, USER_ROLE} from '../shared/constants/global.constant';
import {AppRuleComponent} from './app-rule/app-rule.component';
import {RulesResolverService} from './services/rules.resolver.service';

const routes = [
  {path: '', redirectTo: 'rules', pathMatch: 'full'},
  {
    path: 'rules',
    component:  AppRuleComponent,
    canActivate: [UserRouteAccessService],
    data: {authorities: [USER_ROLE, ADMIN_ROLE]},
    resolve: {
      rules: RulesResolverService
    }
  },
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
