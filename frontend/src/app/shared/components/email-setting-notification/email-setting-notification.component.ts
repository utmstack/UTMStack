import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {SectionConfigParamType} from '../../types/configuration/section-config-param.type';

@Component({
  selector: 'app-email-setting-notification',
  templateUrl: './email-setting-notification.component.html',
  styleUrls: ['./email-setting-notification.component.scss']
})
export class EmailSettingNotificationComponent implements OnInit {
  @Output() action = new EventEmitter<string>();
  @Input() section: SectionConfigParamType;

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

  goTo() {
    this.action.emit('goTo');
  }
}
