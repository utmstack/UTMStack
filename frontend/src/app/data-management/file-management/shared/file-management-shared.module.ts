import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {TranslateModule} from '@ngx-translate/core';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {NgxJsonViewerModule} from 'ngx-json-viewer';
import {IncidentResponseSharedModule} from '../../../incident-response/shared/incident-response-shared.module';
import {UtmSharedModule} from '../../../shared/utm-shared.module';
import {DataMgmtSharedModule} from '../../data-mgmt-shared/data-mgmt-shared.module';
import {FileAccessMaskComponent} from './components/file-access-mask/file-access-mask.component';
import {FileDataRenderComponent} from './components/file-data-render/file-data-render.component';
import {FileHostOsComponent} from './components/file-host-os/file-host-os.component';
import {FileObjectTypeComponent} from './components/file-object-type/file-object-type.component';
import {FileSaveDataComponent} from './components/file-save-data/file-save-data.component';
import {FileSddlViewComponent} from './components/file-sddl-view/file-sddl-view.component';
import {FileActiveFiltersComponent} from './components/filters/file-active-filters/file-active-filters.component';
import {FileFilterAppliedComponent} from './components/filters/file-filter-applied/file-filter-applied.component';
import {FileFilterComponent} from './components/filters/file-filter/file-filter.component';
import {FileGenericFilterComponent} from './components/filters/file-generic-filter/file-generic-filter.component';

@NgModule({
  declarations: [FileDataRenderComponent,
    FileHostOsComponent,
    FileObjectTypeComponent,
    FileFilterComponent,
    FileGenericFilterComponent,
    FileAccessMaskComponent,
    FileFilterAppliedComponent,
    FileActiveFiltersComponent,
    FileSaveDataComponent,
    FileSddlViewComponent],
  entryComponents: [FileSaveDataComponent],
  exports: [
    FileDataRenderComponent,
    FileFilterComponent,
    FileFilterAppliedComponent
  ],
  imports: [
    CommonModule,
    TranslateModule,
    UtmSharedModule,
    FormsModule,
    NgbModule,
    ReactiveFormsModule,
    NgSelectModule,
    InfiniteScrollModule,
    NgxJsonViewerModule,
    IncidentResponseSharedModule,
    DataMgmtSharedModule
  ]
})
export class FileManagementSharedModule {
}
