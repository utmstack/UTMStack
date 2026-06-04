import {CommonModule} from '@angular/common';
import {NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {GridsterModule} from 'angular-gridster2';
import {NgxDraggableResizerModule} from 'ngx-draggable-resizer';
import {AppManagementSharedModule} from '../../app-management/shared/app-management-shared.module';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {GraphicBuilderSharedModule} from '../shared/graphic-builder-shared.module';
import {VisualizationSharedModule} from '../visualization/visualization-shared.module';
import {DashboardBuilderRoutingModule} from './dashboard-builder-routing.module';
import {DashboardCreateComponent} from './dashboard-create/dashboard-create.component';
import {DashboardDeleteComponent} from './dashboard-delete/dashboard-delete.component';
import { DashboardFilterCreateComponent } from './dashboard-filter-create/dashboard-filter-create.component';
import {DashboardImportComponent} from './dashboard-import/dashboard-import.component';
import {DashboardListComponent} from './dashboard-list/dashboard-list.component';
import {DashboardSaveComponent} from './dashboard-save/dashboard-save.component';

@NgModule({
  declarations: [
    DashboardListComponent,
    DashboardCreateComponent,
    DashboardDeleteComponent,
    DashboardSaveComponent,
    DashboardImportComponent,
    DashboardFilterCreateComponent],
  imports: [
    CommonModule,
    DashboardBuilderRoutingModule,
    UtmSharedModule,
    NgbModule,
    NgxDraggableResizerModule,
    GridsterModule,
    GraphicBuilderSharedModule,
    VisualizationSharedModule,
    ReactiveFormsModule,
    NgSelectModule,
    FormsModule,
    AppManagementSharedModule
  ],
  entryComponents: [
    DashboardDeleteComponent,
    DashboardSaveComponent,
    DashboardFilterCreateComponent,
    DashboardImportComponent],
  schemas: [NO_ERRORS_SCHEMA]
})
export class DashboardBuilderModule {
}
