import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-ad-notifications-config-delete',
  templateUrl: './ad-notifications-config-delete.component.html',
  styleUrls: ['./ad-notifications-config-delete.component.scss']
})
export class AdNotificationsConfigDeleteComponent implements OnInit {
  @Input() notification: any;

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

  deleteConfig() {
  }
}
