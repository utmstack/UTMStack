import {Component, ElementRef, OnDestroy, OnInit, ViewChild} from '@angular/core';
import {Router} from '@angular/router';
import {UtmNetScanService} from '../../../../../../../assets-discover/shared/services/utm-net-scan.service';
import {UtmToastService} from '../../../../../../alert/utm-toast.service';

@Component({
  selector: 'app-utm-notification-assets',
  templateUrl: './utm-notification-assets.component.html',
  styleUrls: ['./utm-notification-assets.component.scss']
})
export class UtmNotificationAssetsComponent implements OnInit, OnDestroy {
  @ViewChild('notify') iconDisplay: ElementRef;
  totalAssets = 0;
  prevTotalAssets = 0;
  interval: any;

  constructor(public router: Router,
              private utmNetScanService: UtmNetScanService,
              private toast: UtmToastService
  ) {
    }


  ngOnInit() {
    this.getTotalAssets();
    this.interval = setInterval(() => {
      this.getTotalAssets();
    }, 15000);
  }

  ngOnDestroy() {
    clearInterval(this.interval);
  }

  setCount(count) {
    const el = this.iconDisplay.nativeElement;
    el.setAttribute('data-count', count > 20 ? '20+' : count);
  }

  private getTotalAssets() {
    this.utmNetScanService.countNewAssets().subscribe(response => {
      this.totalAssets = response.body;
      if (this.prevTotalAssets < this.totalAssets) {
        this.toast.showInfoAssets('There are ' + (this.totalAssets - this.prevTotalAssets) + ' new sources', 'New sources');
        this.prevTotalAssets = this.totalAssets;
        this.setCount(this.totalAssets);
      }
    });
  }

}


