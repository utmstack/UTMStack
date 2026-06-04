import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {InlineSVGModule} from 'ng-inline-svg';
import {MonacoEditorModule} from 'ngx-monaco-editor';
import {AlertManagementSharedModule} from '../data-management/alert-management/shared/alert-management-shared.module';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {LogstashFilterCreateComponent} from './logstash-filters/logstash-filter-create/logstash-filter-create.component';
import {LogstashFiltersComponent} from './logstash-filters/logstash-filters.component';
import {LogstashPipelinesComponent} from './logstash-pipelines/logstash-pipelines.component';
import {LogstashRoutingModule} from './logstash-routing.module';

@NgModule({
  declarations: [LogstashFiltersComponent, LogstashFilterCreateComponent, LogstashPipelinesComponent],
  imports: [
    CommonModule,
    UtmSharedModule,
    FormsModule,
    NgbModule,
    LogstashRoutingModule,
    InlineSVGModule,
    AlertManagementSharedModule,
    NgSelectModule,
    ReactiveFormsModule,
    MonacoEditorModule
  ],
  entryComponents: [LogstashFilterCreateComponent]
})
export class LogstashModule {
}
