import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {NgxEchartsModule} from 'ngx-echarts';
import {InputClassResolve} from '../../shared/util/input-class-resolve';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {GraphicBuilderSharedModule} from '../shared/graphic-builder-shared.module';
import {VisualizationRoutingModule} from './visualization-routing.module';
import {VisualizationSharedModule} from './visualization-shared.module';

@NgModule({
  declarations: [],
  entryComponents: [],
  imports: [
    CommonModule,
    NgxEchartsModule,
    RouterModule,
    VisualizationRoutingModule,
    UtmSharedModule,
    NgbModule,
    NgSelectModule,
    FormsModule,
    GraphicBuilderSharedModule,
    ReactiveFormsModule,
    VisualizationSharedModule
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
  exports: [],
  providers: [InputClassResolve]
})
export class VisualizationModule {
}
