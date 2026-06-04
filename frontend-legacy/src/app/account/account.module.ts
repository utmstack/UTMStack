import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule, NO_ERRORS_SCHEMA} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {accountState} from './account.route';
import {ActivateComponent} from './activate/activate.component';
import {RegisterComponent} from './register/register.component';

@NgModule({
  imports: [
    FormsModule,
    RouterModule.forChild(accountState),
    CommonModule,
    UtmSharedModule
  ],
  declarations: [
    ActivateComponent,
    RegisterComponent,
  ],
  entryComponents: [
    ActivateComponent,
    RegisterComponent],
  exports: [
    ActivateComponent,
    RegisterComponent],
  schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA]
})
export class UtmAccountModule {
}
