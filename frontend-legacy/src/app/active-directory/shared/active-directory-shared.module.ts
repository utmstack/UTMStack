import {CommonModule} from '@angular/common';
import {NgModule} from '@angular/core';
import {NgbModule} from '@ng-bootstrap/ng-bootstrap';
import {InfiniteScrollModule} from 'ngx-infinite-scroll';
import {UtmSharedModule} from '../../shared/utm-shared.module';
import {AdEventComponent} from './components/active-directory-event/active-directory-event.component';
import {AdMemberOfComponent} from './components/ad-member-of/ad-member-of.component';
import {EventTimelineComponent} from './components/event-timeline/event-timeline.component';

@NgModule({
  declarations: [
    AdEventComponent,
    AdMemberOfComponent,
    EventTimelineComponent,
  ],
  imports: [
    CommonModule,
    UtmSharedModule,
    InfiniteScrollModule,
    NgbModule
  ],
  exports: [
    AdMemberOfComponent,
    AdEventComponent,
    EventTimelineComponent,
  ]
})
export class ActiveDirectorySharedModule {
}
