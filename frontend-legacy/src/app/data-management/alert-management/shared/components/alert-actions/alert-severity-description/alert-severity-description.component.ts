import {Component, Input, OnInit} from '@angular/core';
import {ALERT_SEVERITY_DESCRIPTION_FIELD} from '../../../../../../shared/constants/alert/alert-field.constant';
import {ADMIN_ROLE} from '../../../../../../shared/constants/global.constant';
import {getValueFromPropertyPath} from '../../../../../../shared/util/get-value-object-from-property-path.util';
import {replaceBreakLine} from '../../../../../../shared/util/string-util';

@Component({
  selector: 'app-alert-severity-description',
  templateUrl: './alert-severity-description.component.html',
  styleUrls: ['./alert-severity-description.component.scss']
})
export class AlertSeverityDescriptionComponent implements OnInit {
  @Input() alert: object;
  alertSeverityDetail: string;
  ADMIN = ADMIN_ROLE;
  editMode = false;
  editing = false;

  constructor() {
  }

  ngOnInit() {
    this.alertSeverityDetail = replaceBreakLine(getValueFromPropertyPath(this.alert, ALERT_SEVERITY_DESCRIPTION_FIELD, null));
  }

  editSeverityDescription() {
    this.editMode = false;
  }

}
