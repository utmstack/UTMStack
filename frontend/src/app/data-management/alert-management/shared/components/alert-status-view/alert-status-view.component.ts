import {Component, Input, OnInit} from '@angular/core';
import {AUTOMATIC_REVIEW, CLOSED, IGNORED, OPEN, REVIEW} from '../../../../../shared/constants/alert/alert-status.constant';
import {resolveStatusStyle} from '../../util/alert-util-function';

@Component({
  selector: 'app-alert-status-view',
  templateUrl: './alert-status-view.component.html',
  styleUrls: ['./alert-status-view.component.scss']
})
export class AlertStatusViewComponent implements OnInit {
  @Input() status: number;
  icon: string;
  label: string;
  background: string;
  open = OPEN;
  review = REVIEW;
  ignored = IGNORED;
  closed = CLOSED;
  pending = AUTOMATIC_REVIEW;
  isIncident: boolean;

  constructor() {
  }

  ngOnInit() {
    this.resolveStatusAlertStyle();
  }

  private resolveStatusAlertStyle() {
    const style = resolveStatusStyle(this.status);
    this.icon = style.icon;
    this.background = style.background;
    this.label = style.label;
  }
}
