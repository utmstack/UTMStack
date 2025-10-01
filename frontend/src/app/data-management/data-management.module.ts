import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {NewAlertBehavior} from '../shared/behaviors/new-alert.behavior';
import {DataManagementRouting} from './data-management-routing.module';
import {UtmSharedModule}  from 'src/app/shared/utm-shared.module'

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    DataManagementRouting,
    UtmSharedModule
  ],
  providers: [NewAlertBehavior]
})
export class DataManagementModule {
}
