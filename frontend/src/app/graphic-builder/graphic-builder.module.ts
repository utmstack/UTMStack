import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {AngularEditorModule} from '@kolkov/angular-editor';
import {NgxSpinnerModule} from 'ngx-spinner';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {GraphicBuilderRoutingModule} from './graphic-builder-routing.module';
import {TextBuilderComponent} from './text-builder/text-builder.component';
import {VisualizationSharedModule} from './visualization/visualization-shared.module';

@NgModule({
  declarations: [TextBuilderComponent],
  imports: [
    CommonModule,
    NgxSpinnerModule,
    GraphicBuilderRoutingModule,
    AngularEditorModule,
    FormsModule,
    UtmSharedModule,
    VisualizationSharedModule
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA]
})
export class GraphicBuilderModule {
}
