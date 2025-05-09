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
import {AddAssetComponent} from './app-correlation-management/components/assets/components/components/add-asset/add-asset.component';
import {
  AssetDetailComponent
} from './app-correlation-management/components/assets/components/components/asset-detail/asset-detail.component';
import {AddPatternComponent} from './app-correlation-management/components/patterns/components/add-pattern.component';
import {PatternsComponent} from './app-correlation-management/components/patterns/patterns.component';
import {SidebarComponent} from './app-correlation-management/components/sidebar/sidebar.component';
import {AddTypeComponent} from './app-correlation-management/components/types/components/add-type.component';
import {TypesComponent} from './app-correlation-management/components/types/types.component';
import {AssetManagerService} from './app-correlation-management/services/asset-manager.service';
import {AssetsResolverService} from './app-correlation-management/services/assets.resolver.service';
import {ConfigService} from './app-correlation-management/services/config.service';
import {PatternManagerService} from './app-correlation-management/services/pattern-manager.service';
import {PatternsResolverService} from './app-correlation-management/services/patterns.resolver.service';
import {TypesResolverService} from './app-correlation-management/services/types.resolver.service';
import {AppRuleComponent} from './app-rule/app-rule.component';
import {AddAfterEventComponent} from './app-rule/components/add-after-event/add-after-event.component';
import {AddRuleComponent} from './app-rule/components/add-rule/add-rule.component';
import {AddVariableComponent} from './app-rule/components/add-variable/add-variable.component';
import {DeduplicateFieldsComponent} from './app-rule/components/deduplicate-fields/deduplicate-fields.component';
import {ExpressionConsoleComponent} from './app-rule/components/expression-console/expression-console.component';
import {ImportRuleComponent} from './app-rule/components/import-rules/import-rule.component';
import {ImportRuleService} from './app-rule/components/import-rules/import-rule.service';
import {AddReferenceComponent} from './app-rule/components/reference/add-reference.component';
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
      ImportRuleComponent,
      SidebarComponent,
      AppCorrelationManagementComponent,
      AssetsComponent,
      TypesComponent,
      PatternsComponent,
      RuleGenericFilterComponent,
      GenericFilterComponent,
      AddTypeComponent,
      AddPatternComponent,
      AddAssetComponent,
      AssetDetailComponent,
      AddReferenceComponent,
      AddVariableComponent,
      AddAfterEventComponent,
      DeduplicateFieldsComponent,
      ExpressionConsoleComponent

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
      DataTypeService,
      TypesResolverService,
      ConfigService,
      PatternManagerService,
      PatternsResolverService,
      AssetsResolverService,
      AssetManagerService,
      ImportRuleService
  ],
   entryComponents: [
       AddTypeComponent,
       AddPatternComponent,
       AddAssetComponent,
       ImportRuleComponent
   ]
})
export class RuleManagementModule { }
