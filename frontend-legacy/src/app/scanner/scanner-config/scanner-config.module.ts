import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {InputClassResolve} from '../../shared/util/input-class-resolve';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {ScannerSharedModule} from '../shared/scanner-shared.module';
import {CredentialCreateComponent} from './credential/credential-create/credential-create.component';
import {CredentialDeleteComponent} from './credential/credential-delete/credential-delete.component';
import {CredentialListComponent} from './credential/credential-list/credential-list.component';
import {PortCreateComponent} from './port/port-create/port-create.component';
import {PortDeleteComponent} from './port/port-delete/port-delete.component';
import {PortListComponent} from './port/port-list/port-list.component';
import {PortRangeCreateComponent} from './port/port-range/port-range-create/port-range-create.component';
import {PortRangeDeleteComponent} from './port/port-range/port-range-delete/port-range-delete.component';
import {PortRangeListComponent} from './port/port-range/port-range-list/port-range-list.component';
import {PortFilterComponent} from './port/shared/components/port-filter/port-filter.component';
import {ScannerConfigRoutingModule} from './scanner-config-routing.module';
import {ScannerConfigComponent} from './scanner-config.component';
import {ScheduleCreateComponent} from './schedule/schedule-create/schedule-create.component';
import {ScheduleDeleteComponent} from './schedule/schedule-delete/schedule-delete.component';
import {ScheduleListComponent} from './schedule/schedule-list/schedule-list.component';
import {ScheduleFilterComponent} from './schedule/shared/components/schedule-filter/schedule-filter.component';
import {TargetFilterComponent} from './target/shared/components/target-filter/target-filter.component';
import {TargetCreateComponent} from './target/target-create/target-create.component';
import {TargetDeleteComponent} from './target/target-delete/target-delete.component';
import {TargetDetailComponent} from './target/target-detail/target-detail.component';
import {TargetListComponent} from './target/target-list/target-list.component';
import {TaskFilterComponent} from './task/shared/components/task-filter/task-filter.component';
import {TaskStatusComponent} from './task/shared/components/task-status/task-status.component';
import {TaskCreateComponent} from './task/task-create/task-create.component';
import {TaskDeleteComponent} from './task/task-delete/task-delete.component';
import {TaskDetailComponent} from './task/task-detail/task-detail.component';
import {TaskListComponent} from './task/task-list/task-list.component';

@NgModule({
  imports: [
    CommonModule,
    ScannerConfigRoutingModule,
    NgbModule,
    UtmSharedModule,
    ReactiveFormsModule,
    FormsModule,
    NgSelectModule,
    ScannerSharedModule
  ],
  declarations: [
    ScannerConfigComponent,
    PortListComponent,
    ScheduleListComponent,
    TargetListComponent,
    TaskListComponent,
    TaskCreateComponent,
    TargetCreateComponent,
    PortCreateComponent,
    CredentialListComponent,
    CredentialCreateComponent,
    ScheduleCreateComponent,
    PortRangeCreateComponent,
    CredentialDeleteComponent,
    ScheduleCreateComponent,
    ScheduleDeleteComponent,
    PortDeleteComponent,
    PortRangeListComponent,
    PortRangeDeleteComponent,
    PortFilterComponent,
    ScheduleFilterComponent,
    TargetDeleteComponent,
    TargetFilterComponent,
    TaskDeleteComponent,
    TaskFilterComponent,
    TaskDetailComponent,
    TaskStatusComponent,
    TargetDetailComponent],
  entryComponents: [
    TaskCreateComponent,
    CredentialCreateComponent,
    PortCreateComponent,
    ScheduleCreateComponent,
    CredentialDeleteComponent,
    ScheduleDeleteComponent,
    PortDeleteComponent,
    PortRangeListComponent,
    PortRangeDeleteComponent,
    PortRangeDeleteComponent,
    PortRangeCreateComponent,
    TargetDeleteComponent,
    TaskDeleteComponent,
    TargetCreateComponent
  ],
  exports: [TargetCreateComponent],
  providers: [InputClassResolve]
})

export class ScannerConfigModule {
}
