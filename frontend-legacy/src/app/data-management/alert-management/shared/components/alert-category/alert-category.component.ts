import {Component, Input, OnInit} from '@angular/core';
import {ALERT_CATEGORY_DESCRIPTION_FIELD} from '../../../../../shared/constants/alert/alert-field.constant';
import {ADMIN_ROLE} from '../../../../../shared/constants/global.constant';
import {getValueFromPropertyPath} from '../../../../../shared/util/get-value-object-from-property-path.util';
import {replaceBreakLine} from '../../../../../shared/util/string-util';

@Component({
  selector: 'app-alert-category',
  templateUrl: './alert-category.component.html',
  styleUrls: ['./alert-category.component.scss']
})
export class AlertCategoryComponent implements OnInit {
  @Input() alert: object;
  alertCategoryDetail: string;
  ADMIN = ADMIN_ROLE;
  editMode = false;
  editing = false;

  constructor() {
  }

  ngOnInit() {
    this.alertCategoryDetail = replaceBreakLine(getValueFromPropertyPath(this.alert, ALERT_CATEGORY_DESCRIPTION_FIELD, null));
  }

  editCategory() {
    this.editMode = false;
  }
}
