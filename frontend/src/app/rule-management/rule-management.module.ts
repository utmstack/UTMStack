import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {ResizableModule} from 'angular-resizable-element';
import {AlertManagementSharedModule} from '../data-management/alert-management/shared/alert-management-shared.module';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {AppRuleComponent} from './app-rule/app-rule.component';
import {AddRuleComponent} from './components/add-rule/add-rule.component';
import {RuleDetailComponent} from './components/rule-list/components/rule-detail/rule-detail.component';
import {RuleFieldComponent} from './components/rule-list/components/rule-field/rule-field.component';
import { RuleListComponent } from './components/rule-list/rule-list.component';
import {RuleManagementRoutingModule} from './rule-management.routing.module';
import {RuleService} from './services/rule.service';
import {RulesResolverService} from './services/rules.resolver.service';
import {ReactiveFormsModule} from "@angular/forms";
import {NgSelectModule} from "@ng-select/ng-select";

@NgModule({
  declarations: [
      AppRuleComponent,
      RuleListComponent,
      RuleFieldComponent,
      RuleDetailComponent,
      AddRuleComponent
  ],
    imports: [
        CommonModule,
        NgbModule,
        AlertManagementSharedModule,
        ResizableModule,
        UtmSharedModule,
        RuleManagementRoutingModule,
        ReactiveFormsModule,
        NgSelectModule
    ],
  providers: [
      RuleService,
      RulesResolverService
  ],
   entryComponents: [
      AddRuleComponent
   ]
})
export class RuleManagementModule { }
