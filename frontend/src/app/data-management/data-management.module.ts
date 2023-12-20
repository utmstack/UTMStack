import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {NewAlertBehavior} from '../shared/behaviors/new-alert.behavior';
import {DataManagementRouting} from './data-management-routing.module';

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    DataManagementRouting
  ],
  providers: [NewAlertBehavior]
})
export class DataManagementModule {
}
