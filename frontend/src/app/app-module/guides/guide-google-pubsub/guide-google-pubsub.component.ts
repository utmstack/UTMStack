import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-google-pubsub',
  templateUrl: './guide-google-pubsub.component.html',
  styleUrls: ['./guide-google-pubsub.component.css']
})
export class GuideGooglePubsubComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;
  configValidity: boolean;

  constructor() {
  }

  ngOnInit() {
  }
  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }


}
