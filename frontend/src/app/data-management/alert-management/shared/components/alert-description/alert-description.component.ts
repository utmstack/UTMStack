import {Component, Input, OnInit} from '@angular/core';
import {ALERT_DESCRIPTION_FIELD} from '../../../../../shared/constants/alert/alert-field.constant';
import {ADMIN_ROLE} from '../../../../../shared/constants/global.constant';
import {getValueFromPropertyPath} from '../../../../../shared/util/get-value-object-from-property-path.util';
import {replaceBreakLine} from '../../../../../shared/util/string-util';

@Component({
  selector: 'app-alert-description',
  templateUrl: './alert-description.component.html',
  styleUrls: ['./alert-description.component.scss']
})
export class AlertDescriptionComponent implements OnInit {
  @Input() alert: object;
  alertDetail: string;
  ADMIN = ADMIN_ROLE;
  editMode = false;
  editing = false;

  constructor() {
  }

  ngOnInit() {
    this.alertDetail = replaceBreakLine(getValueFromPropertyPath(this.alert, ALERT_DESCRIPTION_FIELD, null));
  }

}
