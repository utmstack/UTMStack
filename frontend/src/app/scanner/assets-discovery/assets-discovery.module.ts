import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {NgxEchartsModule} from 'ngx-echarts';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {ScannerConfigModule} from '../scanner-config/scanner-config.module';
import {ScannerSharedModule} from '../shared/scanner-shared.module';
import {AssetsDashboardComponent} from './assets-dashboard/assets-dashboard.component';
import {AssetsDiscoveryRoutingModule} from './assets-discovery-routing.module';
import {AssetsHostDetailComponent} from './assets-host-detail/assets-host-detail.component';
import {AssetsHostComponent} from './assets-host/assets-host.component';
import {AssetsHostFilterComponent} from './assets-host/shared/components/assets-host-filter/assets-host-filter.component';
import {AssetHostCardComponent} from './shared/components/asset-host-card/asset-host-card.component';
import {AssetHostTableComponent} from './shared/components/asset-host-table/asset-host-table.component';
import {VulnerabilityDetailComponent} from './shared/components/vulnerability-detail/vulnerability-detail.component';
import {TaskResultComponent} from './task-result/task-result.component';


@NgModule({
  declarations: [
    AssetsDashboardComponent,
    AssetsHostComponent,
    AssetHostCardComponent,
    AssetHostTableComponent,
    AssetsHostDetailComponent,
    AssetsHostFilterComponent,
    VulnerabilityDetailComponent,
    TaskResultComponent
  ],
  entryComponents: [VulnerabilityDetailComponent],
  imports: [
    CommonModule,
    NgbModule,
    UtmSharedModule,
    RouterModule,
    AssetsDiscoveryRoutingModule,
    ReactiveFormsModule,
    FormsModule,
    NgSelectModule,
    NgxEchartsModule,
    ScannerSharedModule,
    ScannerConfigModule
  ],
  exports: []
})
export class AssetsDiscoveryModule {
}
