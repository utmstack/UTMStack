import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {ALERT_TAGS_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {getValueFromPropertyPath} from '../../../../shared/util/get-value-object-from-property-path.util';

@Component({
  selector: 'app-alert-tags-render',
  templateUrl: './alert-tags-render.component.html',
  styleUrls: ['./alert-tags-render.component.scss']
})
export class AlertTagsRenderComponent implements OnInit {
  @Input() icon: any;
  @Input() alert: any;
  @Input() tags: string[] = [];
  @Output() alertTagChange = new EventEmitter<string[]>();

  constructor() {
  }

  ngOnInit() {
    const tagAlert = getValueFromPropertyPath(alert, ALERT_TAGS_FIELD, null);
    this.tags = tagAlert ? tagAlert.split(',') : [];
  }

  onTagsChange($event: string[]) {
    this.alertTagChange.emit($event);
  }
}
