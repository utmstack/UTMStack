import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {InlineSVGModule} from 'ng-inline-svg';
import {InstanceFormModalComponent} from './components/instance-form-modal/instance-form-modal.component';
import {InstanceSwitcherComponent} from './components/instance-switcher/instance-switcher.component';
import {TeamManagementModalComponent} from './components/team-management-modal/team-management-modal.component';
import {TeamUserFormModalComponent} from './components/team-user-form-modal/team-user-form-modal.component';
import {WelcomeComponent} from './pages/welcome/welcome.component';

@NgModule({
  imports: [
    CommonModule,
    FormsModule,
    NgbModule,
    InlineSVGModule
  ],
  declarations: [
    InstanceSwitcherComponent,
    InstanceFormModalComponent,
    TeamManagementModalComponent,
    TeamUserFormModalComponent,
    WelcomeComponent
  ],
  exports: [
    InstanceSwitcherComponent,
    TeamManagementModalComponent
  ],
  entryComponents: [
    InstanceFormModalComponent,
    TeamManagementModalComponent,
    TeamUserFormModalComponent
  ]
})
export class FederationModule {}
