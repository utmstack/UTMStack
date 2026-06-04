import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {UtmSharedModule} from '../shared/utm-shared.module';
import {ThreatWindRoutingModule} from './threatwind-routing.module';
import {ThreatwindComponent} from './threatwind.component';


@NgModule({
  declarations: [ ThreatwindComponent ],
  imports: [
    CommonModule,
    ThreatWindRoutingModule,
    UtmSharedModule,
    InfiniteScrollModule,
    NgbModule
  ]
})
export class ThreatWindModule {
}
