import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {ResizableModule} from 'angular-resizable-element';
import {AlertManagementSharedModule} from '../data-management/alert-management/shared/alert-management-shared.module';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {AppRuleComponent} from './app-rule/app-rule.component';
import { RuleListComponent } from './rule-list/rule-list.component';
import {RuleManagementRoutingModule} from './rule-management.routing.module';
import {RuleService} from './services/rule.service';
import {RulesResolverService} from './services/rules.resolver.service';
import {RuleFieldComponent} from "./rule-list/components/rule-field/rule-field.component";
import {RuleDetailComponent} from "./rule-list/components/rule-detail/rule-detail.component";

@NgModule({
  declarations: [
      AppRuleComponent,
      RuleListComponent,
      RuleFieldComponent,
      RuleDetailComponent
  ],
  imports: [
    CommonModule,
    NgbModule,
    AlertManagementSharedModule,
    ResizableModule,
    UtmSharedModule,
    RuleManagementRoutingModule
  ],
  providers: [
      RuleService,
      RulesResolverService
  ]
})
export class RuleManagementModule { }
