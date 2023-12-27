import {CommonModule} from '@angular/common';
import {CUSTOM_ELEMENTS_SCHEMA, NgModule} from '@angular/core';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {RouterModule} from '@angular/router';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {AdminRoutingModule} from './admin-routing.module';
import {SourcesComponent} from './sources/sources.component';
import {UserMgmtDeleteDialogComponent} from './user/user-delete/user-management-delete-dialog.component';
import {UserMgmtDetailComponent} from './user/user-detail/user-management-detail.component';
import {UserMgmtComponent} from './user/user-list/user-management.component';
import {UserMgmtUpdateComponent} from './user/user-update/user-management-update.component';

@NgModule({
  declarations: [
    UserMgmtComponent,
    UserMgmtUpdateComponent,
    UserMgmtDetailComponent,
    UserMgmtDeleteDialogComponent,
    SourcesComponent
  ],
  imports: [
    CommonModule,
    RouterModule,
    AdminRoutingModule,
    FormsModule,
    NgbModule,
    ReactiveFormsModule,
    UtmSharedModule
  ],
  exports: [
    UserMgmtComponent,
    UserMgmtUpdateComponent,
    UserMgmtDetailComponent,
    UserMgmtDeleteDialogComponent,
  ],
  entryComponents: [
    UserMgmtDeleteDialogComponent
  ],
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  providers: []
})
export class AdminModule {
}
