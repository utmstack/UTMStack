import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {ResizableModule} from 'angular-resizable-element';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {FileDetailComponent} from './file-detail/file-detail.component';
import {FileManagementRoutingModule} from './file-management-routing.module';
import {FileViewComponent} from './file-view/file-view.component';
import {FileManagementSharedModule} from './shared/file-management-shared.module';

@NgModule({
  declarations: [FileViewComponent, FileDetailComponent],
  imports: [
    CommonModule,
    FileManagementRoutingModule,
    FileManagementSharedModule,
    ResizableModule,
    UtmSharedModule,
    NgbModule
  ]
})
export class FileManagementModule {
}
