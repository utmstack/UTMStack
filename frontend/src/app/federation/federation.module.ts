import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {InlineSVGModule} from 'ng-inline-svg';
import {InstanceFormModalComponent} from './components/instance-form-modal/instance-form-modal.component';
import {InstanceSwitcherComponent} from './components/instance-switcher/instance-switcher.component';
import {FederationRoutingModule} from './federation-routing.module';
import {WelcomeComponent} from './pages/welcome/welcome.component';

@NgModule({
  imports: [
    CommonModule,
    FormsModule,
    NgbModule,
    InlineSVGModule,
    FederationRoutingModule
  ],
  declarations: [
    InstanceSwitcherComponent,
    InstanceFormModalComponent,
    WelcomeComponent
  ],
  exports: [
    InstanceSwitcherComponent
  ],
  entryComponents: [
    InstanceFormModalComponent
  ]
})
export class FederationModule {}
