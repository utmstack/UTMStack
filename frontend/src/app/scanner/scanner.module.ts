import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectModule} from '@ng-select/ng-select';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {ScannerRoutingModule} from './scanner-routing.module';
import {ScannerSharedModule} from './shared/scanner-shared.module';

@NgModule({
  imports: [
    CommonModule,
    ScannerRoutingModule,
    NgbModule,
    UtmSharedModule,
    ReactiveFormsModule,
    FormsModule,
    NgSelectModule,
    ScannerSharedModule
  ],
  providers: [],

})

export class ScannerModule {
}
