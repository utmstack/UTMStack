import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {ALERT_CATEGORY_FIELD, ALERT_NAME_FIELD, ALERT_SOLUTION_FIELD} from '../../../../../shared/constants/alert/alert-field.constant';
import {ADMIN_ROLE} from '../../../../../shared/constants/global.constant';
import {getValueFromPropertyPath} from '../../../../../shared/util/get-value-object-from-property-path.util';
import {replaceBreakLine} from '../../../../../shared/util/string-util';

@Component({
  selector: 'app-alert-proposed-solution',
  templateUrl: './alert-proposed-solution.component.html',
  styleUrls: ['./alert-proposed-solution.component.scss']
})
export class AlertProposedSolutionComponent implements OnInit {
  @Input() alert: object;
  alertSolution: string;
  ADMIN = ADMIN_ROLE;
  editMode = false;
  editing = false;
  editFlag: boolean;
  @Output() docSolutionCreated = new EventEmitter<'done'>();

  constructor() {
  }

  ngOnInit() {
    this.alertSolution = replaceBreakLine(getValueFromPropertyPath(this.alert, ALERT_SOLUTION_FIELD, null));
    this.editFlag = this.alertSolution !== null;
  }

  editSolution() {
    this.editing = true;
    // if (this.editFlag) {
    //   this.update();
    // } else {
    //   this.create();
    // }
  }

  createDocObject() {
    return {
      alertCategory: getValueFromPropertyPath(this.alert, ALERT_CATEGORY_FIELD, null),
      alertName: getValueFromPropertyPath(this.alert, ALERT_NAME_FIELD, null),
      solution: this.alertSolution
    };
  }

}
