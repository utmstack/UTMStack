import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {ResizableModule} from 'angular-resizable-element';
import {NgxEchartsModule} from 'ngx-echarts';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {IncidentResponseSharedModule} from '../incident-response/shared/incident-response-shared.module';
import {IncidentSharedModule} from '../incident/incident-shared/incident-shared.module';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {AssetGroupCreateComponent} from './asset-groups/asset-group-create/asset-group-create.component';
import {AssetGroupsComponent} from './asset-groups/asset-groups.component';
import {AssetGroupDetailComponent} from './asset-groups/shared/components/asset-group-detail/asset-group-detail.component';
import {AssetGroupMetricsComponent} from './asset-groups/shared/components/asset-group-metrics/asset-group-metrics.component';
import {AssetsDiscoverRoutingModule} from './assets-discover-routing.module';
import {AssetsViewComponent} from './assets-view/assets-view.component';
import {AssetByGroupComponent} from './shared/components/asset-by-group/asset-by-group.component';
import {AssetCreateComponent} from './shared/components/asset-create/asset-create.component';
import {AssetDetailComponent} from './shared/components/asset-detail/asset-detail.component';
import {AssetEditAliasComponent} from './shared/components/asset-edit-alias/asset-edit-alias.component';
import {AssetFieldRenderComponent} from './shared/components/asset-field-render/asset-field-render.component';
import {AssetGroupAddComponent} from './shared/components/asset-group-add/asset-group-add.component';
import {AssetIpComponent} from './shared/components/asset-ip/asset-ip.component';
import {AssetIsAliveComponent} from './shared/components/asset-is-alive/asset-is-alive.component';
import {AssetMetricsComponent} from './shared/components/asset-metrics/asset-metrics.component';
import {AssetOsComponent} from './shared/components/asset-os/asset-os.component';
import {AssetPortCreateComponent} from './shared/components/asset-port-create/asset-port-create.component';
import {AssetSaveReportComponent} from './shared/components/asset-save-report/asset-save-report.component';
import {AssetSeverityComponent} from './shared/components/asset-severity/asset-severity.component';
import {AssetSoftwareAddComponent} from './shared/components/asset-software-add/asset-software-add.component';
import {AssetSoftwareSelectComponent} from './shared/components/asset-software-select/asset-software-select.component';
import {AssetSoftwaresComponent} from './shared/components/asset-softwares/asset-softwares.component';
import {AssetStatusComponent} from './shared/components/asset-status/asset-status.component';
import {AssetViewHelpComponent} from './shared/components/asset-view-help/asset-view-help.component';
import {AssetsApplyNoteComponent} from './shared/components/assets-apply-note/assets-apply-note.component';
import {AssetsApplyTypeComponent} from './shared/components/assets-apply-type/assets-apply-type.component';
import {AssetsPortsComponent} from './shared/components/assets-ports/assets-ports.component';
import {AssetActiveFilterComponent} from './shared/components/filters/asset-active-filter/asset-active-filter.component';
import {AssetFilterApplyingComponent} from './shared/components/filters/asset-filter-applying/asset-filter-applying.component';
import {AssetFilterIsAliveComponent} from './shared/components/filters/asset-filter-is-alive/asset-filter-is-alive.component';
import {AssetGenericFilterComponent} from './shared/components/filters/asset-generic-filter/asset-generic-filter.component';
import {AssetsFilterComponent} from './shared/components/filters/assets-filter/assets-filter.component';
import {SourceDataTypeConfigComponent} from './source-data-type-config/source-data-type-config.component';
import {AssetsApplyTypeModule} from "./shared/components/assets-apply-type/assets-apply-type.module";
import {AssetsApplyNoteModule} from "./shared/components/assets-apply-note/assets-apply-note.module";
import {AssetsGroupAddModule} from "./shared/components/asset-group-add/assets-group-add.module";
import {InlineSVGModule} from "ng-inline-svg";

@NgModule({
  declarations: [
    AssetsViewComponent,
    AssetsFilterComponent,
    AssetStatusComponent,
    AssetSeverityComponent,
    AssetFieldRenderComponent,
    AssetOsComponent,
    AssetIsAliveComponent,
    AssetFilterIsAliveComponent,
    AssetGenericFilterComponent,
    AssetIpComponent,
    AssetDetailComponent,
    // AssetsApplyTypeComponent,
    // AssetsApplyNoteComponent,
    AssetMetricsComponent,
    AssetsPortsComponent,
    AssetSoftwaresComponent,
    AssetActiveFilterComponent,
    AssetFilterApplyingComponent,
    AssetSaveReportComponent,
    AssetViewHelpComponent,
    AssetGroupsComponent,
    AssetGroupCreateComponent,
    //AssetGroupAddComponent,
    AssetGroupMetricsComponent,
    AssetGroupDetailComponent,
    AssetByGroupComponent,
    AssetCreateComponent,
    AssetSoftwareSelectComponent,
    AssetSoftwareAddComponent,
    AssetPortCreateComponent,
    AssetEditAliasComponent,
    SourceDataTypeConfigComponent],
  entryComponents: [
    AssetSaveReportComponent,
    AssetViewHelpComponent,
    AssetSoftwareAddComponent,
    AssetGroupCreateComponent,
    AssetCreateComponent,
    AssetDetailComponent,
    SourceDataTypeConfigComponent],
    imports: [
        CommonModule,
        AssetsDiscoverRoutingModule,
        UtmSharedModule,
        NgbModule,
        ResizableModule,
        InfiniteScrollModule,
        FormsModule,
        NgSelectModule,
        NgxEchartsModule,
        ReactiveFormsModule,
        IncidentResponseSharedModule,
        IncidentSharedModule,
        AssetsApplyNoteModule,
        AssetsApplyTypeModule,
        AssetsGroupAddModule,
        InlineSVGModule
    ],
  schemas: [NO_ERRORS_SCHEMA, CUSTOM_ELEMENTS_SCHEMA]
})
export class AssetsDiscoverModule {
}
