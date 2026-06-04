import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {InlineSVGModule} from 'ng-inline-svg';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {AppModuleActivateButtonComponent} from './components/app-module-activate-button/app-module-activate-button.component';
import {AppModuleActivateModalComponent} from './components/app-module-activate-modal/app-module-activate-modal.component';
import {AppModuleCardComponent} from './components/app-module-card/app-module-card.component';
import {AppModuleChecksComponent} from './components/app-module-checks/app-module-checks.component';
import {AppModuleDeactivateComponent} from './components/app-module-deactivate/app-module-deactivate.component';
import {IntCreateGroupComponent} from './components/int-create-group/int-create-group.component';
import {NgSelectModule} from "@ng-select/ng-select";

@NgModule({
  declarations: [AppModuleActivateModalComponent, AppModuleChecksComponent,
    AppModuleDeactivateComponent, AppModuleActivateButtonComponent, AppModuleCardComponent,
    IntCreateGroupComponent],
  imports: [
    CommonModule,
    UtmSharedModule,
    InlineSVGModule,
    NgbModule,
    ReactiveFormsModule,
    NgSelectModule,
    FormsModule
  ],
  exports: [AppModuleActivateButtonComponent, AppModuleCardComponent,
    IntCreateGroupComponent],
  entryComponents: [AppModuleActivateModalComponent, AppModuleChecksComponent, AppModuleDeactivateComponent, IntCreateGroupComponent]
})
export class AppModuleSharedModule {
}
