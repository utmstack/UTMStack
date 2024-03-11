import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {SectionConfigParamType} from "../../types/configuration/section-config-param.type";



@Component({
  selector: 'app-email-setting-notification',
  templateUrl: './email-setting-notificaction.component.html',
  styleUrls: ['./email-setting-notificaction.component.scss']
})
export class EmailSettingNotificactionComponent implements OnInit {
  @Output() action = new EventEmitter<string>();
  @Input ()  section: SectionConfigParamType;

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

    goTo() {
        this.action.emit('goTo');
    }
}
