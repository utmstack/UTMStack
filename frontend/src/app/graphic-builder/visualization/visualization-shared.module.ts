import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {InputClassResolve} from '../../shared/util/input-class-resolve';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {GraphicBuilderSharedModule} from '../shared/graphic-builder-shared.module';
import {VisualizationChangeNameComponent} from './shared/components/visualization-change-name/visualization-change-name.component';
import {VisualizationFilterComponent} from './shared/components/visualization-filter/visualization-filter.component';
import {VisualizationSelectComponent} from './shared/components/visualization-select/visualization-select.component';
import {VisualizationCreateComponent} from './visualization-create/visualization-create.component';
import {VisualizationDeleteComponent} from './visualization-delete/visualization-delete.component';
import {VisualizationImportComponent} from './visualization-import/visualization-import.component';
import {VisualizationListComponent} from './visualization-list/visualization-list.component';
import {VisualizationSaveComponent} from './visualization-save/visualization-save.component';

@NgModule({
  declarations: [
    VisualizationListComponent,
    VisualizationDeleteComponent,
    VisualizationCreateComponent,
    VisualizationSaveComponent,
    VisualizationImportComponent,
    VisualizationFilterComponent,
    VisualizationChangeNameComponent,
    VisualizationSelectComponent
  ],
  entryComponents: [
    VisualizationDeleteComponent,
    VisualizationCreateComponent,
    VisualizationSaveComponent,
    VisualizationImportComponent,
    VisualizationChangeNameComponent,
    VisualizationSelectComponent],
  imports: [
    CommonModule,
    UtmSharedModule,
    NgbModule,
    NgSelectModule,
    FormsModule,
    ReactiveFormsModule,
    GraphicBuilderSharedModule,

  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
  exports: [
    VisualizationChangeNameComponent,
    VisualizationSelectComponent
  ],
  providers: [InputClassResolve]
})
export class VisualizationSharedModule {
}
