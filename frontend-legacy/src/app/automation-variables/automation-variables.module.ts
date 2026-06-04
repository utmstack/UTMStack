import {NgModule} from '@angular/core';
import {CommonModule} from '@angular/common';
import {VariablesComponent} from "./variables/variables.component";
import {UtmSharedModule} from "../shared/utm-shared.module";
import {NgbModule} from "@ng-bootstrap/ng-bootstrap";
import {AutomationVariablesRoutingModule} from "./automation-variables-routing.module";

@NgModule({
  declarations: [VariablesComponent],
  imports: [
    CommonModule,
    UtmSharedModule,
    NgbModule,
    AutomationVariablesRoutingModule
  ]
})
export class AutomationVariablesModule {
}
