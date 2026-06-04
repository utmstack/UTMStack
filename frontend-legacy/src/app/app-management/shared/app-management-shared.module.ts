import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {NgSelectModule} from '@ng-select/ng-select';
import {UtmSharedModule} from '../../shared/utm-shared.module';

@NgModule({
  declarations: [],
  imports: [
    CommonModule,
    FormsModule,
    UtmSharedModule,
    NgSelectModule
  ],
  entryComponents: [],
  exports: [],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA]
})
export class AppManagementSharedModule {
}
