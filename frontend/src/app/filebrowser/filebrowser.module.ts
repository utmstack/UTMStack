import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {FileBrowserRoutingModule} from './filebrowser-routing.module';
import {RuleFieldBrowserComponent} from './rule-field-browser/rule-field-browser.component';
import {RuleFieldCardComponent} from './rule-field-browser/rule-field-card/rule-field-card.component';
import {RuleFileManagementComponent} from './rule-file-management/rule-file-management.component';

@NgModule({
  declarations: [RuleFileManagementComponent, RuleFieldBrowserComponent, RuleFieldCardComponent],
  imports: [
    CommonModule,
    FileBrowserRoutingModule,
    UtmSharedModule,
    InfiniteScrollModule,
    NgbModule
  ]
})
export class FileBrowserModule {
}
