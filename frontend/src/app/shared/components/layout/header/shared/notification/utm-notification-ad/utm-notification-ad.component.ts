import {Component, ElementRef, OnDestroy, OnInit, ViewChild} from '@angular/core';
import {UtmToastService} from '../../../../../../alert/utm-toast.service';
import {searchEventType} from '../../../../../../constants/active-directory-event.const';
import {ActiveAdModuleActiveService} from '../../../../../../services/active-modules/active-ad-module.service';

@Component({
  selector: 'app-utm-notification-ad',
  templateUrl: './utm-notification-ad.component.html',
  styleUrls: ['./utm-notification-ad.component.scss']
})
export class UtmNotificationAdComponent implements OnInit, OnDestroy {
  adEvents: {
    eventId: number,
    eventMessage?: string,
    objectName?: string,
    objectSid?: string,
    objectType?: string
    timestamp?: any
  }[] = [];
  @ViewChild('adChanges') adChanges: ElementRef;
  total = 0;
  private interval: number;
  newEvents = false;
  private prevTotal = -1;

  constructor(private toast: UtmToastService,
              private adModuleActiveService: ActiveAdModuleActiveService) {
  }

  ngOnDestroy() {
    clearInterval(this.interval);
  }

  ngOnInit() {
    this.getAdChanges();
    this.interval = setInterval(() => {
      this.getAdChanges();
    }, 30000);
  }

  setCount(count) {
    const el = this.adChanges.nativeElement;
    el.setAttribute('data-count', count > 20 ? '20+' : count);
  }

  resolveMessage(eventId: number): string {
    const event = searchEventType(eventId);
    if (event) {
      return event.message;
    } else {
      return 'Event info was not found';
    }
  }

  private getAdChanges() {
    this.adModuleActiveService.getAdChanges().subscribe(response => {
      this.adEvents = this.adEvents.concat(response.body);
      this.prevTotal = this.total;
      this.total = response.body.length;

      if (this.total > this.prevTotal) {
        this.toast.showInfoAssets('There are ' + this.total + ' new changes in active directory',
          'Active directory');
        this.newEvents = true;
      }
    });
  }


}
