import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {ResizableModule} from 'angular-resizable-element';
import { InfiniteScrollModule } from 'ngx-infinite-scroll';
import {AlertManagementSharedModule} from '../data-management/alert-management/shared/alert-management-shared.module';
import {FileBrowserModule} from '../filebrowser/filebrowser.module';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {AppCorrelationManagementComponent} from './app-correlation-management/app-correlation-management.component';
import {AssetsComponent} from './app-correlation-management/components/assets/assets.component';
import {PatternsComponent} from './app-correlation-management/components/patterns/patterns.component';
import {SidebarComponent} from './app-correlation-management/components/sidebar/sidebar.component';
import {TypesComponent} from './app-correlation-management/components/types/types.component';
import {AppRuleComponent} from './app-rule/app-rule.component';
import {AddRuleComponent} from './app-rule/components/add-rule/add-rule.component';
import {RuleGenericFilterComponent} from './app-rule/components/rule-generic-filter/rule-generic-filter.component';
import {RuleDetailComponent} from './app-rule/components/rule-list/components/rule-detail/rule-detail.component';
import {RuleFieldComponent} from './app-rule/components/rule-list/components/rule-field/rule-field.component';
import {RuleListComponent} from './app-rule/components/rule-list/rule-list.component';
import {RuleManagementRoutingModule} from './rule-management.routing.module';
import { DataTypeService } from './services/data-type.service';
import {FilterService} from './services/filter.service';
import {RuleResolverService} from './services/rule.resolver.service';
import {RuleService} from './services/rule.service';
import {RulesResolverService} from './services/rules.resolver.service';
import {GenericFilterComponent} from './share/generic-filter/generic-filter.component';


@NgModule({
  declarations: [
      AppRuleComponent,
      RuleListComponent,
      RuleFieldComponent,
      RuleDetailComponent,
      AddRuleComponent,
      SidebarComponent,
      AppCorrelationManagementComponent,
      AssetsComponent,
      TypesComponent,
      PatternsComponent,
      RuleGenericFilterComponent,
      GenericFilterComponent
  ],
    imports: [
        CommonModule,
        NgbModule,
        AlertManagementSharedModule,
        ResizableModule,
        UtmSharedModule,
        RuleManagementRoutingModule,
        ReactiveFormsModule,
        NgSelectModule,
        InfiniteScrollModule,
        FileBrowserModule,
        FormsModule
    ],
  providers: [
      RuleService,
      RulesResolverService,
      RuleResolverService,
      FilterService,
      DataTypeService
  ],
   entryComponents: [
      AddRuleComponent
   ]
})
export class RuleManagementModule { }
