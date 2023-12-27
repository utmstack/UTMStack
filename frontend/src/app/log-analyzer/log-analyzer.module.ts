import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {ResizableModule} from 'angular-resizable-element';
import {NgxEchartsModule} from 'ngx-echarts';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {NgxJsonViewerModule} from 'ngx-json-viewer';
import {ComplianceSharedModule} from '../compliance/shared/compliance-shared.module';
import {InputClassResolve} from '../shared/util/input-class-resolve';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {LogAnalyzerFieldCardComponent} from './explorer/log-analyzer-field/log-analyzer-field-card/log-analyzer-field-card.component';
// tslint:disable-next-line:max-line-length
import {LogAnalyzerFieldDetailComponent} from './explorer/log-analyzer-field/log-analyzer-field-card/log-analyzer-field-detail/log-analyzer-field-detail.component';
import {LogAnalyzerFieldComponent} from './explorer/log-analyzer-field/log-analyzer-field.component';
import {LogAnalyzerTabsComponent} from './explorer/log-analyzer-tabs/log-analyzer-tabs.component';
import {LogAnalyzerViewComponent} from './explorer/log-analyzer-view/log-analyzer-view.component';
import {LogAnalyzerRoutingModule} from './log-analyzer-routing.module';
import {LogAnalyzerQueryCreateComponent} from './queries/log-analyzer-query-create/log-analyzer-query-create.component';
import {LogAnalyzerQueryDeleteComponent} from './queries/log-analyzer-query-delete/log-analyzer-query-delete.component';
import {LogAnalyzerQueryListComponent} from './queries/log-analyzer-query-list/log-analyzer-query-list.component';
import {AnalyzerBarChartComponent} from './shared/components/analyzer-bar-chart/analyzer-bar-chart.component';
import {TabContentComponent} from './shared/components/tab-content/tab-content.component';
import {TabContentDirective} from './shared/directive/tab-content.directive';
import {TabService} from './shared/services/tab.service';

@NgModule({
  declarations: [
    LogAnalyzerViewComponent,
    LogAnalyzerFieldComponent,
    LogAnalyzerFieldDetailComponent,
    LogAnalyzerFieldCardComponent,
    TabContentDirective,
    LogAnalyzerTabsComponent,
    TabContentComponent,
    AnalyzerBarChartComponent,
    LogAnalyzerQueryListComponent,
    LogAnalyzerQueryCreateComponent,
    LogAnalyzerQueryDeleteComponent,
    LogAnalyzerFieldCardComponent],
  imports: [
    CommonModule,
    LogAnalyzerRoutingModule,
    NgSelectModule,
    FormsModule,
    NgbModule,
    UtmSharedModule,
    NgxJsonViewerModule,
    NgxEchartsModule,
    RouterModule,
    ReactiveFormsModule,
    ComplianceSharedModule,
    ResizableModule,
    InfiniteScrollModule
  ],
  entryComponents: [
    LogAnalyzerViewComponent,
    LogAnalyzerQueryCreateComponent,
    LogAnalyzerQueryDeleteComponent,
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
  exports: [],
  providers: [TabService, InputClassResolve]
})
export class LogAnalyzerModule {
}
