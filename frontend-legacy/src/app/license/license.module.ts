import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {ReactiveFormsModule} from '@angular/forms';
import {NgbTooltipModule} from '@ng-bootstrap/ng-bootstrap';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {LicenseRoutingModule} from './license-routing.module';
import {UtmLicenseActivateComponent} from './shared/components/utm-license-activate/utm-license-activate.component';
import {UtmLicenseComponent} from './utm-license/utm-license.component';

@NgModule({
  declarations: [UtmLicenseComponent, UtmLicenseActivateComponent],
  imports: [
    CommonModule,
    LicenseRoutingModule,
    UtmSharedModule,
    ReactiveFormsModule,
    NgbTooltipModule
  ],
  entryComponents: [UtmLicenseActivateComponent]
})
export class LicenseModule {
}
